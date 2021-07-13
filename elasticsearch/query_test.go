package elasticsearch_test

import (
	"testing"

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
			name:  "filter aliased",
			input: "qwe==asd",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "filter with single quoted strings",
			input: "qwe=='das asd'",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with double quoted strings",
			input: `qwe=="das asd"`,
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with backtick quoted strings",
			input: "qwe==`das asd`",
			output: el.Filter{
				Operation: el.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with key with delimiters",
			input: "qwe.asd.zxc==asd",
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
			f, err := el.ParseFilter(tt.input)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.output, *f)
			}
		})
	}
}
