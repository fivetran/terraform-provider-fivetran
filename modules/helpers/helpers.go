package helpers

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SetContextTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		return ctx, cancel
	}
	return ctx, func() {}
}

func ValidateStringBooleanValue(val any, key string) (warns []string, errs []error) {
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

func FilterList(list []interface{}, filter func(elem interface{}) bool) *interface{} {
	for _, v := range list {
		if filter(v) {
			return &v
		}
	}
	return nil
}

func TryReadValue(source map[string]interface{}, key string) interface{} {
	if v, ok := source[key]; ok {
		return v
	}
	return nil
}

func TryReadListValue(source map[string]interface{}, key string) []interface{} {
	if v, ok := source[key]; ok {
		return v.([]interface{})
	}
	return nil
}

func CopySensitiveStringValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, localKey, upstreamKey string) {
	if upstreamKey == "" {
		upstreamKey = localKey
	}
	if localConfig == nil {
		// when using upstream value - use upstream key for source
		CopyStringValue(targetConfig, upstreamConfig, localKey, upstreamKey)
	} else {
		// when copying local value - use locak key for source
		CopyStringValue(targetConfig, *localConfig, localKey, "")
	}
}

func CopySensitiveListValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, targetKey, sourceKey string) {
	if localConfig != nil {
		if sourceKey == "" {
			sourceKey = targetKey
		}
		MapAddXInterface(targetConfig, targetKey, (*localConfig)[sourceKey].(*schema.Set).List())
	} else {
		CopyList(targetConfig, upstreamConfig, targetKey, sourceKey)
	}
}

func CopyStringValue(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].(string); ok {
		MapAddStr(target, targetKey, v)
	}
}

func CopyBooleanValue(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].(bool); ok {
		MapAddStr(target, targetKey, BoolToStr(v))
	}
}

func CopyIntegerValue(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].(float64); ok {
		MapAddStr(target, targetKey, strconv.Itoa((int(v))))
	}
}

func CopyList(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].([]interface{}); ok {
		MapAddXInterface(target, targetKey, v)
	}
}

func CopyIntegersList(target, source map[string]interface{}, targetKey, sourceKey string) {
	if sourceKey == "" {
		sourceKey = targetKey
	}
	if v, ok := source[sourceKey].([]interface{}); ok {
		result := make([]interface{}, len(v))
		for i, iv := range v {
			result[i] = strconv.Itoa(int(iv.(float64)))
		}
		MapAddXInterface(target, targetKey, result)
	}
}

// strToBool receives a string and returns a boolean
func StrToBool(s string) bool {
	return strings.ToLower(s) == "true"
}

// boolToStr receives a boolean and returns a string
func BoolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// boolPointertoStr receives a bool pointer and returns a string.
// An empty string is returned if the pointer is nil.
func BoolPointerToStr(b *bool) string {
	if b == nil {
		return ""
	}
	return BoolToStr(*b)
}

// strToInt receives a string and returns an int. A zero is returned
// if an error is found while converting the string to int.
func StrToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// intToStr receives an int and returns a string.
func IntToStr(i int) string {
	return strconv.Itoa(i)
}

// intPointerToStr receives an int pointer and returns a string.
// An empty string is returned if the pointer is nil.
func IntPointerToStr(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}

// xInterfaceStrXStr receives a []interface{} of type string and returns a []string
func XInterfaceStrXStr(xi []interface{}) []string {
	xs := make([]string, len(xi))
	for i, v := range xi {
		xs[i] = v.(string)
	}
	return xs
}

// xInterfaceStrXStr receives a []interface{} of type string and returns a []string
func XInterfaceStrXIneger(xi []interface{}) []int {
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
func MapAddStr(msi map[string]interface{}, k, v string) {
	if v != "" {
		msi[k] = v
	}
}

// mapAddXInterface adds a non-empty []interface{} to a map[string]interface{}
func MapAddXInterface(msi map[string]interface{}, k string, v []interface{}) {
	if len(v) > 0 {
		msi[k] = v
	}
}

// mapAddXInterface adds a non-empty []interface{} to a map[string]interface{}
func MapAddXString(msi map[string]interface{}, k string, v []string) {
	if len(v) > 0 {
		msi[k] = v
	}
}

// newDiag receives a diag.Severity, a summary, a detail, and returns a diag.Diagnostic
func NewDiag(severity diag.Severity, summary, detail string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: severity,
		Summary:  summary,
		Detail:   detail,
	}
}

// newAppendDiag receives diag.Diagnostics, a diag.Severity, a summary, and a detail. It makes a new
// diag.Diagnostic, appends it to the diag.Diagnostics and returns the diag.Diagnostics.
func NewDiagAppend(diags diag.Diagnostics, severity diag.Severity, summary, detail string) diag.Diagnostics {
	diags = append(diags, NewDiag(severity, summary, detail))
	return diags
}

func CopyMap(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range source {
		result[k] = v
	}
	return result
}

func CopyMapDeep(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range source {
		if vmap, ok := v.(map[string]interface{}); ok {
			result[k] = CopyMapDeep(vmap)
		} else {
			result[k] = v
		}
	}
	return result
}

func FilterMap(
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

func ContextDelay(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	select {
	case <-ctx.Done():
		t.Stop()
		return fmt.Errorf("interrupted: context deadline exceeded")
	case <-t.C:
	}
	return nil
}

// intersection accepts two slices of same type as arguments and returns three slices:
// uniques for the first argument, intersection and uniques for second argument
// results are collections of distinct elements (sets)
func Intersection[T comparable](a, b []T) (uniqueA, intersection, uniqueB []T) {
	hashA := make(map[T]bool)
	hashB := make(map[T]bool)
	for _, ai := range a {
		hashA[ai] = true
	}
	for _, bi := range b {
		hashB[bi] = true
	}

	for ai := range hashA {
		if _, ok := hashB[ai]; ok {
			intersection = append(intersection, ai)
			delete(hashB, ai)
		} else {
			uniqueA = append(uniqueA, ai)
		}
	}
	for bi := range hashB {
		uniqueB = append(uniqueB, bi)
	}
	return uniqueA, intersection, uniqueB
}

func StringInt32Hash(s string) int {
	h := fnv.New32a()
	var hashKey = []byte(s)
	h.Write(hashKey)
	return int(h.Sum32())
}
