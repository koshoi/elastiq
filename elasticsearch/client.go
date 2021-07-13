package elasticsearch

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/koshoi/elastiq/config"
)

type Client interface {
	Query(ctx context.Context, env string, q *Query, o Options) (io.Reader, error)
}

type client struct {
	config *config.Config
}

type response struct {
	Hits struct {
		Hits []struct {
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func jsonOutput(records []map[string]interface{}) (io.Reader, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	for _, v := range records {
		err := enc.Encode(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal final result: %w", err)
		}
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func applyOutputFilters(record map[string]interface{}, o *config.Output) map[string]interface{} {
	if o.Only != nil {
		final := map[string]interface{}{}
		for _, k := range o.Only {
			final[k] = record[k]
		}

		return final
	}

	if o.Exclude != nil {
		for _, k := range o.Exclude {
			delete(record, k)
		}
	}

	return record
}

func applyOutput(r io.Reader, o *config.Output) (io.Reader, error) {
	j, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read all response: %w", err)
	}

	resp := response{}

	err = json.Unmarshal(j, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	records := make([]map[string]interface{}, 0, len(resp.Hits.Hits))
	for _, v := range resp.Hits.Hits {
		records = append(records, applyOutputFilters(v.Source, o))
	}

	if o.Format == "json" {
		return jsonOutput(records)
	}

	return nil, fmt.Errorf("format='%s' is not implemented", o.Format)
}

func (c *client) Query(ctx context.Context, env string, q *Query, o Options) (io.Reader, error) {
	body, err := ComposeRequest(q)
	if err != nil {
		return nil, fmt.Errorf("failed to compose request: %w", err)
	}

	e, err := c.config.GetEnv(env)
	if err != nil {
		return nil, err
	}

	index := q.Index
	if index == "" {
		index = e.Index
	}

	output, err := c.config.GetOutput(e, q.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to get output: %w", err)
	}

	if index == "" {
		return nil, fmt.Errorf("neither index was specified, nor default index for env was found")
	}

	ep := fmt.Sprintf("%s/%s/_search?pretty", e.GetEndpoint(), index)
	req, err := http.NewRequestWithContext(ctx, "POST", ep, bytes.NewReader(body))
	req.Header.Add("content-type", "application/json")
	if e.Authorization != nil {
		req.Header.Add(
			"Authorization",
			fmt.Sprintf(
				"Basic %s",
				base64.StdEncoding.EncodeToString(
					[]byte(
						e.Authorization.User+":"+e.Authorization.Password,
					),
				),
			),
		)
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

	return applyOutput(res.Body, output)
}

func NewClient(cfg *config.Config) Client {
	return &client{config: cfg}
}
