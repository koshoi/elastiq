package elasticsearch

import (
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
	StartFrom StartFrom             `json:"search_after,omitempty"`
	Sort      []map[string]RawOrder `json:"sort"`
	Query     RawQuery              `json:"query"`
}

func ComposeRequest(q *Query, sf StartFrom) ([]byte, error) {
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
		if f.Operation == IN {
			shouldsCnt++
			shoulds = append(shoulds, rf...)
		} else if f.Operation == NEQ {
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
