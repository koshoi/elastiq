package elasticsearch

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/koshoi/elastiq/config"
)

const maxRecordsPerRequest = 10000

// for development
// const maxRecordsPerRequest = 2

type Client interface {
	Query(ctx context.Context, env *config.Env, q *Query, o Options) (io.Reader, error)
}

type client struct {
	config *config.Config
}

type response struct {
	Hits struct {
		Hits []struct {
			Source map[string]JValue `json:"_source"`
			Sort   StartFrom         `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *client) Query(ctx context.Context, e *config.Env, q *Query, o Options) (io.Reader, error) {
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

	ep := fmt.Sprintf("%s/%s/_search?pretty", e.GetEndpoint(), index)
	iteration := 0
	sf := StartFrom(nil)

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

func NewClient(cfg *config.Config) Client {
	return &client{config: cfg}
}
