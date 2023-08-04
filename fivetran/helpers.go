package fivetran

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func validateStringBooleanValue(val any, key string) (warns []string, errs []error) {
	v := val.(string)
	if v == "" {
		return
	}

	if strings.ToLower(v) == "true" || strings.ToLower(v) == "false" {
		if strings.ToLower(v) != v {
			warns = append(warns, "For %q please use lower case boolean value `true` or `false`")
		}
		return
	}

	errs = append(errs, fmt.Errorf("%q must be a boolean value `true` or `false`; got: %s", key, v))
	return
}

func filterList(list []interface{}, filter func(elem interface{}) bool) *interface{} {
	for _, v := range list {
		if filter(v) {
			return &v
		}
	}
	return nil
}

func tryReadValue(source map[string]interface{}, key string) interface{} {
	if v, ok := source[key]; ok {
		return v
	}
	return nil
}

func tryReadListValue(source map[string]interface{}, key string) []interface{} {
	if v, ok := source[key]; ok {
		return v.([]interface{})
	}
	return nil
}

// tryCopyStringValue copies string value from map `source` to map `target` if `key` represented in `source` map
func tryCopyStringValue(target, source map[string]interface{}, key string) {
	if v, ok := source[key].(string); ok {
		mapAddStr(target, key, v)
	}
}

// tryReadBooleanValue copies bool value from map `source` to map `target` if `key` represented in `source` map
func tryCopyBooleanValue(target, source map[string]interface{}, key string) {
	if v, ok := source[key].(bool); ok {
		mapAddStr(target, key, boolToStr(v))
	}
}

// tryReadIntegerValue copies int value from map `source` to map `target` if `key` represented in `source` map
func tryCopyIntegerValue(target, source map[string]interface{}, key string) {
	if v, ok := source[key].(float64); ok {
		mapAddStr(target, key, strconv.Itoa((int(v))))
	}
}

// tryReadList copies abstract list ()`[]interface{}`) from map `source` to map `target` if `key` represented in `source` map
func tryCopyList(target, source map[string]interface{}, key string) {
	if v, ok := source[key].([]interface{}); ok {
		mapAddXInterface(target, key, v)
	}
}

// List of integers is represented on terraform side as list of strings for simplicity, but on upstream side it's strict list of integers
func tryCopyIntegersList(target, source map[string]interface{}, key string) {
	if v, ok := source[key].([]interface{}); ok {
		result := make([]interface{}, len(v))
		for i, iv := range v {
			result[i] = strconv.Itoa(int(iv.(float64)))
		}
		mapAddXInterface(target, key, result)
	}
}

// strToBool receives a string and returns a boolean
func strToBool(s string) bool {
	return strings.ToLower(s) == "true"
}

// boolToStr receives a boolean and returns a string
func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// boolPointertoStr receives a bool pointer and returns a string.
// An empty string is returned if the pointer is nil.
func boolPointerToStr(b *bool) string {
	if b == nil {
		return ""
	}
	return boolToStr(*b)
}

// strToInt receives a string and returns an int. A zero is returned
// if an error is found while converting the string to int.
func strToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// intToStr receives an int and returns a string.
func intToStr(i int) string {
	return strconv.Itoa(i)
}

// intPointerToStr receives an int pointer and returns a string.
// An empty string is returned if the pointer is nil.
func intPointerToStr(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}

// xStrXInterface receives a []string and returns a []interface{}
// func xStrXInterface(xs []string) []interface{} {
// 	xi := make([]interface{}, len(xs))
// 	for i, v := range xs {
// 		xi[i] = v
// 	}
// 	return xi
// }

// xInterfaceStrXStr receives a []interface{} of type string and returns a []string
func xInterfaceStrXStr(xi []interface{}) []string {
	xs := make([]string, len(xi))
	for i, v := range xi {
		xs[i] = v.(string)
	}
	return xs
}

// xInterfaceStrXStr receives a []interface{} of type string and returns a []string
func xInterfaceStrXIneger(xi []interface{}) []int {
	xs := make([]int, len(xi))
	for i, v := range xi {
		integerValue, e := strconv.Atoi(v.(string))
		if e != nil {
			panic(e)
		}
		xs[i] = integerValue
	}
	return xs
}

// mapAddStr adds a non-empty string to a map[string]interface{}
func mapAddStr(msi map[string]interface{}, k, v string) {
	if v != "" {
		msi[k] = v
	}
}

