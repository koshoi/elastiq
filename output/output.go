package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"elastiq/config"
)

func JSONOutput(records []map[string]interface{}) (io.Reader, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	for _, v := range records {
		err := enc.Encode(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal final result: %w", err)
		}
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func decodeHTTPRequest(str string) interface{} {
	result := map[string]interface{}{}

	lines := strings.Split(str, "\n")

	if len(lines) < 2 {
		return nil
	}

	if strings.HasSuffix(lines[0], "\r") {
		lines = strings.Split(str, "\r\n")
		if len(lines) < 2 {
			return nil
		}
	}

	words := strings.Split(lines[0], " ")
	if len(words) != 3 {
		return nil
	}

	switch words[0] {
	case "POST", "GET", "HEAD", "PUT", "DELETE":
	default:
		return nil
	}

	result["method"] = words[0]
	result["url"] = words[1]
	result["version"] = words[2]

	bodyPos := -1
	headers := map[string]interface{}{}
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			bodyPos = i + 1
			break
		}

		words := strings.Split(line, ": ")

		if len(words) < 2 {
			return nil
		}

		headers[words[0]] = strings.Join(words[1:], ": ")
	}

	result["headers"] = headers

	if bodyPos > 0 {
		result["body"] = strings.Join(lines[bodyPos:], "\n")
	}

	return result
}

func DecodeString(str string, dmap map[string]bool) (interface{}, bool) {
	raw := []byte(str)
	if dmap["json"] {
		{
			v := map[string]interface{}{}
			if err := json.Unmarshal(raw, &v); err == nil {
				return v, true
			}
		}

		{
			v := []interface{}{}
			if err := json.Unmarshal(raw, &v); err == nil {
				return v, true
			}
		}
	}

	{
		if dmap["http"] {
			v := decodeHTTPRequest(str)
			if v != nil {
				return v, true
			}
		}
	}

	return str, false
}

func RecursiveDecode(i interface{}, dmap map[string]bool) interface{} {
	switch vv := i.(type) {
	case map[string]interface{}:
		for k, v := range vv {
			vv[k] = RecursiveDecode(v, dmap)
		}

	case []interface{}:
		for k, v := range vv {
			vv[k] = RecursiveDecode(v, dmap)
		}

	case string:
		if v, changed := DecodeString(vv, dmap); changed {
			return RecursiveDecode(v, dmap)
		}

		return vv
	}

	return i
}

func ApplyOutputFilters(record map[string]interface{}, o *config.Output) map[string]interface{} {
	if o.Only != nil {
		final := map[string]interface{}{}
		for _, k := range o.Only {
			final[k] = record[k]
		}

		record = final
	} else if o.Exclude != nil {
		for _, k := range o.Exclude {
			delete(record, k)
		}
	}

	if len(o.Decode) > 0 {
		for k, v := range record {
			record[k] = RecursiveDecode(v, o.Decode)
		}
	}

	return record
}
