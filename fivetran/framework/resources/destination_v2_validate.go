package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fivetran/go-fivetran/metadata"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *destinationV2) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.GetSkipPlanTimeValidation() {
		return
	}

	var data model.DestinationV2ResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Service.IsNull() || data.Service.IsUnknown() || data.Service.ValueString() == "" {
		return
	}

	configMap, diags := core.DynamicToMapPreserveUnknown(ctx, data.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	meta, err := r.destinationMetadata(ctx, data.Service.ValueString())
	if err != nil {
		if errors.Is(err, errUnconfiguredClient) {
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Validate Destination V2 Configuration",
			fmt.Sprintf("Unable to fetch metadata for service %q. Terraform cannot safely validate dynamic config fields without metadata. Fix metadata access or set provider skip_plan_time_validation = true to bypass this check temporarily. Original error: %v", data.Service.ValueString(), err),
		)
		return
	}

	validateDynamicObject(configMap, &meta.Config, path.Root("config"), &resp.Diagnostics)
	rejectReadonlyDynamicFields(configMap, &meta.Config, path.Root("config"), &resp.Diagnostics)
	validateDestinationV2RootFields(data, meta, &resp.Diagnostics)
}

// rejectReadonlyDynamicFields reports readonly fields set in HCL. The API ignores
// them on write and returns its own value on read, so configuring one could only
// produce a permanent diff.
func rejectReadonlyDynamicFields(values map[string]interface{}, slot *metadata.Property, root path.Path, diags *diag.Diagnostics) {
	if slot == nil {
		return
	}
	for name, value := range values {
		prop := core.SlotProp(slot, name)
		if prop == nil {
			continue
		}
		fieldPath := root.AtName(name)
		if prop.Readonly {
			diags.AddAttributeError(
				fieldPath,
				"Readonly Dynamic Field",
				fmt.Sprintf("Field %q is readonly: its value is set by Fivetran and cannot be configured. Remove it from config.", name),
			)
			continue
		}
		if nested, ok := value.(map[string]interface{}); ok {
			rejectReadonlyDynamicFields(nested, prop, fieldPath, diags)
		}
	}
}

// validateDestinationV2RootFields checks root fields against the capabilities the
// destination metadata advertises. These are warnings, not errors: the metadata
// lists what the destination type supports, while the API also applies
// account-level rules (allowed regions, PrivateLink permission) at apply time.
func validateDestinationV2RootFields(data model.DestinationV2ResourceModel, meta *metadata.DestinationMetadata, diags *diag.Diagnostics) {
	service := data.Service.ValueString()

	if value, ok := knownNonEmptyString(data.Region); ok && len(meta.SupportedRegions) > 0 && !containsString(meta.SupportedRegions, value) {
		diags.AddAttributeWarning(
			path.Root("region"),
			"Unsupported Destination Region",
			fmt.Sprintf("Destination metadata for service %q does not list region %q. Supported regions: %s.", service, value, strings.Join(meta.SupportedRegions, ", ")),
		)
	}

	if value, ok := knownNonEmptyString(data.NetworkingMethod); ok && len(meta.SupportedNetworkingMethods) > 0 && !containsString(meta.SupportedNetworkingMethods, value) {
		diags.AddAttributeWarning(
			path.Root("networking_method"),
			"Unsupported Destination Networking Method",
			fmt.Sprintf("Destination metadata for service %q does not list networking method %q. Supported networking methods: %s.", service, value, strings.Join(meta.SupportedNetworkingMethods, ", ")),
		)
	}

	if _, ok := knownNonEmptyString(data.HybridDeploymentAgentId); ok && !meta.SupportsHybridDeployment {
		diags.AddAttributeWarning(
			path.Root("hybrid_deployment_agent_id"),
			"Hybrid Deployment Not Supported",
			fmt.Sprintf("Destination metadata for service %q reports that hybrid deployment is not supported.", service),
		)
	}

	if _, ok := knownNonEmptyString(data.PrivateLinkId); ok && !containsString(meta.PrivateLinkTypes, "SELF_SERVICE") {
		diags.AddAttributeWarning(
			path.Root("private_link_id"),
			"PrivateLink ID Not Supported",
			fmt.Sprintf("Destination metadata for service %q does not list self-service PrivateLink, which private_link_id requires.", service),
		)
	}
}

func knownNonEmptyString(value types.String) (string, bool) {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return "", false
	}
	return value.ValueString(), true
}

func containsString(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
