package datadog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"elastiq/client"
	"elastiq/config"
	"elastiq/jvalue"
	"elastiq/output"
	"elastiq/query"
)

type ddclient struct {
	config *config.Config
}

type response struct {
	Data []struct {
		Attributes struct {
			Attributes map[string]jvalue.JValue `json:"attributes"`
		} `json:"attributes"`
	} `json:"data"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

func (c *ddclient) Query(ctx context.Context, e *config.Env, q *query.Query, o query.Options) (io.Reader, error) {
	total := q.Limit

	output, err := c.config.GetOutput(e, q.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	if o.FromStdin {
		return applyOutputFromReader(os.Stdin, output)
	}

	ep := fmt.Sprintf("%s/api/v2/logs/events/search", e.GetEndpoint())
	iteration := 0
	sf := query.StartFrom(nil)

	readers := []io.Reader{}

	for total > 0 && iteration < 100 {
		iteration++

		var req *http.Request
		var body []byte

		if sf != nil {
			ssf := *sf
			ep = fmt.Sprint(ssf[0])
			req, err = http.NewRequestWithContext(ctx, "GET", ep, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create http request: %w", err)
			}
		} else {
			ddq, err := composeRequest(q, sf)
			if err != nil {
				return nil, fmt.Errorf("failed to compose request: %w", err)
			}

			body, err = json.Marshal(ddq)
			if err != nil {
				return nil, fmt.Errorf("failed to marsahl request: %w", err)
			}

			req, err = http.NewRequestWithContext(ctx, "POST", ep, bytes.NewReader(body))
			if err != nil {
				return nil, fmt.Errorf("failed to create http request: %w", err)
			}
		}

		req.Header.Add("content-type", "application/json")
		req.Header.Add("DD-API-KEY", e.DDAPIKey)
		req.Header.Add("DD-APPLICATION-KEY", e.DDAppKey)

		if o.AsCurl {
			str := fmt.Sprintf("curl -d '%s'", string(body))
			for k, v := range req.Header {
				str += fmt.Sprintf(" -H '%s: %s'", k, v[0])
			}
			str += fmt.Sprintf(" '%s'\n", ep)
			return strings.NewReader(str), nil
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http request failed: %w", err)
		}

		if res.StatusCode != 200 {
			errBody, _ := ioutil.ReadAll(res.Body)
			return nil, fmt.Errorf("got unexpected http code=%d, body='%s'", res.StatusCode, string(errBody))
		}

		if o.Raw {
			return res.Body, nil
		}

		if o.Recursive != nil {
			output.Decode = config.FromStringList(*o.Recursive)
		}

		resp, err := parseResponse(res.Body)
		if err != nil {
			return nil, err
		}

		if len(resp.Data) == 0 {
			break
		}

		total -= len(resp.Data)
		if resp.Links.Next != "" {
			next := []interface{}{resp.Links.Next}
			sf = &next
		}

		reader, err := applyOutput(resp, output)
		if err != nil {
			return nil, err
		}

		readers = append(readers, reader)

		if total >= q.Limit {
			break
		}
	}

	return io.MultiReader(readers...), nil
}

func NewClient(cfg *config.Config) client.Client {
	return &ddclient{config: cfg}
}

func parseResponse(r io.Reader) (*response, error) {
	j, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read all: %w", err)
	}

	resp := response{}

	err = json.Unmarshal(j, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

func applyOutputFromReader(r io.Reader, o *config.Output) (io.Reader, error) {
	resp, err := parseResponse(r)
	if err != nil {
		return nil, err
	}

	return applyOutput(resp, o)
}

func applyOutput(resp *response, o *config.Output) (io.Reader, error) {
	records := make([]map[string]interface{}, 0, len(resp.Data))
	for _, v := range resp.Data {
		r := make(map[string]interface{}, len(v.Attributes.Attributes))
		for k, v := range v.Attributes.Attributes {
			r[k] = v.Unwrap()
		}
		records = append(records, output.ApplyOutputFilters(r, o))
	}

	if o.Format == "json" {
		return output.JSONOutput(records)
	}

	return nil, fmt.Errorf("format='%s' is not implemented", o.Format)
}
