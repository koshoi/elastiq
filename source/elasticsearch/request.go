package elasticsearch

import (
	"elastiq/query"
	"encoding/json"
	"fmt"
)

type RawOrder struct {
	Order string `json:"order"`
}

type RawFilter struct {
	Filter               []interface{} `json:"filter"`
	Should               interface{}   `json:"should"`
	MustNot              []interface{} `json:"must_not"`
	MinimumShouldMatches int           `json:"minimum_should_match"`
}

type RawQuery struct {
	Bool RawFilter `json:"bool"`
}

type ElasticRequest struct {
	Limit     int                   `json:"size"`
	StartFrom query.StartFrom       `json:"search_after,omitempty"`
	Sort      []map[string]RawOrder `json:"sort"`
	Query     RawQuery              `json:"query"`
}

func ComposeRequest(q *query.Query, sf query.StartFrom) ([]byte, error) {
	order := q.Order
	if order == nil {
		order = &query.Order{
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
		Limit:     limit,
		StartFrom: sf,
		Sort: []map[string]RawOrder{
			{
				order.By: o,
			},
		},
	}

	rfs := []interface{}{}
	mustnots := []interface{}{}
	shoulds := []interface{}{}
	shouldsCnt := 0

	for _, f := range q.Filters {
		rf, err := ComposeFilter(f)
		if err != nil {
			return nil, fmt.Errorf("failed to compose filter for %v: %w", f, err)
		}
		if f.Operation == query.IN {
			shouldsCnt++
			shoulds = append(shoulds, rf...)
		} else if f.Operation == query.NEQ {
			mustnots = append(mustnots, rf...)
		} else {
			rfs = append(rfs, rf...)
		}
	}

	rq := RawQuery{
		Bool: RawFilter{
			Filter:               rfs,
			Should:               shoulds,
			MustNot:              mustnots,
			MinimumShouldMatches: shouldsCnt,
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

func ComposeFilter(f *query.Filter) ([]interface{}, error) {
	var res interface{} = nil

	switch f.Operation {
	case query.EQ:
		res = map[string]interface{}{
			"match_phrase": map[string]string{
				f.Key: f.Value[0],
			},
		}

	case query.TEQ:
		res = map[string]interface{}{
			"term": map[string]interface{}{
				f.Key: map[string]string{
					"value": f.Value[0],
				},
			},
		}

	case query.NEQ:
		res = map[string]interface{}{
			"match_phrase": map[string]string{
				f.Key: f.Value[0],
			},
		}

	case query.GT:
		res = rangeStatement("gt", f.Key, f.Value[0])

	case query.GTE:
		res = rangeStatement("gte", f.Key, f.Value[0])

	case query.LT:
		res = rangeStatement("lt", f.Key, f.Value[0])

	case query.LTE:
		res = rangeStatement("lte", f.Key, f.Value[0])

	case query.BT:
		res = map[string]interface{}{
			"range": map[string]interface{}{
				f.Key: map[string]string{
					"gte": f.Value[0],
					"lte": f.Value[1],
				},
			},
		}

	case query.IN:
		shoulds := []interface{}{}
		for _, v := range f.Value {
			shoulds = append(shoulds, map[string]interface{}{
				"match_phrase": map[string]string{
					f.Key: v,
				},
			})
		}

		return shoulds, nil

	case query.LK:
		res = map[string]interface{}{
			"match_phrase": map[string]string{
				f.Key: f.Value[0],
			},
		}

	case query.EX:
		res = map[string]interface{}{
			"exists": map[string]string{
				"field": f.Key,
			},
		}
	}

	if res != nil {
		return []interface{}{res}, nil
	}

	return nil, fmt.Errorf("unknown operation='%s'", f.Operation)
}
