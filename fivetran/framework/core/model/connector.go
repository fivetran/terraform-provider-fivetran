package model

import (
    "fmt"
    "strings"

    gfcommon "github.com/fivetran/go-fivetran/common"
    "github.com/fivetran/go-fivetran/connections"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/common"
    "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/diag"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ConnectorDatasourceModel struct {
    Id          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    ConnectedBy types.String `tfsdk:"connected_by"`
    CreatedAt   types.String `tfsdk:"created_at"`
    GroupId     types.String `tfsdk:"group_id"`
    Service     types.String `tfsdk:"service"`

    DestinationSchema types.Object `tfsdk:"destination_schema"`

    SucceededAt     types.String `tfsdk:"succeeded_at"`
    FailedAt        types.String `tfsdk:"failed_at"`
    ServiceVersion  types.String `tfsdk:"service_version"`
    SyncFrequency   types.Int64  `tfsdk:"sync_frequency"`
    ScheduleType    types.String `tfsdk:"schedule_type"`
    Paused          types.Bool   `tfsdk:"paused"`
    PauseAfterTrial types.Bool   `tfsdk:"pause_after_trial"`
    DailySyncTime   types.String `tfsdk:"daily_sync_time"`
    
    DataDelaySensitivity    types.String `tfsdk:"data_delay_sensitivity"`
    DataDelayThreshold      types.Int64  `tfsdk:"data_delay_threshold"`

    ProxyAgentId             types.String `tfsdk:"proxy_agent_id"`
    NetworkingMethod         types.String `tfsdk:"networking_method"`
    HybridDeploymentAgentId  types.String `tfsdk:"hybrid_deployment_agent_id"`
    PrivateLinkId            types.String `tfsdk:"private_link_id"`
    Status types.Object `tfsdk:"status"`

    Config types.Object `tfsdk:"config"`
}

var (
    codeMessageAttrTypes = map[string]attr.Type{
        "code":    types.StringType,
        "message": types.StringType,
    }
)

func readCommonResponse(r gfcommon.CommonResponse) attr.Value {
    result, _ := types.ObjectValue(codeMessageAttrTypes,
        map[string]attr.Value{
            "code":    types.StringValue(r.Code),
            "message": types.StringValue(r.Message),
        })
    return result
}

func (d *ConnectorDatasourceModel) ReadFromResponse(resp connections.DetailsWithCustomConfigNoTestsResponse) {
    responseContainer := ConnectorModelContainer{}
    responseContainer.ReadFromResponseData(resp.Data.DetailsResponseDataCommon, resp.Data.Config)
    d.ReadFromContainer(responseContainer)

	d.SucceededAt = types.StringValue(resp.Data.SucceededAt.String())
	d.FailedAt = types.StringValue(resp.Data.FailedAt.String())
	d.ServiceVersion = types.StringValue(fmt.Sprintf("%v", *resp.Data.ServiceVersion))
	d.SyncFrequency = types.Int64Value(int64(*resp.Data.SyncFrequency))
	d.ScheduleType = types.StringValue(resp.Data.ScheduleType)
	d.Paused = types.BoolValue(*resp.Data.Paused)
	d.PauseAfterTrial = types.BoolValue(*resp.Data.PauseAfterTrial)
    
    d.DataDelaySensitivity = types.StringValue(resp.Data.DataDelaySensitivity)

    if resp.Data.DataDelayThreshold != nil {
        d.DataDelayThreshold = types.Int64Value(int64(*resp.Data.DataDelayThreshold))
    } else {
        d.DataDelayThreshold = types.Int64Null()
    }

	if resp.Data.DailySyncTime != "" {
		d.DailySyncTime = types.StringValue(resp.Data.DailySyncTime)
	} else {
		d.DailySyncTime = types.StringNull()
	}

    codeMessageAttrType := types.ObjectType{
        AttrTypes: codeMessageAttrTypes,
    }

    warns := []attr.Value{}
    for _, w := range resp.Data.Status.Warnings {
        warns = append(warns, readCommonResponse(w))
    }
    tasks := []attr.Value{}
    for _, t := range resp.Data.Status.Tasks {
        tasks = append(tasks, readCommonResponse(t))
    }

    wsV, _ := types.SetValue(codeMessageAttrType, warns)
    tsV, _ := types.SetValue(codeMessageAttrType, tasks)

    status, _ := types.ObjectValue(
        map[string]attr.Type{
            "setup_state":        types.StringType,
            "is_historical_sync": types.BoolType,
            "sync_state":         types.StringType,
            "update_state":       types.StringType,
            "tasks":              types.SetType{ElemType: codeMessageAttrType},
            "warnings":           types.SetType{ElemType: codeMessageAttrType},
        },
        map[string]attr.Value{
            "setup_state":        types.StringValue(resp.Data.Status.SetupState),
            "is_historical_sync": types.BoolPointerValue(resp.Data.Status.IsHistoricalSync),
            "sync_state":         types.StringValue(resp.Data.Status.SyncState),
            "update_state":       types.StringValue(resp.Data.Status.UpdateState),
            "warnings":           wsV,
            "tasks":              tsV,
        },
    )
    d.Status = status
}

type ConnectorResourceModel struct {
    Id                types.String `tfsdk:"id"`
    Name              types.String `tfsdk:"name"`
    ConnectedBy       types.String `tfsdk:"connected_by"`
    CreatedAt         types.String `tfsdk:"created_at"`
    GroupId           types.String `tfsdk:"group_id"`
    Service           types.String `tfsdk:"service"`
    DestinationSchema types.Object `tfsdk:"destination_schema"`

	ProxyAgentId           types.String `tfsdk:"proxy_agent_id"`
	NetworkingMethod       types.String `tfsdk:"networking_method"`
    HybridDeploymentAgentId  types.String `tfsdk:"hybrid_deployment_agent_id"`
    PrivateLinkId          types.String `tfsdk:"private_link_id"`

    DataDelaySensitivity    types.String `tfsdk:"data_delay_sensitivity"`
    DataDelayThreshold      types.Int64  `tfsdk:"data_delay_threshold"`

    Config   types.Object   `tfsdk:"config"`
    Auth     types.Object   `tfsdk:"auth"`
    Timeouts timeouts.Value `tfsdk:"timeouts"`

    RunSetupTests     types.Bool `tfsdk:"run_setup_tests"`
    TrustCertificates types.Bool `tfsdk:"trust_certificates"`
    TrustFingerprints types.Bool `tfsdk:"trust_fingerprints"`
}

func (d *ConnectorResourceModel) ReadFromResponse(resp connections.DetailsWithCustomConfigNoTestsResponse, isImporting bool) diag.Diagnostics {
    responseContainer := ConnectorModelContainer{}
    responseContainer.ReadFromResponseData(resp.Data.DetailsResponseDataCommon, resp.Data.Config)
    d.ReadFromContainer(responseContainer, isImporting)
    return nil
}

func (d *ConnectorResourceModel) ReadFromCreateResponse(resp connections.DetailsWithCustomConfigResponse) diag.Diagnostics {
    responseContainer := ConnectorModelContainer{}
    responseContainer.ReadFromResponseData(resp.Data.DetailsResponseDataCommon, resp.Data.Config)
    d.ReadFromContainer(responseContainer, false)
    return nil
}

func (d *ConnectorResourceModel) GetConfigMap(nullOnNull bool) (map[string]interface{}, error) {
    if d.Config.IsNull() && nullOnNull {
        return nil, nil
    }
    result := getValueFromAttrValue(d.Config, common.GetConfigFieldsMap(), nil, d.Service.ValueString()).(map[string]interface{})
    serviceName := d.Service.ValueString()
    serviceFields, err := common.GetFieldsForService(serviceName)
    if err != nil {
        return result, err
    }
    allFields := common.GetConfigFieldsMap()
    err = patchServiceSpecificFields(result, serviceName, serviceFields, allFields)
    return result, err
}

func (d *ConnectorResourceModel) GetAuthMap(nullOnNull bool) (map[string]interface{}, error) {
    if d.Auth.IsNull() && nullOnNull {
        return nil, nil
    }
    serviceName := d.Service.ValueString()
    serviceFields := common.GetAuthFieldsForService(serviceName)
    allFields := common.GetAuthFieldsMap()

    result := getValueFromAttrValue(d.Auth, allFields, nil, serviceName).(map[string]interface{})
    err := patchServiceSpecificFields(result, serviceName, serviceFields, allFields)
    return result, err
}

func (d *ConnectorResourceModel) GetDestinatonSchemaForConfig() (map[string]interface{}, error) {
    if d.DestinationSchema.IsNull() || d.DestinationSchema.IsUnknown() {
        return nil, fmt.Errorf("Field `destination_schema` is required.")
    }
    return getDestinatonSchemaForConfig(d.Service,
        d.DestinationSchema.Attributes()["name"],
        d.DestinationSchema.Attributes()["table"],
        d.DestinationSchema.Attributes()["prefix"],
        d.DestinationSchema.Attributes()["table_group_name"],
    )
}

func (d *ConnectorResourceModel) ReadFromContainer(c ConnectorModelContainer, isImporting bool) {
	d.Id = types.StringValue(c.Id)
	d.Name = types.StringValue(c.Schema)
	d.ConnectedBy = types.StringValue(c.ConnectedBy)
	d.CreatedAt = types.StringValue(c.CreatedAt)
	d.GroupId = types.StringValue(c.GroupId)
	d.Service = types.StringValue(c.Service)

    // as fact - this is computed attribute which user can change
    if !d.DataDelaySensitivity.IsUnknown() && !d.DataDelaySensitivity.IsNull() {
        d.DataDelaySensitivity = types.StringValue(c.DataDelaySensitivity)    
    }
    
    if c.DataDelayThreshold != nil {
        d.DataDelayThreshold = types.Int64Value(int64(*c.DataDelayThreshold))
    } else {
        d.DataDelayThreshold = types.Int64Null()
    }
    
    d.DestinationSchema = getDestinationSchemaValue(c.Service, c.Schema, d.DestinationSchema)

	if c.HybridDeploymentAgentId != "" && !d.HybridDeploymentAgentId.IsUnknown() && !d.HybridDeploymentAgentId.IsNull() {
		d.HybridDeploymentAgentId = types.StringValue(c.HybridDeploymentAgentId)
	} else {
		d.HybridDeploymentAgentId = types.StringNull()
	}

    if c.PrivateLinkId != "" {
        d.PrivateLinkId = types.StringValue(c.PrivateLinkId)
	} else {
        d.PrivateLinkId = types.StringNull()
	}

	if c.ProxyAgentId != "" {
		d.ProxyAgentId = types.StringValue(c.ProxyAgentId)
	} else {
		d.ProxyAgentId = types.StringNull()
	}

	if c.NetworkingMethod != "" {
		d.NetworkingMethod = types.StringValue(c.NetworkingMethod)
	}

	if isImporting || (!d.Config.IsNull() && !d.Config.IsUnknown()) {
		d.Config = getValue(
			types.ObjectType{AttrTypes: getAttrTypes(common.GetConfigFieldsMap())},
			c.Config,
			getValueFromAttrValue(d.Config, common.GetConfigFieldsMap(), nil, c.Service).(map[string]interface{}),
			common.GetConfigFieldsMap(), nil, c.Service, isImporting, false).(basetypes.ObjectValue)
	}
}

func (d *ConnectorDatasourceModel) ReadFromContainer(c ConnectorModelContainer) {
	d.Id = types.StringValue(c.Id)
	d.Name = types.StringValue(c.Schema)
	d.ConnectedBy = types.StringValue(c.ConnectedBy)
	d.CreatedAt = types.StringValue(c.CreatedAt)
	d.GroupId = types.StringValue(c.GroupId)
	d.Service = types.StringValue(c.Service)

    // as fact - this is computed attribute which user can change
    if !d.DataDelaySensitivity.IsUnknown() && !d.DataDelaySensitivity.IsNull() {
        d.DataDelaySensitivity = types.StringValue(c.DataDelaySensitivity)    
    }
    
    if c.DataDelayThreshold != nil {
        d.DataDelayThreshold = types.Int64Value(int64(*c.DataDelayThreshold))
    } else {
        d.DataDelayThreshold = types.Int64Null()
    }

    d.DestinationSchema = getDestinationSchemaValue(c.Service, c.Schema, d.DestinationSchema)
    
    if c.PrivateLinkId != "" {
        d.PrivateLinkId = types.StringValue(c.PrivateLinkId)
	} else {
        d.PrivateLinkId = types.StringNull()
	}

	if c.ProxyAgentId != "" {
		d.ProxyAgentId = types.StringValue(c.ProxyAgentId)
	} else {
		d.ProxyAgentId = types.StringNull()
	}

	if c.NetworkingMethod != "" {
		d.NetworkingMethod = types.StringValue(c.NetworkingMethod)
	}

    if c.HybridDeploymentAgentId != "" && !d.HybridDeploymentAgentId.IsUnknown() && !d.HybridDeploymentAgentId.IsNull() {
        d.HybridDeploymentAgentId = types.StringValue(c.HybridDeploymentAgentId)
    } else {
        d.HybridDeploymentAgentId = types.StringNull()
    }

    d.Config = getValue(
        types.ObjectType{AttrTypes: getAttrTypes(common.GetConfigFieldsMap())},
        c.Config,
        c.Config,
        common.GetConfigFieldsMap(),
        nil,
        c.Service, false, false).(basetypes.ObjectValue)
}

func (d *ConnectorResourceModel) HasUpdates(plan ConnectorResourceModel, state ConnectorResourceModel) (bool, map[string]interface{}, map[string]interface{}, error) {
    stateConfigMap, err := state.GetConfigMap(false)
    // this is not expected - state should contain only known fields relative to service
    // but we have to check error just in case
    if err != nil {
        return false, nil, nil, err
    }

    stateAuthMap, err := state.GetAuthMap(false)
    // this is not expected - state should contain only known fields relative to service
    // but we have to check error just in case
    if err != nil {
        return false, nil, nil, err
    }

    planConfigMap, err := plan.GetConfigMap(false)
    if err != nil {
        return false, nil, nil, err
    }

    planAuthMap, err := plan.GetAuthMap(false)
    if err != nil {
        return false, nil, nil, err
    }

    patch := PrepareConfigAuthPatch(stateConfigMap, planConfigMap, plan.Service.ValueString(), common.GetConfigFieldsMap())
    authPatch := PrepareConfigAuthPatch(stateAuthMap, planAuthMap, plan.Service.ValueString(), common.GetAuthFieldsMap())

    if len(patch) > 0 || 
            len(authPatch) > 0 || 
            !plan.ProxyAgentId.Equal(state.ProxyAgentId) ||
            !plan.PrivateLinkId.Equal(state.PrivateLinkId) ||
            !plan.HybridDeploymentAgentId.Equal(state.HybridDeploymentAgentId) ||
            !plan.DataDelaySensitivity.Equal(state.DataDelaySensitivity) ||
            !plan.DataDelayThreshold.Equal(state.DataDelayThreshold) ||
            !plan.NetworkingMethod.Equal(state.NetworkingMethod) {
                return true, patch, authPatch, nil
            } else {
                return false, nil, nil, nil
            }
}

type ConnectorModelContainer struct {
	Id          string
	Name        string
	ConnectedBy string
	CreatedAt   string
	GroupId     string
	Service     string
	Schema      string

    ProxyAgentId            string
    NetworkingMethod        string
    HybridDeploymentAgentId string
    PrivateLinkId           string

    DataDelaySensitivity string
    DataDelayThreshold   *int

	Config map[string]interface{}

    RunSetupTests     bool
    TrustCertificates bool
    TrustFingerprints bool
}

func (c *ConnectorModelContainer) ReadFromResponseData(data connections.DetailsResponseDataCommon, config map[string]interface{}) {
	c.Id = data.ID
	c.Name = data.Schema
	c.ConnectedBy = data.ConnectedBy
	c.CreatedAt = data.CreatedAt.String()
	c.GroupId = data.GroupID
	c.Service = data.Service
	c.Schema = data.Schema

    c.DataDelaySensitivity = data.DataDelaySensitivity
    c.DataDelayThreshold = data.DataDelayThreshold

	c.Config = config

	if data.ProxyAgentId != "" {
		c.ProxyAgentId = data.ProxyAgentId
	}

	if data.NetworkingMethod != "" {
		c.NetworkingMethod = data.NetworkingMethod
	}

    if data.PrivateLinkId != "" {
        c.PrivateLinkId = data.PrivateLinkId
	}

    if data.HybridDeploymentAgentId != "" {
        c.HybridDeploymentAgentId = data.HybridDeploymentAgentId
    }
}

func getDestinatonSchemaForConfig(serviceId, nameAttr, tableAttr, prefixAttr, tableGroupNameAttr attr.Value) (map[string]interface{}, error) {
	service := serviceId.(basetypes.StringValue).ValueString()
	if _, ok := common.GetDestinationSchemaFields()[service]; !ok {
		return nil, fmt.Errorf("unknown connector service: `%v`", service)
	}
	if common.GetDestinationSchemaFields()[service]["schema_prefix"] {
		if prefixAttr.IsNull() || prefixAttr.IsUnknown() {
			return nil, fmt.Errorf("`destination_schema.prefix` field is required to create `%v` connector", service)
		}
		if !nameAttr.IsNull() {
			return nil, fmt.Errorf("`destination_schema.name` field can't be set for `%v` connector", service)
		}
		if !tableAttr.IsNull() {
			return nil, fmt.Errorf("`destination_schema.table` field can't be set for `%v` connector", service)
		}
        if !tableGroupNameAttr.IsNull() {
            return nil, fmt.Errorf("`destination_schema.table_group_name` field can't be set for `%v` connector", service)
        }
		prefix := prefixAttr.(types.String).ValueString()
		return map[string]interface{}{
			"schema_prefix": prefix,
		}, nil
	} else {
		if nameAttr.IsNull() {
			return nil, fmt.Errorf("`destination_schema.name` field is required to create `%v` connector", service)
		}
		result := map[string]interface{}{
			"schema": nameAttr.(types.String).ValueString(),
		}
        if common.GetDestinationSchemaFields()[service]["table"] {
            if !tableAttr.IsNull() && tableAttr.(types.String).ValueString() != "" {
                result["table"] = tableAttr.(types.String).ValueString()
            }
        }

        if common.GetDestinationSchemaFields()[service]["table_group_name"] {
            if !tableGroupNameAttr.IsNull() && tableGroupNameAttr.(types.String).ValueString() != "" {
                result["table_group_name"] = tableGroupNameAttr.(types.String).ValueString()
            }
        }

		return result, nil
	}
}

func getDestinationSchemaValue(service, schema string, destinationSchema types.Object ) types.Object {
    r, _ := types.ObjectValue(
        map[string]attr.Type{
            "name":             types.StringType,
            "table":            types.StringType,
            "prefix":           types.StringType,
            "table_group_name": types.StringType,
        },
        getDestinationSchemaValuesMap(service, schema, destinationSchema),
    )
    return r
}

func getDestinationSchemaValuesMap(service, schema string, destinationSchema types.Object) map[string]attr.Value {
    if _, ok := common.GetDestinationSchemaFields()[service]; !ok {
        panic(fmt.Errorf("unknown connector service: `%v`", service))
    }

    if common.GetDestinationSchemaFields()[service]["schema_prefix"] {
        return map[string]attr.Value{
            "name":             types.StringNull(),
            "table":            types.StringNull(),
            "prefix":           types.StringValue(schema),
            "table_group_name": types.StringNull(),
        }
    } else {
        result := map[string]attr.Value{
            "table":            types.StringNull(),
            "prefix":           types.StringNull(),
            "table_group_name": types.StringNull(),
        }
        s := strings.Split(schema, ".")
        result["name"] = types.StringValue(s[0])
        if len(s) > 1 {
            if common.GetDestinationSchemaFields()[service]["table_group_name"] &&
                !destinationSchema.IsNull() && 
                !destinationSchema.Attributes()["table_group_name"].IsNull() && 
                !destinationSchema.Attributes()["table_group_name"].IsUnknown() {
                result["table_group_name"] = types.StringValue(s[1])                
            }

            if common.GetDestinationSchemaFields()[service]["table"] &&
               !destinationSchema.IsNull() && 
               !destinationSchema.Attributes()["table"].IsNull() && 
               !destinationSchema.Attributes()["table"].IsUnknown() {
                result["table"] = types.StringValue(s[1])                
            }
        }

        return result
    }
}