// mapAddXInterface adds a non-empty []interface{} to a map[string]interface{}
func mapAddXInterface(msi map[string]interface{}, k string, v []interface{}) {
	if len(v) > 0 {
		msi[k] = v
	}
}

// newDiag receives a diag.Severity, a summary, a detail, and returns a diag.Diagnostic
func newDiag(severity diag.Severity, summary, detail string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: severity,
		Summary:  summary,
		Detail:   detail,
	}
}

// newAppendDiag receives diag.Diagnostics, a diag.Severity, a summary, and a detail. It makes a new
// diag.Diagnostic, appends it to the diag.Diagnostics and returns the diag.Diagnostics.
func newDiagAppend(diags diag.Diagnostics, severity diag.Severity, summary, detail string) diag.Diagnostics {
	diags = append(diags, newDiag(severity, summary, detail))
	return diags
}

func copyMap(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range source {
		result[k] = v
	}
	return result
}

func copyMapDeep(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range source {
		if vmap, ok := v.(map[string]interface{}); ok {
			result[k] = copyMapDeep(vmap)
		} else {
			result[k] = v
		}
	}
	return result
}

func filterMap(
	source map[string]interface{},
	filter func(interface{}) bool,
	accept func(interface{}) interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range source {
		if filter(v) {
			if accept != nil {
				result[k] = accept(v)
			} else {
				result[k] = v
			}
		}
	}
	return result
}

func readDestinationSchema(schema string, service string) []interface{} {
	destination_schema := make([]interface{}, 1)

	prefix_required_services := make(map[string]bool)

	// this list reflects all db-like connectors we know that should use destination_schema.prefix field
	prefix_required_services["azure_sql_managed_db"] = true
	prefix_required_services["oracle_fusion_cloud_apps_hcm"] = true
	prefix_required_services["oracle_fusion_cloud_apps_fscm"] = true
	prefix_required_services["google_cloud_postgresql"] = true
	prefix_required_services["google_cloud_mysql"] = true
	prefix_required_services["sql_server_hva"] = true
	prefix_required_services["oracle_rac"] = true
	prefix_required_services["oracle_ebs"] = true
	prefix_required_services["google_cloud_sqlserver"] = true
	prefix_required_services["maria"] = true
	prefix_required_services["teradata"] = true
	prefix_required_services["mongo_sharded"] = true
	prefix_required_services["oracle_fusion_cloud_apps_crm"] = true
	prefix_required_services["dynamics_365_fo"] = true
	prefix_required_services["magento_mysql_rds"] = true
	prefix_required_services["documentdb"] = true
	prefix_required_services["hana_sap_hva_s4"] = true
	prefix_required_services["sql_server_rds"] = true
	prefix_required_services["maria_azure"] = true
	prefix_required_services["db2i_hva"] = true
	prefix_required_services["azure_postgres"] = true
	prefix_required_services["postgres"] = true
	prefix_required_services["aurora_postgres"] = true
	prefix_required_services["mysql"] = true
	prefix_required_services["oracle"] = true
	prefix_required_services["mysql_azure"] = true
	prefix_required_services["mongo"] = true
	prefix_required_services["hana_sap_hva_ecc"] = true
	prefix_required_services["maria_rds"] = true
	prefix_required_services["airtable"] = true
	prefix_required_services["aurora"] = true
	prefix_required_services["db2i_sap_hva"] = true
	prefix_required_services["oracle_hva"] = true
	prefix_required_services["postgres_rds"] = true
	prefix_required_services["cosmos"] = true
	prefix_required_services["oracle_sap_hva"] = true
	prefix_required_services["db2"] = true
	prefix_required_services["sql_server"] = true
	prefix_required_services["azure_sql_db"] = true
	prefix_required_services["sap_hana_db"] = true
	prefix_required_services["mysql_rds"] = true
	prefix_required_services["snowflake_db"] = true
	prefix_required_services["magento_mysql"] = true
	prefix_required_services["heroku_postgres"] = true
	prefix_required_services["sql_server_sap_ecc_hva"] = true
	prefix_required_services["oracle_rds"] = true
	prefix_required_services["bigquery_db"] = true

	ds := make(map[string]interface{})

	if prefix_required_services[service] {
		mapAddStr(ds, "prefix", schema)
	} else {
		s := strings.Split(schema, ".")
		mapAddStr(ds, "name", s[0])
		if len(s) > 1 {
			mapAddStr(ds, "table", s[1])
		}
	}

	destination_schema[0] = ds
	return destination_schema
}
