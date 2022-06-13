package query

import (
	"elastiq/timetools"
	"fmt"
	"strings"
	"text/scanner"
	"time"
	"unicode"
)

type FilterOperation string

const (
	EQ  FilterOperation = "eq"  // equals
	TEQ FilterOperation = "teq" // term equals
	NEQ FilterOperation = "neq" // not equals
	GT  FilterOperation = "gt"  // greater than
	GTE FilterOperation = "gte" // greater or equal than
	LT  FilterOperation = "lt"  // less than
	LTE FilterOperation = "lte" // less or equal than
	IN  FilterOperation = "in"  // in
	LK  FilterOperation = "lk"  // like
	BT  FilterOperation = "bt"  // between
	BTT FilterOperation = "btt" // between time
	EX  FilterOperation = "ex"  // exists
	NEX FilterOperation = "nex" // not exists
)

type Filter struct {
	Key       string
	Value     []string
	Operation FilterOperation
}

type TimeFilterSettings struct {
	TimeZone   *time.Location
	TimeFormat string
	Now        *time.Time
}

// golang's strconv.Unquote probably works better with double quotes
// but it does not work with single and backticks as I expected
func unquoteValue(str string) string {
	if str[0] == str[len(str)-1] {
		switch string(str[0]) {
		case "'", "\"", "`":
			return str[1 : len(str)-1]
		default:
			return str
		}
	} else {
		return str
	}
}

func ParseFilter(filter string, tfs TimeFilterSettings, aliases map[string]string) (*Filter, error) {
	f := Filter{}

	var s scanner.Scanner
	s.Init(strings.NewReader(filter))
	s.Filename = "filter"
	s.IsIdentRune = func(ch rune, i int) bool {
		switch ch {
		case '.', ',', ':', '/', '-', '_', '(', ')', '@':
			return true
		}

		return unicode.IsLetter(ch) || unicode.IsDigit(ch)
	}
	s.Error = func(s *scanner.Scanner, msg string) {}

	tokens := []string{}
	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tokens = append(tokens, s.TokenText())
	}

	if len(tokens) < 2 {
		return nil, fmt.Errorf("insufficient tokens to compose filter, at least 2 required")
	}

	f.Key = unquoteValue(tokens[0])
	op := tokens[1]

	value := ""
	if len(tokens) > 2 {
		value = tokens[2]
	}

	switch op {
	case ">", "<", "=":
		if value == "=" {
			if len(tokens) < 4 {
				return nil, fmt.Errorf("missing value")
			} else if len(tokens) > 4 {
				return nil, fmt.Errorf("too many values")
			}

			f.Value = []string{unquoteValue(tokens[3])}
			switch op {
			case ">":
				f.Operation = GTE
			case "<":
				f.Operation = LTE
			case "=":
				f.Operation = TEQ
			}
		} else {
			if len(tokens) < 3 {
				return nil, fmt.Errorf("missing value")
			} else if len(tokens) > 3 {
				return nil, fmt.Errorf("too many values")
			}

			f.Value = []string{unquoteValue(value)}
			switch op {
			case ">":
				f.Operation = GT
			case "<":
				f.Operation = LT
			case "=":
				f.Operation = EQ
			}
		}

	case "!":
		if value != "=" {
			return nil, fmt.Errorf("failed to parse 'not equals' filter: expected '=', got '%s'", value)
		}

		if len(tokens) != 4 {
			return nil, fmt.Errorf("missing value")
		}

		f.Operation = NEQ
		f.Value = []string{unquoteValue(tokens[3])}

	case "in", "IN":
		if len(tokens) == 2 {
			return nil, fmt.Errorf("missing values")
		}
		f.Value = []string{}
		for i := 2; i < len(tokens); i++ {
			f.Value = append(f.Value, unquoteValue(tokens[i]))
		}
		f.Operation = IN

	case "bt", "BT", "between", "BETWEEN", "<>":
		if len(tokens) < 4 {
			return nil, fmt.Errorf("missing value")
		} else if len(tokens) > 4 {
			return nil, fmt.Errorf("too many values")
		}
		f.Operation = BT
		f.Value = []string{unquoteValue(value), unquoteValue(tokens[3])}

	case "time", "intime", "TIME", "INTIME":
		if len(tokens) < 4 {
			return nil, fmt.Errorf("missing values")
		} else if len(tokens) > 4 {
			return nil, fmt.Errorf("too many values")
		}

		now := time.Now().In(tfs.TimeZone)
		if tfs.Now != nil {
			now = (*tfs.Now).In(tfs.TimeZone)
		}

		from, err := timetools.ParseDate(unquoteValue(tokens[2]), now)
		if err != nil {
			return nil, fmt.Errorf("failed to parse str='%s' as date: %w", tokens[2], err)
		}

		to, err := timetools.ParseDate(unquoteValue(tokens[3]), now)
		if err != nil {
			return nil, fmt.Errorf("failed to parse str='%s' as date: %w", tokens[3], err)
		}

		f.Operation = BTT
		f.Value = []string{
			unquoteValue(timetools.FormatDate(from.In(tfs.TimeZone), tfs.TimeFormat)),
			unquoteValue(timetools.FormatDate(to.In(tfs.TimeZone), tfs.TimeFormat)),
		}

	case "ex", "EX", "exists", "EXISTS", "^":
		if len(tokens) > 2 {
			return nil, fmt.Errorf("too many values")
		}

		f.Operation = EX
		f.Value = []string{}

	// case "like", "LIKE", "~":
	// 	if len(tokens) < 3 {
	// 		return nil, fmt.Errorf("missing value")
	// 	} else if len(tokens) > 3 {
	// 		return nil, fmt.Errorf("too many value")
	// 	}
	// 	f.Operation = LK
	// 	f.Value = []string{unquoteValue(value)}

	default:
		return nil, fmt.Errorf(
			"unknown operation='%s', allowed operations are [%s]",
			op,
			strings.Join([]string{
				"=", "==",
				">", ">=", "<", "<=",
				// "like", "LIKE", "~",
				"in", "IN",
				"bt", "BT", "between", "BETWEEN",
				// "time", "intime", "TIME", "INTIME"
			}, ", "),
		)
	}

	if alias, ok := aliases[f.Key]; ok {
		f.Key = alias
	}

	return &f, nil
}
