package resources

import (
	"fmt"
	"sort"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// warnAuthFieldsInConfig warns when `config` contains a field that the service expects in `auth`.
// The API keeps such fields in the auth section and never returns them, so the provider keeps
// the configured value in state and can't detect drift for it.
func warnAuthFieldsInConfig(configAttr types.Object, service types.String, configRoot path.Path, diags *diag.Diagnostics) {
	if configAttr.IsNull() || configAttr.IsUnknown() {
		return
	}
	if service.IsNull() || service.IsUnknown() {
		return
	}
	serviceName := service.ValueString()

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
		if !common.IsAuthOnlyFieldForService(serviceName, fn) {
			continue
		}
		diags.AddAttributeWarning(
			configRoot.AtName(fn),
			fmt.Sprintf("Field `%v` Should Be Set In `auth`", fn),
			fmt.Sprintf("Service `%v` expects field `%v` in the `auth` block, not in `config`. "+
				"The Fivetran API doesn't return auth fields, so the configured value is kept in state as is and changes made outside of Terraform aren't detected. "+
				"Move it to the `auth` block.", serviceName, fn),
		)
	}
}
