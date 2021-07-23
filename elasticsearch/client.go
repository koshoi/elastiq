package elasticsearch

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/koshoi/elastiq/config"
)

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
		} `json:"hits"`
	} `json:"hits"`
}

func (c *client) Query(ctx context.Context, e *config.Env, q *Query, o Options) (io.Reader, error) {
	body, err := ComposeRequest(q)
	if err != nil {
		return nil, fmt.Errorf("failed to compose request: %w", err)
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

	if o.Recursive != nil {
		output.Decode = config.FromStringList(*o.Recursive)
	}

	return applyOutput(res.Body, output)
}

func NewClient(cfg *config.Config) Client {
	return &client{config: cfg}
}
