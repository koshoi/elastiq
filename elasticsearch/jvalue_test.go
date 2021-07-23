package elasticsearch_test

import (
	"encoding/json"
	"testing"

	el "github.com/koshoi/elastiq/elasticsearch"
	"github.com/stretchr/testify/require"
)

func j(v interface{}) el.JValue {
	return el.JValue{V: v}
}

func TestJValue_Unmarshal(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output interface{}
	}{
		{
			name:  "simple JSON",
			input: `{"a":"b"}`,
			output: map[string]el.JValue{
				"a": j("b"),
			},
		},
		{
			name:  "deep JSON",
			input: `{"a":{"b":{"c":"d"}}}`,
			output: map[string]el.JValue{
				"a": j(map[string]el.JValue{
					"b": j(map[string]el.JValue{
						"c": j("d"),
					}),
				}),
			},
		},
		{
			name:  "JSON with float",
			input: `{"a":1.1}`,
			output: map[string]el.JValue{
				"a": j(1.1),
			},
		},
		{
			name:  "JSON with int that looses precission in float",
			input: `{"a":603427666509977819}`,
			output: map[string]el.JValue{
				"a": j(603427666509977819),
			},
		},
		{
			name:  "JSON with negative int that looses precission in float",
			input: `{"a":-603427666509977819}`,
			output: map[string]el.JValue{
				"a": j(-603427666509977819),
			},
		},
		{
			name:  "JSON with negative int64",
			input: `{"a":-9223372036854775807}`,
			output: map[string]el.JValue{
				"a": j(-9223372036854775807),
			},
		},
		{
			name:  "JSON with uint64",
			input: `{"a":18446744073709551615}`,
			output: map[string]el.JValue{
				"a": j(uint64(18446744073709551615)),
			},
		},
		{
			name:  "deep JSON with uint64",
			input: `{"a":{"b":18446744073709551615}}`,
			output: map[string]el.JValue{
				"a": j(map[string]el.JValue{
					"b": j(uint64(18446744073709551615)),
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v map[string]el.JValue
			err := json.Unmarshal([]byte(tt.input), &v)
			require.NoError(t, err)
			require.Equal(t, tt.output, v)
		})
	}
}
