package elasticsearch_test

import (
	"testing"
	"time"

	el "github.com/koshoi/elastiq/elasticsearch"
	"github.com/stretchr/testify/require"
)

func TestParseFilter(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output el.Filter
		err    bool
	}{
		{
			name:  "filter",
			input: "qwe=asd",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "filter with whitespaces",
			input: "qwe =    asd",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "strict filter",
			input: "qwe==asd",
			output: el.Filter{
				Operation: el.TEQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "filter with single quoted strings",
			input: "qwe='das asd'",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with double quoted strings",
			input: `qwe="das asd"`,
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with backtick quoted strings",
			input: "qwe=`das asd`",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with key with delimiters",
			input: "qwe.asd.zxc=asd",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe.asd.zxc",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "greater than",
			input: "qwe > asd",
			output: el.Filter{
				Operation: el.GT,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "less than",
			input: "qwe < asd",
			output: el.Filter{
				Operation: el.LT,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "greater than or equals",
			input: "qwe >= asd",
			output: el.Filter{
				Operation: el.GTE,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "less than or equals",
			input: "qwe <= asd",
			output: el.Filter{
				Operation: el.LTE,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "in with single value",
			input: "qwe in asd",
			output: el.Filter{
				Operation: el.IN,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "in with multiple values",
			input: "qwe in asd zxc lkj",
			output: el.Filter{
				Operation: el.IN,
				Key:       "qwe",
				Value:     []string{"asd", "zxc", "lkj"},
			},
		},
		{
			name:  "in with multiple quoted values",
			input: "qwe in asd 'zxc lkj'",
			output: el.Filter{
				Operation: el.IN,
				Key:       "qwe",
				Value:     []string{"asd", "zxc lkj"},
			},
		},
		{
			name:  "between",
			input: "qwe between asd zxc",
			output: el.Filter{
				Operation: el.BT,
				Key:       "qwe",
				Value:     []string{"asd", "zxc"},
			},
		},
		{
			name:  "intime",
			input: "qwe intime -1d now",
			output: el.Filter{
				Operation: el.BT,
				Key:       "qwe",
				Value:     []string{"2021-07-13T15:38:34Z", "2021-07-14T15:38:34Z"},
			},
		},
		{
			name:  "no input",
			input: "",
			err:   true,
		},
		{
			name:  "insufficient input",
			input: "qwe",
			err:   true,
		},
		{
			name:  "insufficient input for equals",
			input: "qwe=",
			err:   true,
		},
		{
			name:  "insufficient input for equals",
			input: "qwe in ",
			err:   true,
		},
		{
			name:  "insufficient input for between",
			input: "qwe bt asd",
			err:   true,
		},
		{
			name:  "unknown operator",
			input: "qwe unknown qwe",
			err:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2021, time.July, 14, 15, 38, 34, 0, time.UTC)
			f, err := el.ParseFilter(tt.input, el.TimeFilterSettings{
				TimeZone:   time.UTC,
				TimeFormat: time.RFC3339,
				Now:        &now,
			})
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.output, *f)
			}
		})
	}
}

func TestComposeFilter(t *testing.T) {
	tests := []struct {
		name   string
		err    bool
		filter el.Filter
		output []interface{}
	}{
		{
			name: "basic equals",
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.EQ,
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
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.TEQ,
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
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.NEQ,
			},
			output: []interface{}{
				map[string]interface{}{
					"must_not": map[string]interface{}{
						"match_phrase": map[string]string{
							"qwe": "asd",
						},
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
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.GTE,
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
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.LT,
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
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.LTE,
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
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: el.IN,
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
			filter: el.Filter{
				Key:       "qwe",
				Value:     []string{"asd", "zxc", "lkj"},
				Operation: el.IN,
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
