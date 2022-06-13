package datadog

import (
	"elastiq/query"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type DataDogRequest struct {
	Filter struct {
		Query string `json:"query"`
		From  int64  `json:"from"`
		To    int64  `json:"to"`
	} `json:"filter"`
}

var ddSpecialChars = []string{
	"\\",
	"+", "-", "=", "*",
	"&&", "||",
	">", "<",
	"!", "?",
	"(", ")",
	"{", "}",
	"[", "]",
	"^",
	"\"",
	"“", "”",
	"~",
	":",
	"/",
}

func ddEscapeFilter(key, value string) string {
	if key == "msg" {
		for _, sc := range ddSpecialChars {
			value = strings.ReplaceAll(value, sc, "?")
		}

		return value
	}

	for _, sc := range ddSpecialChars {
		value = strings.ReplaceAll(value, sc, "\\"+sc)
	}

	return value
}

func composeRequest(q *query.Query, sf query.StartFrom) ([]byte, error) {
	ddq := DataDogRequest{}

	queryFilters := []string{}
	for _, f := range q.Filters {
		keyPart := f.Key + ":"
		if f.Key == "msg" {
			keyPart = ""
		}

		switch f.Operation {
		case query.EQ:
			queryFilters = append(queryFilters, fmt.Sprintf("%s*%s*", keyPart, ddEscapeFilter(f.Key, f.Value[0])))

		case query.TEQ:
			queryFilters = append(queryFilters, fmt.Sprintf("%s%s", keyPart, ddEscapeFilter(f.Key, f.Value[0])))

		case query.NEQ:
			queryFilters = append(queryFilters, fmt.Sprintf("-%s%s", keyPart, ddEscapeFilter(f.Key, f.Value[0])))

		case query.BT:
			queryFilters = append(queryFilters, fmt.Sprintf("%s[%s TO %s]", keyPart, f.Value[0], f.Value[1]))

		case query.BTT:
			from, err := strconv.ParseInt(f.Value[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse unix timestamp='%s': %w", f.Value[0], err)
			}

			to, err := strconv.ParseInt(f.Value[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse unix timestamp='%s': %w", f.Value[1], err)
			}

			ddq.Filter.From = from * 1000
			ddq.Filter.To = to * 1000
		}
	}

	ddq.Filter.Query = strings.Join(queryFilters, " ")
	qq, err := json.Marshal(ddq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dd json query: %w", err)
	}

	return qq, nil
}
