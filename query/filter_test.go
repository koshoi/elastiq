package query_test

import (
	"testing"
	"time"

	q "elastiq/query"

	"github.com/stretchr/testify/require"
)

func TestParseFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		output  q.Filter
		err     bool
		aliases map[string]string
	}{
		{
			name:  "filter",
			input: "qwe=asd",
			output: q.Filter{
				Operation: q.EQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "filter with aliases",
			input: "app=myapplication",
			output: q.Filter{
				Operation: q.EQ,
				Key:       "kubernetes.labels.app",
				Value:     []string{"myapplication"},
			},
			aliases: map[string]string{
				"app": "kubernetes.labels.app",
				"env": "kubernetes.labels.environment",
			},
		},
		{
			name:  "filter with not required aliases",
			input: "qwe=asd",
			output: q.Filter{
				Operation: q.EQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
			aliases: map[string]string{
				"app": "kubernetes.labels.app",
				"env": "kubernetes.labels.environment",
			},
		},
		{
			name:  "filter with whitespaces",
			input: "qwe =    asd",
			output: q.Filter{
				Operation: q.EQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "strict filter",
			input: "qwe==asd",
			output: q.Filter{
				Operation: q.TEQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "not equals filter",
			input: "qwe!=asd",
			output: q.Filter{
				Operation: q.NEQ,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "not equals filter with quoted value",
			input: "qwe!='asd zxc'",
			output: q.Filter{
				Operation: q.NEQ,
				Key:       "qwe",
				Value:     []string{"asd zxc"},
			},
		},
		{
			name:  "filter with single quoted strings",
			input: "qwe='das asd'",
			output: q.Filter{
				Operation: q.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with double quoted strings",
			input: `qwe="das asd"`,
			output: q.Filter{
				Operation: q.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with backtick quoted strings",
			input: "qwe=`das asd`",
			output: q.Filter{
				Operation: q.EQ,
				Key:       "qwe",
				Value:     []string{"das asd"},
			},
		},
		{
			name:  "filter with key with delimiters",
			input: "qwe.asd.zxc=asd",
			output: q.Filter{
				Operation: q.EQ,
				Key:       "qwe.asd.zxc",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "greater than",
			input: "qwe > asd",
			output: q.Filter{
				Operation: q.GT,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "less than",
			input: "qwe < asd",
			output: q.Filter{
				Operation: q.LT,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "greater than or equals",
			input: "qwe >= asd",
			output: q.Filter{
				Operation: q.GTE,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "less than or equals",
			input: "qwe <= asd",
			output: q.Filter{
				Operation: q.LTE,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "in with single value",
			input: "qwe in asd",
			output: q.Filter{
				Operation: q.IN,
				Key:       "qwe",
				Value:     []string{"asd"},
			},
		},
		{
			name:  "in with multiple values",
			input: "qwe in asd zxc lkj",
			output: q.Filter{
				Operation: q.IN,
				Key:       "qwe",
				Value:     []string{"asd", "zxc", "lkj"},
			},
		},
		{
			name:  "in with multiple quoted values",
			input: "qwe in asd 'zxc lkj'",
			output: q.Filter{
				Operation: q.IN,
				Key:       "qwe",
				Value:     []string{"asd", "zxc lkj"},
			},
		},
		{
			name:  "between",
			input: "qwe between asd zxc",
			output: q.Filter{
				Operation: q.BT,
				Key:       "qwe",
				Value:     []string{"asd", "zxc"},
			},
		},
		{
			name:  "intime",
			input: "qwe intime -1d now",
			output: q.Filter{
				Operation: q.BTT,
				Key:       "qwe",
				Value:     []string{"2021-07-13T15:38:34Z", "2021-07-14T15:38:34Z"},
			},
		},
		{
			name:  "exists",
			input: "qwe ex",
			output: q.Filter{
				Operation: q.EX,
				Key:       "qwe",
				Value:     []string{},
			},
		},
		{
			name:  "exists short",
			input: "qwe^",
			output: q.Filter{
				Operation: q.EX,
				Key:       "qwe",
				Value:     []string{},
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
		{
			name:  "invalid operator for greater",
			input: "qwe >! qwe",
			err:   true,
		},
		{
			name:  "invalid operator for not equals",
			input: "qwe !> qwe",
			err:   true,
		},
		{
			name:  "invalid operator for strict equals",
			input: "qwe =! qwe",
			err:   true,
		},
		{
			name:  "invalid operator for less",
			input: "qwe <> qwe",
			err:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2021, time.July, 14, 15, 38, 34, 0, time.UTC)
			f, err := q.ParseFilter(tt.input, q.TimeFilterSettings{
				TimeZone:   time.UTC,
				TimeFormat: time.RFC3339,
				Now:        &now,
			}, tt.aliases)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.output, *f)
			}
		})
	}
}
