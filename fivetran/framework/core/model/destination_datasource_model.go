package model

import (
	"github.com/fivetran/go-fivetran/destinations"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type DestinationDatasourceModel struct {
	Id             types.String `tfsdk:"id"`
	GroupId        types.String `tfsdk:"group_id"`
	Service        types.String `tfsdk:"service"`
	Region         types.String `tfsdk:"region"`
	TimeZoneOffset types.String `tfsdk:"time_zone_offset"`
	SetupStatus    types.String `tfsdk:"setup_status"`
	Config         types.Object `tfsdk:"config"`
}

var _ destinationModel = &DestinationDatasourceModel{}

func (d *DestinationDatasourceModel) SetId(value string) {
	d.Id = types.StringValue(value)
}
func (d *DestinationDatasourceModel) SetGroupId(value string) {
	d.GroupId = types.StringValue(value)
}
func (d *DestinationDatasourceModel) SetService(value string) {
	d.Service = types.StringValue(value)
}
func (d *DestinationDatasourceModel) SetRegion(value string) {
	d.Region = types.StringValue(value)
}
func (d *DestinationDatasourceModel) SetTimeZonOffset(value string) {
	d.TimeZoneOffset = types.StringValue(value)
}
func (d *DestinationDatasourceModel) SetSetupStatus(value string) {
	d.SetupStatus = types.StringValue(value)
}

func (d *DestinationDatasourceModel) SetConfig(value map[string]interface{}) {
	if d.Service.IsNull() || d.Service.IsUnknown() {
		panic("Service type is null. Can't handle config without service type.")
	}
	// WA for inconsistent BQ response - it returns just "location" instead of "data_set_location"
	if l, ok := value["location"]; ok {
		value["data_set_location"] = l
	}
	service := d.Service.ValueString()
	d.Config = getValue(
		types.ObjectType{AttrTypes: getAttrTypes(common.GetDestinationFieldsMap())},
		value,
		value,
		common.GetDestinationFieldsMap(),
		nil,
		service).(basetypes.ObjectValue)
}

func (d *DestinationDatasourceModel) ReadFromResponse(resp destinations.DestinationDetailsCustomResponse) {
	var model destinationModel = d
	readFromResponse(model, resp.Data.DestinationDetailsBase, resp.Data.Config)
}
