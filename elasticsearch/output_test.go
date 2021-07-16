package elasticsearch_test

import (
	"testing"

	el "github.com/koshoi/elastiq/elasticsearch"
	"github.com/stretchr/testify/require"
)

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  interface{}
		changed bool
	}{
		{
			name:    "basic string",
			input:   "Just a string",
			output:  "Just a string",
			changed: false,
		},
		{
			name:  "JSON string",
			input: `{"a":"b","c":[1,2,3]}`,
			output: map[string]interface{}{
				"a": "b",
				"c": []interface{}{1.0, 2.0, 3.0},
			},
			changed: true,
		},
		{
			name:  "YAML string",
			input: "---\n  a: b\n  c:\n    - 1\n    - 2\n    - 3\n",
			output: map[string]interface{}{
				"a": "b",
				"c": []interface{}{1, 2, 3},
			},
			changed: true,
		},
		{
			name:  "POST HTTP request",
			input: http1,
			output: map[string]interface{}{
				"method":  "POST",
				"url":     "/api/v1/method",
				"version": "HTTP/1.1",
				"headers": map[string]interface{}{
					"Host":           "somehost",
					"Content-Length": "17",
				},
				"body": `{"random":"body"}`,
			},
			changed: true,
		},
		{
			name:  "HTTP request with empty body",
			input: http2,
			output: map[string]interface{}{
				"method":  "POST",
				"url":     "/api/v1/method",
				"version": "HTTP/1.1",
				"headers": map[string]interface{}{
					"Host":           "somehost",
					"Content-Length": "0",
				},
				"body": ``,
			},
			changed: true,
		},
		{
			name:  "GET HTTP request",
			input: http3,
			output: map[string]interface{}{
				"method":  "GET",
				"url":     "/api/v1/method",
				"version": "HTTP/1.1",
				"headers": map[string]interface{}{
					"Host": "somehost",
				},
				"body": ``,
			},
			changed: true,
		},
		{
			name:  "HTTP request with ambigous header",
			input: http4,
			output: map[string]interface{}{
				"method":  "GET",
				"url":     "/api/v1/method",
				"version": "HTTP/1.1",
				"headers": map[string]interface{}{
					"Host":          "somehost",
					"SpecialHeader": "value: with: special: delimiter: ",
				},
				"body": ``,
			},
			changed: true,
		},
		{
			name:    "invalid HTTP request",
			input:   http5,
			output:  http5,
			changed: false,
		},
		{
			name:    "invalid HTTP method",
			input:   http6,
			output:  http6,
			changed: false,
		},
		{
			name:  "HTTP request with \\r\\n in them",
			input: http7,
			output: map[string]interface{}{
				"method":  "GET",
				"url":     "/api/v1/method",
				"version": "HTTP/1.1",
				"headers": map[string]interface{}{
					"Host": "somehost",
				},
				"body": ``,
			},
			changed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, changed := el.DecodeString(tt.input)
			require.Equal(t, tt.changed, changed)
			require.Equal(t, tt.output, output)
		})
	}
}

func TestRecursiveDecode(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		output interface{}
	}{
		{
			name:   "simple string",
			input:  "just a string",
			output: "just a string",
		},
		{
			name: "casual map",
			input: map[string]interface{}{
				"a": "b",
				"c": 1,
			},
			output: map[string]interface{}{
				"a": "b",
				"c": 1,
			},
		},
		{
			name: "map with JSON",
			input: map[string]interface{}{
				"a": `{"key":"value"}`,
				"c": 1,
			},
			output: map[string]interface{}{
				"a": map[string]interface{}{
					"key": "value",
				},
				"c": 1,
			},
		},
		{
			name: "map with JSON with JSON",
			input: map[string]interface{}{
				"a": `{"key":"{\"key2\":\"value2\"}"}`,
				"c": 1,
			},
			output: map[string]interface{}{
				"a": map[string]interface{}{
					"key": map[string]interface{}{
						"key2": "value2",
					},
				},
				"c": 1,
			},
		},
		{
			name: "map with JSON hidden inside",
			input: map[string]interface{}{
				"a": "b",
				"c": map[string]interface{}{
					"d": map[string]interface{}{
						"e": `{"key":"value"}`,
					},
					"f": 1,
				},
			},
			output: map[string]interface{}{
				"a": "b",
				"c": map[string]interface{}{
					"d": map[string]interface{}{
						"e": map[string]interface{}{
							"key": "value",
						},
					},
					"f": 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := el.RecursiveDecode(tt.input)
			require.Equal(t, tt.output, res)
		})
	}
}

var http1 = `POST /api/v1/method HTTP/1.1
Host: somehost
Content-Length: 17

{"random":"body"}`

var http2 = `POST /api/v1/method HTTP/1.1
Host: somehost
Content-Length: 0

`

var http3 = `GET /api/v1/method HTTP/1.1
Host: somehost

`

var http4 = `GET /api/v1/method HTTP/1.1
Host: somehost
SpecialHeader: value: with: special: delimiter: 

`

var http5 = `GET /api/v1/method HTTP/1.1
Host: somehost
InvalidHeaderLine

`

var http6 = `HELLO /api/v1/method HTTP/1.1
Host: somehost

`

var http7 = "GET /api/v1/method HTTP/1.1\r\nHost: somehost\r\n\r\n"
