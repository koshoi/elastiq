package elasticsearch

import (
	"encoding/json"
	"strconv"
	"strings"
)

func (jv *JValue) Unwrap() interface{} {
	v := jv.V
	switch vv := v.(type) {
	case []JValue:
		res := make([]interface{}, 0, len(vv))
		for i, v := range vv {
			res[i] = v.Unwrap()
		}
		return res

	case map[string]JValue:
		res := make(map[string]interface{}, len(vv))
		for k, v := range vv {
			res[k] = v.Unwrap()
		}
		return res

	case JValue:
		return vv.Unwrap()
	}

	return v
}

type JValue struct {
	V interface{}
}

func (jv *JValue) UnmarshalJSON(data []byte) error {
	numStr := string(data)
	if numStr[0] == '-' {
		numStr = string(numStr[1:])
	}

	if strings.IndexFunc(numStr, func(c rune) bool { return c < '0' || c > '9' }) == -1 {
		s := string(data)
		i, err := strconv.Atoi(s)
		if err == nil {
			jv.V = i
			return nil
		}

		i64, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			jv.V = i64
			return nil
		}

		u64, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			jv.V = u64
			return nil
		}
	}

	if data[0] == '[' {
		v := []JValue{}
		err := json.Unmarshal(data, &v)
		if err == nil {
			jv.V = v
			return nil
		}
	} else if data[0] == '{' {
		v := map[string]JValue{}
		err := json.Unmarshal(data, &v)
		if err == nil {
			jv.V = v
			return nil
		}
	}

	var v interface{}
	err := json.Unmarshal(data, &v)
	jv.V = v
	return err
}

func (jv JValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(jv.V)
}
