package elasticsearch_test

import (
	"testing"

	"elastiq/query"
	"elastiq/source/elasticsearch"

	"github.com/stretchr/testify/require"
)

func TestComposeFilter(t *testing.T) {
	tests := []struct {
		name   string
		err    bool
		filter query.Filter
		output []interface{}
	}{
		{
			name: "basic equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.EQ,
			},
			output: []interface{}{
				map[string]interface{}{
					"match_phrase": map[string]string{
						"qwe": "asd",
					},
				},
			},
		},
		{
			name: "basic strict equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.TEQ,
			},
			output: []interface{}{
				map[string]interface{}{
					"term": map[string]interface{}{
						"qwe": map[string]string{
							"value": "asd",
						},
					},
				},
			},
		},
		{
			name: "basic not equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.NEQ,
			},
			output: []interface{}{
				map[string]interface{}{
					"match_phrase": map[string]string{
						"qwe": "asd",
					},
				},
			},
		},
		{
			name: "basic greater",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.GT,
			},
			output: []interface{}{
				map[string]interface{}{
					"range": map[string]interface{}{
						"qwe": map[string]string{
							"gt": "asd",
						},
					},
				},
			},
		},
		{
			name: "basic greater or equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.GTE,
			},
			output: []interface{}{
				map[string]interface{}{
					"range": map[string]interface{}{
						"qwe": map[string]string{
							"gte": "asd",
						},
					},
				},
			},
		},
		{
			name: "basic less than",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.LT,
			},
			output: []interface{}{
				map[string]interface{}{
					"range": map[string]interface{}{
						"qwe": map[string]string{
							"lt": "asd",
						},
					},
				},
			},
		},
		{
			name: "basic less than or equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.LTE,
			},
			output: []interface{}{
				map[string]interface{}{
					"range": map[string]interface{}{
						"qwe": map[string]string{
							"lte": "asd",
						},
					},
				},
			},
		},
		{
			name: "basic in",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.IN,
			},
			output: []interface{}{
				map[string]interface{}{
					"match_phrase": map[string]string{
						"qwe": "asd",
					},
				},
			},
		},
		{
			name: "basic in with multiple values",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd", "zxc", "lkj"},
				Operation: query.IN,
			},
			output: []interface{}{
				map[string]interface{}{
					"match_phrase": map[string]string{
						"qwe": "asd",
					},
				},
				map[string]interface{}{
					"match_phrase": map[string]string{
						"qwe": "zxc",
					},
				},
				map[string]interface{}{
					"match_phrase": map[string]string{
						"qwe": "lkj",
					},
				},
			},
		},
		{
			name: "basic exists",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{},
				Operation: query.EX,
			},
			output: []interface{}{
				map[string]interface{}{
					"exists": map[string]string{
						"field": "qwe",
					},
				},
			},
		},
		{
			name: "unknown operator",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.FilterOperation("unknown"),
			},
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := elasticsearch.ComposeFilter(&tt.filter)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.output, res)
			}
		})
	}
}
