package helpers

// NormalizeRowFilter treats the documented default operator AND as omitted.
// Keep all other fields intact so missing or changed clauses remain visible.
func NormalizeRowFilter(value interface{}) interface{} {
	filter, ok := value.(map[string]interface{})
	if !ok || filter["operator"] != "AND" {
		return value
	}
	clauses, ok := filter["column_clauses"].([]interface{})
	if !ok || len(clauses) < 2 || len(clauses) > 10 {
		return value
	}
	normalized := make(map[string]interface{}, len(filter))
	for key, field := range filter {
		if key != "operator" {
			normalized[key] = field
		}
	}
	return normalized
}
