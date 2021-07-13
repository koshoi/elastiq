package elasticsearch_test

import (
	"testing"

	el "github.com/koshoi/elastiq/elasticsearch"
	"github.com/stretchr/testify/require"
)

func TestComposeFilter(t *testing.T) {
	tests := []struct {
		name   string
		err    bool
		filter el.Filter
		output interface{}
	}{
		{
			name: "basic equals",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.EQ,
			},
			output: map[string]interface{}{
				"match_phrase": map[string]string{
					"qwe": "asd",
				},
			},
		},
		{
			name: "basic not equals",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.NEQ,
			},
			output: map[string]interface{}{
				"must_not": map[string]interface{}{
					"match_phrase": map[string]string{
						"qwe": "asd",
					},
				},
			},
		},
		{
			name: "basic greater",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.GT,
			},
			output: map[string]interface{}{
				"range": map[string]interface{}{
					"qwe": map[string]string{
						"gt": "asd",
					},
				},
			},
		},
		{
			name: "basic greater or equals",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.GTE,
			},
			output: map[string]interface{}{
				"range": map[string]interface{}{
					"qwe": map[string]string{
						"gte": "asd",
					},
				},
			},
		},
		{
			name: "basic less than",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.LT,
			},
			output: map[string]interface{}{
				"range": map[string]interface{}{
					"qwe": map[string]string{
						"lt": "asd",
					},
				},
			},
		},
		{
			name: "basic less than or equals",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.LTE,
			},
			output: map[string]interface{}{
				"range": map[string]interface{}{
					"qwe": map[string]string{
						"lte": "asd",
					},
				},
			},
		},
		{
			name: "basic in",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.IN,
			},
			output: map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"match_phrase": map[string]string{
							"qwe": "asd",
						},
					},
				},
			},
		},
		{
			name: "basic in with multiple values",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd", "zxc", "lkj"},
				Operation: el.IN,
			},
			output: map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"match_phrase": map[string]string{
							"qwe": "asd",
						},
					},
					{
						"match_phrase": map[string]string{
							"qwe": "zxc",
						},
					},
					{
						"match_phrase": map[string]string{
							"qwe": "lkj",
						},
					},
				},
			},
		},
		{
			name: "unknown operator",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.FilterOperation("unknown"),
			},
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := el.ComposeFilter(&tt.filter)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.output, res)
			}
		})
	}
}
