package elasticsearch

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

const maxRecordsPerRequest = 10000

// for development
// const maxRecordsPerRequest = 2

type elasticlient struct {
	config *config.Config
}

type response struct {
	Hits struct {
		Hits []struct {
			Source map[string]jvalue.JValue `json:"_source"`
			Sort   query.StartFrom          `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *elasticlient) Query(ctx context.Context, e *config.Env, q *query.Query, o query.Options) (io.Reader, error) {
	total := q.Limit

	// maximum records per request from elasticsearch
	if q.Limit > maxRecordsPerRequest {
		q.Limit = maxRecordsPerRequest
	}

	index := q.Index
	if index == "" {
		index = e.Index
	}

	output, err := c.config.GetOutput(e, q.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	if o.FromStdin {
		return applyOutputFromReader(os.Stdin, output)
	}

	if index == "" {
		return nil, fmt.Errorf("neither index was specified, nor default index for env was found")
	}

	ep := fmt.Sprintf("%s/%s/_search?pretty=true", e.GetEndpoint(), index)
	iteration := 0
	sf := query.StartFrom(nil)

	readers := []io.Reader{}

	for total > 0 && iteration < 100 {
		iteration++

		body, err := ComposeRequest(q, sf)
		if err != nil {
			return nil, fmt.Errorf("failed to compose request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", ep, bytes.NewReader(body))
		req.Header.Add("content-type", "application/json")
		if e.Authorization != nil {
			for k, v := range e.Authorization.Header {
				req.Header.Add(k, v.GetValue())
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create http request: %w", err)
		}

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
			return nil, fmt.Errorf("got unexpected http code=%d", res.StatusCode)
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

		if len(resp.Hits.Hits) == 0 {
			break
		}

		total -= len(resp.Hits.Hits)
		sf = resp.Hits.Hits[len(resp.Hits.Hits)-1].Sort

		reader, err := applyOutput(resp, output)
		if err != nil {
			return nil, err
		}

		readers = append(readers, reader)

		if len(resp.Hits.Hits) < q.Limit {
			break
		}
	}

	return io.MultiReader(readers...), nil
}

func NewClient(cfg *config.Config) client.Client {
	return &elasticlient{config: cfg}
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
	records := make([]map[string]interface{}, 0, len(resp.Hits.Hits))
	for _, v := range resp.Hits.Hits {
		r := make(map[string]interface{}, len(v.Source))
		for k, v := range v.Source {
			r[k] = v.Unwrap()
		}
		records = append(records, output.ApplyOutputFilters(r, o))
	}

	if o.Format == "json" {
		return output.JSONOutput(records)
	}

	return nil, fmt.Errorf("format='%s' is not implemented", o.Format)
}
