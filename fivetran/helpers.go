package fivetran

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// strToBool receives a string and returns a boolean
func strToBool(s string) bool {
	if s == "true" || s == "TRUE" || s == "True" {
		return true
	}
	return false
}

// boolToStr receives a boolean and returns a string
func boolToStr(b bool) string {
	if b == true {
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
// This is currently not in use.
func intToStr(i int) string {
	if i == 0 {
		return ""
	}
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
func xStrXInterface(xs []string) []interface{} {
	xi := make([]interface{}, len(xs))
	for i, v := range xs {
		xi[i] = v
	}
	return xi
}

// xInterfaceStrXStr receives a []interface{} of type string and returns a []string
func xInterfaceStrXStr(xi []interface{}) []string {
	xs := make([]string, len(xi))
	for i, v := range xi {
		xs[i] = v.(string)
	}
	return xs
}

// mapAddStr adds a non-empty string to a map[string]interface{}
func mapAddStr(msi map[string]interface{}, k, v string) {
	if v != "" {
		msi[k] = v
	}
}

// mapAddInt adds a non-zero int to a map[string]interface{}
func mapAddInt(msi map[string]interface{}, k string, v int) {
	if v != 0 {
		msi[k] = v
	}
}

// mapAddIntPointer adds a non-nil *int to a map[string]interface{}.
// This is currently not in use.
func mapAddIntP(msi map[string]interface{}, k string, v *int) {
	if v != nil {
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

// debug is a temporary function. It should be improved to accept a variadic parameter
// and its name should change to logDebug
func debug(v interface{}) {
	log.Println(fmt.Sprintf("[DEBUG] FIVETRAN: %s", v))
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
	prefix_required_services["airtable"] = true
	prefix_required_services["dynamics_365_fo"] = true
	prefix_required_services["mongo"] = true
	prefix_required_services["mongo_sharded"] = true
	prefix_required_services["aurora"] = true
	prefix_required_services["mysql_azure"] = true
	prefix_required_services["maria_azure"] = true
	prefix_required_services["maria"] = true
	prefix_required_services["mysql"] = true
	prefix_required_services["google_cloud_mysql"] = true
	prefix_required_services["magento_mysql"] = true
	prefix_required_services["magento_mysql_rds"] = true
	prefix_required_services["maria_rds"] = true
	prefix_required_services["mysql_rds"] = true
	prefix_required_services["oracle"] = true
	prefix_required_services["oracle_rac"] = true
	prefix_required_services["oracle_rds"] = true
	prefix_required_services["oracle_ebs"] = true
	prefix_required_services["aurora_postgres"] = true
	prefix_required_services["azure_postgres"] = true
	prefix_required_services["postgres"] = true
	prefix_required_services["google_cloud_postgresql"] = true
	prefix_required_services["heroku_postgres"] = true
	prefix_required_services["postgres_rds"] = true
	prefix_required_services["azure_sql_db"] = true
	prefix_required_services["sql_server"] = true
	prefix_required_services["sql_server_rds"] = true

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
