package elasticsearch

import (
	"encoding/json"
	"fmt"
)

type RawOrder struct {
	Order string `json:"order"`
}

type RawFilter struct {
	Filter []interface{} `json:"filter"`
}

type RawQuery struct {
	Bool RawFilter `json:"bool"`
}

type ElasticRequest struct {
	Limit int                   `json:"size"`
	Sort  []map[string]RawOrder `json:"sort"`
	Query RawQuery              `json:"query"`
}

func ComposeRequest(q *Query) ([]byte, error) {
	order := q.Order
	if order == nil {
		order = &Order{
			By:        "@timestamp",
			Ascending: false,
		}
	}

	limit := q.Limit
	if limit == 0 {
		limit = 10
	}

	o := RawOrder{
		Order: "desc",
	}

	if order.Ascending {
		o.Order = "asc"
	}

	elkr := ElasticRequest{
		Limit: limit,
		Sort: []map[string]RawOrder{
			{
				order.By: o,
			},
		},
	}

	rfs := []interface{}{}

	for _, f := range q.Filters {
		rf, err := ComposeFilter(f)
		if err != nil {
			return nil, fmt.Errorf("failed to compose filter for %v: %w", f, err)
		}
		rfs = append(rfs, rf)
	}

	rq := RawQuery{
		Bool: RawFilter{
			Filter: rfs,
		},
	}

	elkr.Query = rq

	j, err := json.Marshal(elkr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return j, nil
}

func rangeStatement(op, key, value string) map[string]interface{} {
	return map[string]interface{}{
		"range": map[string]interface{}{
			key: map[string]string{
				op: value,
			},
		},
	}
}

func ComposeFilter(f *Filter) (interface{}, error) {
	switch f.Operation {
	case EQ:
		return map[string]interface{}{
			"match_phrase": map[string]string{
				f.Key: f.Value[0],
			},
		}, nil

	case NEQ:
		return map[string]interface{}{
			"must_not": map[string]interface{}{
				"match_phrase": map[string]string{
					f.Key: f.Value[0],
				},
			},
		}, nil

	case GT:
		return rangeStatement("gt", f.Key, f.Value[0]), nil

	case GTE:
		return rangeStatement("gte", f.Key, f.Value[0]), nil

	case LT:
		return rangeStatement("lt", f.Key, f.Value[0]), nil

	case LTE:
		return rangeStatement("lte", f.Key, f.Value[0]), nil

	case BT:
		return map[string]interface{}{
			"range": map[string]interface{}{
				"range": map[string]interface{}{
					f.Key: map[string]string{
						"gte": f.Value[0],
						"lte": f.Value[1],
					},
				},
			},
		}, nil

	case IN:
		shoulds := []map[string]interface{}{}
		for _, v := range f.Value {
			shoulds = append(shoulds, map[string]interface{}{
				"match_phrase": map[string]string{
					f.Key: v,
				},
			})
		}

		return map[string]interface{}{
			"should": shoulds,
		}, nil

	case LK:
		return map[string]interface{}{
			"match_phrase": map[string]string{
				f.Key: f.Value[0],
			},
		}, nil
	}

	return nil, fmt.Errorf("unknown operation='%s'", f.Operation)
}
