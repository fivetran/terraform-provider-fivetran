package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Row filters are kept as decoded JSON (map[string]interface{}, []interface{}, scalars), the same shape the API uses.
// A nil filter means there's no filter (upstream) or the filter has to be deleted (local).

// AND is the default filter operator, the API omits it in responses
const defaultRowFilterOperator = "AND"

func decodeRowFilter(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	var result interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	return result
}

func encodeRowFilter(value interface{}) json.RawMessage {
	result, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage("null")
	}
	return result
}

// projectRowFilter returns the upstream filter in the shape of the local one, so they can be compared and upstream can be saved into state.
// Upstream-only fields are dropped (e.g. computed `column_type`), except a non-default filter `operator`, which changes the filter meaning.
// Local fields missing upstream are kept only when they hold the default value.
func projectRowFilter(upstream, local interface{}) interface{} {
	switch l := local.(type) {
	case map[string]interface{}:
		u, ok := upstream.(map[string]interface{})
		if !ok {
			return upstream
		}
		result := map[string]interface{}{}
		for k, lv := range l {
			if uv, ok := u[k]; ok && uv != nil {
				result[k] = projectRowFilter(uv, lv)
			} else if isDefaultRowFilterValue(k, lv) {
				result[k] = lv
			}
		}
		if _, ok := l["operator"]; !ok {
			if op, ok := u["operator"]; ok && op != nil && fmt.Sprint(op) != defaultRowFilterOperator {
				result["operator"] = op
			}
		}
		return result
	case []interface{}:
		u, ok := upstream.([]interface{})
		if !ok || len(u) != len(l) {
			return upstream
		}
		result := make([]interface{}, len(l))
		for i := range l {
			result[i] = projectRowFilter(u[i], l[i])
		}
		return result
	default:
		// values are strings upstream, they might be numbers or booleans in the configuration
		if upstream != nil && local != nil && fmt.Sprint(upstream) == fmt.Sprint(local) {
			return local
		}
		return upstream
	}
}

func isDefaultRowFilterValue(key string, value interface{}) bool {
	if value == nil {
		return true
	}
	if key == "operator" {
		return fmt.Sprint(value) == defaultRowFilterOperator
	}
	if list, ok := value.([]interface{}); ok {
		return len(list) == 0
	}
	return false
}

func rowFilterEqual(upstream, local interface{}) bool {
	if local == nil || upstream == nil {
		return local == nil && upstream == nil
	}
	return reflect.DeepEqual(projectRowFilter(upstream, local), local)
}

// ValidateRowFilter checks the structure of a configured row filter, the API validates the clauses themselves
func ValidateRowFilter(value interface{}) error {
	if value == nil {
		return nil
	}
	filter, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("`row_filter` must be an object or null.")
	}
	for _, key := range []string{"name", "description"} {
		if v, ok := filter[key].(string); !ok || v == "" {
			return fmt.Errorf("`row_filter.%v` is required.", key)
		}
	}
	if op, ok := filter["operator"]; ok && op != nil {
		if s, ok := op.(string); !ok || (s != "AND" && s != "OR") {
			return fmt.Errorf("`row_filter.operator` must be `AND` or `OR`.")
		}
	}
	clauses, ok := filter["column_clauses"].([]interface{})
	if !ok || len(clauses) == 0 {
		return fmt.Errorf("`row_filter.column_clauses` must contain at least one clause.")
	}
	for i, c := range clauses {
		clause, ok := c.(map[string]interface{})
		if !ok {
			return fmt.Errorf("`row_filter.column_clauses[%v]` must be an object.", i)
		}
		if v, ok := clause["column"].(string); !ok || v == "" {
			return fmt.Errorf("`row_filter.column_clauses[%v].column` is required.", i)
		}
		if v, ok := clause["operator"].(string); !ok || v == "" {
			return fmt.Errorf("`row_filter.column_clauses[%v].operator` is required.", i)
		}
		if values, ok := clause["values"]; ok && values != nil {
			if _, ok := values.([]interface{}); !ok {
				return fmt.Errorf("`row_filter.column_clauses[%v].values` must be a list.", i)
			}
		}
	}
	return nil
}
