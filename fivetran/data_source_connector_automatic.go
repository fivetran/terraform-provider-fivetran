package fivetran

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConnectorAutomatic() *schema.Resource {
	var result = &schema.Resource{
		ReadContext: dataSourceConnectorAutomaticRead,
		Schema: map[string]*schema.Schema{
			"id":                 {Type: schema.TypeString, Required: true},
			"group_id":           {Type: schema.TypeString, Computed: true},
			"service":            {Type: schema.TypeString, Computed: true},
			"service_version":    {Type: schema.TypeString, Computed: true},
			"name":               {Type: schema.TypeString, Computed: true},
			"destination_schema": dataSourceConnectorAutomaticDestinationSchemaSchema(),
			"connected_by":       {Type: schema.TypeString, Computed: true},
			"created_at":         {Type: schema.TypeString, Computed: true},
			"succeeded_at":       {Type: schema.TypeString, Computed: true},
			"failed_at":          {Type: schema.TypeString, Computed: true},
			"sync_frequency":     {Type: schema.TypeString, Computed: true},
			"daily_sync_time":    {Type: schema.TypeString, Computed: true},
			"schedule_type":      {Type: schema.TypeString, Computed: true},
			"paused":             {Type: schema.TypeString, Computed: true},
			"pause_after_trial":  {Type: schema.TypeString, Computed: true},
			"status":             dataSourceConnectorAutomaticSchemaStatus(),
			"config":             dataSourceConnectorAutomaticSchemaConfig(),
		},
	}
	return result
}

func dataSourceConnectorAutomaticDestinationSchemaSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":   {Type: schema.TypeString, Computed: true},
				"table":  {Type: schema.TypeString, Computed: true},
				"prefix": {Type: schema.TypeString, Computed: true},
			},
		},
	}
}

func dataSourceConnectorAutomaticSchemaStatus() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"setup_state":        {Type: schema.TypeString, Computed: true},
				"sync_state":         {Type: schema.TypeString, Computed: true},
				"update_state":       {Type: schema.TypeString, Computed: true},
				"is_historical_sync": {Type: schema.TypeString, Computed: true},
				"tasks": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"code":    {Type: schema.TypeString, Computed: true},
							"message": {Type: schema.TypeString, Computed: true},
						},
					},
				},
				"warnings": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"code":    {Type: schema.TypeString, Computed: true},
							"message": {Type: schema.TypeString, Computed: true},
						},
					},
				},
			},
		},
	}
}

func dataSourceConnectorAutomaticSchemaConfig() *schema.Schema {
	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		oasProperties := getSchemaAndProperties(path)
		for key, value := range oasProperties {
			if existingValue, ok := properties[key]; ok {
				if existingValue.Type == schema.TypeList {
					if _, ok := existingValue.Elem.(map[string]*schema.Schema); ok {
						continue
					}
					value = updateExistingValue(existingValue, value)
				}
			}
			properties[key] = value
		}
	}

	return &schema.Schema{Type: schema.TypeList, Optional: true, Computed: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: properties,
		},
	}
}

func dataSourceConnectorAutomaticRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(d.Get("id").(string)).DoCustomMerged(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)
	mapAddStr(msi, "service", resp.Data.Service)
	mapAddStr(msi, "service_version", intPointerToStr(resp.Data.ServiceVersion))
	mapAddStr(msi, "name", resp.Data.Schema)
	mapAddXInterface(msi, "destination_schema", dataSourceConnectorAutomaticReadDestinationSchema(resp.Data.Schema, resp.Data.Service))
	mapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
	mapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
	mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
	mapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
	mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
	mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
	mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))
	mapAddXInterface(msi, "status", dataSourceConnectorAutomaticReadStatus(&resp))
	mapAddXInterface(msi, "config", dataSourceConnectorAutomaticReadConfig(&resp))
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func dataSourceConnectorAutomaticReadDestinationSchema(schema string, service string) []interface{} {
	return readDestinationSchema(schema, service)
}

// dataSourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func dataSourceConnectorAutomaticReadStatus(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	status := make([]interface{}, 1)

	s := make(map[string]interface{})
	mapAddStr(s, "setup_state", resp.Data.Status.SetupState)
	mapAddStr(s, "sync_state", resp.Data.Status.SyncState)
	mapAddStr(s, "update_state", resp.Data.Status.UpdateState)
	mapAddStr(s, "is_historical_sync", boolPointerToStr(resp.Data.Status.IsHistoricalSync))
	mapAddXInterface(s, "tasks", dataSourceConnectorReadStatusFlattenTasks(resp))
	mapAddXInterface(s, "warnings", dataSourceConnectorReadStatusFlattenWarnings(resp))
	status[0] = s

	return status
}

// dataSourceConnectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func dataSourceConnectorAutomaticReadConfig(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	configArray := make([]interface{}, 1)

	configResult := make(map[string]interface{})

	responseConfig := resp.Data.CustomConfig

	responseConfigFromStruct := structToMap(resp.Data.Config)
	for responseProperty, value := range responseConfigFromStruct {
		reflectedValue := reflect.ValueOf(value)
		if reflectedValue.Kind() == reflect.Slice && reflect.TypeOf(value).Elem().Kind() != reflect.String {
			var valueArray []interface{}
			for i := 0; i < reflectedValue.Len(); i++ {
				valueArray = append(valueArray, reflectedValue.Index(i).Interface())
			}

			childPropertiesFromStruct := structToMap(valueArray[0])
			valueArray[0] = childPropertiesFromStruct
			responseConfig[responseProperty] = valueArray
			continue
		}
		responseConfig[responseProperty] = value
	}

	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		newProperties := getSchemaAndProperties(path)
		for newPropertyKey, newPropertySchema := range newProperties {
			properties[newPropertyKey] = newPropertySchema
		}
	}

	for property, propertySchema := range properties {
		if propertySchema.Type == schema.TypeSet || propertySchema.Type == schema.TypeList {
			if values, ok := responseConfig[property].([]string); ok {
				configResult[property] = xStrXInterface(values)
				continue
			}
			if interfaceValues, ok := responseConfig[property].([]interface{}); ok {
				if _, ok := interfaceValues[0].(map[string]interface{}); ok {
					configResult[property] = interfaceValues
				} else {
					configResult[property] = xInterfaceStrXStr(interfaceValues)
				}
				continue
			}
		}
		if value, ok := responseConfig[property].(string); ok && value != "" {
			valueType := propertySchema.Type
			switch valueType {
			case schema.TypeBool:
				configResult[property] = strToBool(value)
			case schema.TypeInt:
				configResult[property] = strToInt(value)
			default:
				configResult[property] = value
			}
		}
	}
	configArray[0] = configResult

	return configArray
}

/*
This function will convert object from struct to map[string]interface{}
based on JSON tag in structs.
*/
func structToMap(item interface{}) map[string]interface{} {

	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToMap(field)
			} else {
				res[tag] = field
			}
		}
	}
	return res
}
