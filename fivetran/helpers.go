package fivetran

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func copySensitiveStringValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, localKey, upstreamKey string) {
	if upstreamKey == "" {
		upstreamKey = localKey
	}
	if localConfig == nil {
		// when using upstream value - use upstream key for source
		copyStringValue(targetConfig, upstreamConfig, localKey, upstreamKey)
	} else {
		// when copying local value - use locak key for source
		copyStringValue(targetConfig, *localConfig, localKey, "")
	}
}

func copySensitiveListValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, targetKey, sourceKey string) {
	if localConfig != nil {
		if sourceKey == "" {
			sourceKey = targetKey
		}
		mapAddXInterface(targetConfig, targetKey, (*localConfig)[sourceKey].(*schema.Set).List())
	} else {
		copyList(targetConfig, upstreamConfig, targetKey, sourceKey)
	}
}

func copyStringValue(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].(string); ok {
		mapAddStr(target, targetKey, v)
	}
}

func copyBooleanValue(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].(bool); ok {
		mapAddStr(target, targetKey, boolToStr(v))
	}
}

func copyIntegerValue(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].(float64); ok {
		mapAddStr(target, targetKey, strconv.Itoa((int(v))))
	}
}

func copyList(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].([]interface{}); ok {
		mapAddXInterface(target, targetKey, v)
	}
}

func copyIntegersList(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].([]interface{}); ok {
		result := make([]interface{}, len(v))
		for i, iv := range v {
			result[i] = strconv.Itoa(int(iv.(float64)))
		}
		mapAddXInterface(target, targetKey, result)
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
