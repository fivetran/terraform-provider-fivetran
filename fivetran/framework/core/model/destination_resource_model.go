package model

import (
	"github.com/fivetran/go-fivetran/destinations"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type DestinationResourceModel struct {
	Id            				 types.String 	`tfsdk:"id"`
	GroupId       				 types.String 	`tfsdk:"group_id"`
	Service       				 types.String 	`tfsdk:"service"`
	Region        				 types.String 	`tfsdk:"region"`
	TimeZoneOffset				 types.String 	`tfsdk:"time_zone_offset"`
	SetupStatus    				 types.String 	`tfsdk:"setup_status"`
	DaylightSavingTimeEnabled	 types.Bool   	`tfsdk:"daylight_saving_time_enabled"`
	Config  					 types.Object   `tfsdk:"config"`
	Timeouts					 timeouts.Value `tfsdk:"timeouts"`
    HybridDeploymentAgentId      types.String   `tfsdk:"hybrid_deployment_agent_id"`
    NetworkingMethod             types.String   `tfsdk:"networking_method"`
    PrivateLinkId                types.String `tfsdk:"private_link_id"`

	RunSetupTests    			 types.Bool 	`tfsdk:"run_setup_tests"`
	TrustCertificates			 types.Bool 	`tfsdk:"trust_certificates"`
	TrustFingerprints			 types.Bool 	`tfsdk:"trust_fingerprints"`
}

var _ destinationModel = &DestinationResourceModel{}

func (d *DestinationResourceModel) SetId(value string) {
	d.Id = types.StringValue(value)
}
func (d *DestinationResourceModel) SetGroupId(value string) {
	d.GroupId = types.StringValue(value)
}
func (d *DestinationResourceModel) SetService(value string) {
	d.Service = types.StringValue(value)
}
func (d *DestinationResourceModel) SetRegion(value string) {
	d.Region = types.StringValue(value)
}
func (d *DestinationResourceModel) SetTimeZonOffset(value string) {
	d.TimeZoneOffset = types.StringValue(value)
}
func (d *DestinationResourceModel) SetSetupStatus(value string) {
	d.SetupStatus = types.StringValue(value)
}
func (d *DestinationResourceModel) SetDaylightSavingTimeEnabled(value bool) {
	d.DaylightSavingTimeEnabled = types.BoolValue(value)
}
func (d *DestinationResourceModel) SetHybridDeploymentAgentId(value string) {
    if value != "" {
        d.HybridDeploymentAgentId = types.StringValue(value)
    } else {
        d.HybridDeploymentAgentId = types.StringNull()
    }
}
func (d *DestinationResourceModel) SetNetworkingMethod(value string) {
    if value != "" {
        d.NetworkingMethod = types.StringValue(value)
    }
}
func (d *DestinationResourceModel) SetPrivateLinkId(value string) {
    if value != "" {
        d.PrivateLinkId = types.StringValue(value)
    } else {
        d.PrivateLinkId = types.StringNull()
    }
}
func (d *DestinationResourceModel) SetConfig(value map[string]interface{}) {
	if d.Service.IsNull() || d.Service.IsUnknown() {
		panic("Service type is null. Can't handle config without service type.")
	}
	// WA for inconsistent BQ response - it returns just "location" instead of "data_set_location"
	if l, ok := value["location"]; ok {
		value["data_set_location"] = l
	}

	service := d.Service.ValueString()
	config := d.Config
	d.Config = getValue(
		types.ObjectType{AttrTypes: getAttrTypes(common.GetDestinationFieldsMap())},
		value,
		getValueFromAttrValue(config, common.GetDestinationFieldsMap(), nil, service).(map[string]interface{}),
		common.GetDestinationFieldsMap(),
		nil,
		service, false).(basetypes.ObjectValue)
}

func (d *DestinationResourceModel) ReadFromResponse(resp destinations.DestinationDetailsCustomResponse) {
	var model destinationModel = d
	readFromResponse(model, resp.Data.DestinationDetailsBase, resp.Data.Config)
}

func (d *DestinationResourceModel) ReadFromResponseWithTests(resp destinations.DestinationDetailsWithSetupTestsCustomResponse) {
	var model destinationModel = d
	readFromResponse(model, resp.Data.DestinationDetailsBase, resp.Data.Config)
}

func (d *DestinationResourceModel) ReadFromLegacyResponse(resp destinations.DestinationDetailsWithSetupTestsResponse) {
	var model destinationModel = d
	readFromResponse(model, resp.Data.DestinationDetailsBase, map[string]interface{}{})
}

func (d *DestinationResourceModel) GetConfigMap(nullOnNull bool) (map[string]interface{}, error) {
	if d.Config.IsNull() && nullOnNull {
		return nil, nil
	}
	result := getValueFromAttrValue(d.Config, common.GetDestinationFieldsMap(), nil, d.Service.ValueString()).(map[string]interface{})
	serviceName := d.Service.ValueString()
	serviceFields := common.GetDestinationFieldsForService(serviceName)
	allFields := common.GetDestinationFieldsMap()
	err := patchServiceSpecificFields(result, serviceName, serviceFields, allFields)
	return result, err
}

func (d *DestinationResourceModel) HasUpdates(plan DestinationResourceModel, state DestinationResourceModel) (bool, map[string]interface{}, error) {
    stateConfigMap, err := state.GetConfigMap(false)
    // this is not expected - state should contain only known fields relative to service
    // but we have to check error just in case
    if err != nil {
        return false, nil, err
    }

    planConfigMap, err := plan.GetConfigMap(false)
    if err != nil {
        return false, nil, err
    }

    patch := PrepareConfigAuthPatch(stateConfigMap, planConfigMap, plan.Service.ValueString(), common.GetConfigFieldsMap())

    if len(patch) > 0 || 
            !plan.TimeZoneOffset.Equal(state.TimeZoneOffset) ||
            !plan.DaylightSavingTimeEnabled.Equal(state.DaylightSavingTimeEnabled) ||
            !plan.Region.Equal(state.Region) ||
            !plan.HybridDeploymentAgentId.Equal(state.HybridDeploymentAgentId) ||
            !plan.NetworkingMethod.Equal(state.NetworkingMethod) ||
            !plan.PrivateLinkId.Equal(state.PrivateLinkId) {
                return true, patch, nil
            } else {
                return false, nil, nil
            }
}
