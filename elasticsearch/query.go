package elasticsearch

import (
	"fmt"
	"strings"
	"text/scanner"
	"unicode"
)

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

func ParseFilter(filter string) (*Filter, error) {
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

	tokens := []string{}
	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tokens = append(tokens, s.TokenText())
	}

	if len(tokens) < 3 {
		return nil, fmt.Errorf("insufficient tokens to compose filter, at least 3 required")
	}

	f.Key = unquoteValue(tokens[0])
	op := tokens[1]
	value := tokens[2]
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
				f.Operation = EQ
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
		if len(tokens) < 3 {
			return nil, fmt.Errorf("missing value")
		} else if len(tokens) > 3 {
			return nil, fmt.Errorf("too many values")
		}
		f.Operation = BT
		f.Value = []string{unquoteValue(value)}

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

	return &f, nil
}
