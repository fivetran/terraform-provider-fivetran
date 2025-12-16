package fivetrantypes

import (
	"encoding/json"
	"strings"
)

func jsonEqual(s1, s2 string) (bool, error) {
	s1, err := normalizeJSONString(s1)
	if err != nil {
		return false, err
	}

	s2, err = normalizeJSONString(s2)
	if err != nil {
		return false, err
	}

	return s1 == s2, nil
}

func normalizeJSONString(jsonStr string) (string, error) {
	dec := json.NewDecoder(strings.NewReader(jsonStr))

	dec.UseNumber()

	var temp interface{}
	if err := dec.Decode(&temp); err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(&temp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}