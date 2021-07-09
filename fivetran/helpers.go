package fivetran

import (
	"fmt"
	"log"

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

// debug is a temp function. It should be improved to accept a variadic parameter
// and its name should change to logDebug
func debug(v interface{}) {
	log.Println(fmt.Sprintf("[DEBUG] FIVETRAN: %s", v))
}
