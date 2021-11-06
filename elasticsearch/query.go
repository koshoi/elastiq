package elasticsearch

import (
	"fmt"
	"strings"
)

type Order struct {
	By        string
	Ascending bool
}

type StartFrom *[]interface{}

func GetOrder(order string) (*Order, error) {
	parts := strings.Split(order, "/")
	if len(parts) > 2 {
		return nil, fmt.Errorf("order='%s' splited in too many parts", order)
	}

	if len(parts) == 1 {
		parts = append(parts, "desc")
	}

	key := parts[0]
	o := parts[1]

	if key == "" {
		return nil, fmt.Errorf("got order='%s', order key can not be empty", order)
	}

	res := Order{
		By: key,
	}

	switch o {
	case "desc", "DESC", "descenging", "DESCENDING", "d", "D":
		res.Ascending = false
	case "asc", "ASC", "ascending", "ASCENDING", "a", "A":
		res.Ascending = true
	default:
		return nil, fmt.Errorf("got order='%s', unknown order direction='%s'", order, o)
	}

	return &res, nil
}

type Query struct {
	Filters []*Filter
	Order   *Order
	Limit   int
	Index   string
	Output  string
}

type Options struct {
	Debug     bool
	Recursive *[]string
	Raw       bool
	AsCurl    bool
	FromStdin bool
}
