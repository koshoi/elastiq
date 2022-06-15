package datadog

import (
	"elastiq/query"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_composeRequest(t *testing.T) {
	tests := []struct {
		name   string
		err    bool
		filter query.Filter
		output DataDogFilter
	}{
		{
			name: "basic equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.EQ,
			},
			output: DataDogFilter{
				Query: "qwe:*asd*",
			},
		},
		{
			name: "basic strict equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.TEQ,
			},
			output: DataDogFilter{
				Query: "qwe:asd",
			},
		},
		{
			name: "basic not equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.NEQ,
			},
			output: DataDogFilter{
				Query: "-qwe:asd",
			},
		},
		{
			name: "msg equals",
			filter: query.Filter{
				Key:       "msg",
				Value:     []string{"asd qwe"},
				Operation: query.EQ,
			},
			output: DataDogFilter{
				Query: "*asd?qwe*",
			},
		},
		{
			name: "timestamp",
			filter: query.Filter{
				Key:       "",
				Value:     []string{"10", "20"},
				Operation: query.BTT,
			},
			output: DataDogFilter{
				Query: "",
				From:  10000,
				To:    20000,
			},
		},
		{
			name: "basic greater",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.GT,
			},
			err: true,
		},
		{
			name: "basic greater or equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.GTE,
			},
			err: true,
		},
		{
			name: "basic less than",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.LT,
			},
			err: true,
		},
		{
			name: "basic less than or equals",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.LTE,
			},
			err: true,
		},
		{
			name: "basic in",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd"},
				Operation: query.IN,
			},
			err: true,
		},
		{
			name: "basic in with multiple values",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{"asd", "zxc", "lkj"},
				Operation: query.IN,
			},
			err: true,
		},
		{
			name: "basic exists",
			filter: query.Filter{
				Key:       "qwe",
				Value:     []string{},
				Operation: query.EX,
			},
			err: true,
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
			q := query.Query{}
			q.Filters = []*query.Filter{&tt.filter}

			res, err := composeRequest(&q, nil)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.output, res.Filter)
			}
		})
	}
}
