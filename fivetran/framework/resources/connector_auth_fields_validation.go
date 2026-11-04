package resources

import (
	"fmt"
	"sort"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// rejectAuthFieldsInConfig fails validation when `config` contains a non-sensitive field that the service
// expects in `auth` only. The API keeps such fields in the auth section and never returns them,
// so they can't be read back into `config` state.
// Sensitive fields are not rejected: their configured value is kept in state as is, so they don't break apply.
func rejectAuthFieldsInConfig(configAttr types.Object, service types.String, configRoot path.Path, diags *diag.Diagnostics) {
	if configAttr.IsNull() || configAttr.IsUnknown() {
		return
	}
	if service.IsNull() || service.IsUnknown() {
		return
	}
	serviceName := service.ValueString()
	serviceConfigFields, err := common.GetFieldsForService(serviceName)
	if err != nil {
		// unknown service is reported on apply
		return
	}
	serviceAuthFields := common.GetAuthFieldsForService(serviceName)
	allConfigFields := common.GetConfigFieldsMap()

	configAttrs := configAttr.Attributes()
	fieldNames := make([]string, 0, len(configAttrs))
	for fn := range configAttrs {
		fieldNames = append(fieldNames, fn)
	}
	sort.Strings(fieldNames)

	for _, fn := range fieldNames {
		if configAttrs[fn].IsNull() {
			continue
		}
		if _, isAuthField := serviceAuthFields[fn]; !isAuthField {
			continue
		}
		if _, isConfigField := serviceConfigFields[fn]; isConfigField {
			continue
		}
		if allConfigFields[fn].GetIsSensitive(serviceName) {
			continue
		}
		diags.AddAttributeError(
			configRoot.AtName(fn),
			fmt.Sprintf("Field `%v` Must Be Set In `auth`", fn),
			fmt.Sprintf("Service `%v` expects field `%v` in the `auth` block, not in `config`. Move it to the `auth` block.", serviceName, fn),
		)
	}
}
