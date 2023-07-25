package fivetran

import (
	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type FieldValueType int64
type FieldType int64

const (
	String     FieldValueType = 0
	Integer    FieldValueType = 1
	Boolean    FieldValueType = 2
	StringList FieldValueType = 3
	ObjectList FieldValueType = 4
)

type configField struct {
	readonly       bool
	sensitive      bool
	nullable       bool
	fieldValueType FieldValueType
	itemFields     map[string]configField
	itemKeyField   string
	description    string
}

func NewconfigField() configField {
	field := configField{}
	field.fieldValueType = String
	field.readonly = false
	field.sensitive = false
	field.nullable = true
	field.description = ""
	return field
}

var configFields = map[string]configField{
	"latest_version":            {readonly: true},
	"authorization_method":      {readonly: true},
	"service_version":           {readonly: true},
	"last_synced_changes__utc_": {readonly: true},

	"public_key":  {},
	"external_id": {},

	"is_ftps":                           {nullable: false, fieldValueType: Boolean},
	"sftp_is_key_pair":                  {nullable: false, fieldValueType: Boolean},
	"sync_data_locker":                  {nullable: false, fieldValueType: Boolean},
	"enable_all_dimension_combinations": {nullable: false, fieldValueType: Boolean},
	"update_config_on_each_sync":        {nullable: false, fieldValueType: Boolean},
	"on_premise":                        {nullable: false, fieldValueType: Boolean},
	"use_api_keys":                      {nullable: false, fieldValueType: Boolean},
	"is_new_package":                    {nullable: false, fieldValueType: Boolean},
	"is_multi_entity_feature_enabled":   {nullable: false, fieldValueType: Boolean},
	"always_encrypted":                  {nullable: false, fieldValueType: Boolean},
	"is_secure":                         {nullable: false, fieldValueType: Boolean},
	"use_webhooks":                      {nullable: false, fieldValueType: Boolean},
	"eu_region":                         {nullable: false, fieldValueType: Boolean},
	"is_keypair":                        {nullable: false, fieldValueType: Boolean},
	"is_account_level_connector":        {nullable: false, fieldValueType: Boolean},
	"use_oracle_rac":                    {nullable: false, fieldValueType: Boolean},
	"asm_option":                        {nullable: false, fieldValueType: Boolean},
	"is_single_table_mode":              {nullable: false, fieldValueType: Boolean},
	"is_public":                         {nullable: false, fieldValueType: Boolean},
	"empty_header":                      {nullable: false, fieldValueType: Boolean},
	"support_nested_columns":            {nullable: false, fieldValueType: Boolean},
	"is_private_key_encrypted":          {nullable: false, fieldValueType: Boolean},

	"list_strategy":                        {nullable: false},
	"connection_type":                      {nullable: false},
	"sync_method":                          {nullable: false},
	"sync_mode":                            {nullable: false},
	"date_granularity":                     {nullable: false},
	"timeframe_months":                     {nullable: false},
	"report_type":                          {nullable: false},
	"aggregation":                          {nullable: false},
	"config_type":                          {nullable: false},
	"prebuilt_report":                      {nullable: false},
	"action_report_time":                   {nullable: false},
	"click_attribution_window":             {nullable: false},
	"view_attribution_window":              {nullable: false},
	"view_through_attribution_window_size": {nullable: false},
	"post_click_attribution_window_size":   {nullable: false},
	"update_method":                        {nullable: false},
	"swipe_attribution_window":             {nullable: false},
	"api_type":                             {nullable: false},
	"auth_type":                            {nullable: false},
	"sync_format":                          {nullable: false},
	"app_sync_mode":                        {nullable: false},
	"sales_account_sync_mode":              {nullable: false},
	"finance_account_sync_mode":            {nullable: false},
	"source":                               {nullable: false},
	"file_type":                            {nullable: false},
	"compression":                          {nullable: false},
	"on_error":                             {nullable: false},
	"append_file_option":                   {nullable: false},
	"engagement_attribution_window":        {nullable: false},
	"conversion_report_time":               {nullable: false},
	"data_access_method":                   {nullable: false},
	"sync_pack_mode":                       {nullable: false},
	"auth":                                 {nullable: false},
	"conversion_window_size":               {nullable: false, fieldValueType: Integer},
	"skip_before":                          {nullable: false, fieldValueType: Integer},
	"skip_after":                           {nullable: false, fieldValueType: Integer},
	"ftp_port":                             {nullable: false, fieldValueType: Integer},
	"sftp_port":                            {nullable: false, fieldValueType: Integer},
	"port":                                 {nullable: false, fieldValueType: Integer},
	"tunnel_port":                          {nullable: false, fieldValueType: Integer},
	"daily_api_call_limit":                 {nullable: false, fieldValueType: Integer},
	"api_quota":                            {nullable: false, fieldValueType: Integer},
	"agent_port":                           {nullable: false, fieldValueType: Integer},
	"replica_id":                           {nullable: false, fieldValueType: Integer},
	"network_code":                         {nullable: false, fieldValueType: Integer},

	"site_id":               {},
	"customer_list_id":      {},
	"asm_oracle_home":       {},
	"asm_tns":               {},
	"pdb_name":              {},
	"agent_host":            {},
	"agent_user":            {},
	"agent_public_cert":     {},
	"agent_ora_home":        {},
	"tns":                   {},
	"asm_user":              {},
	"sap_user":              {},
	"sheet_id":              {},
	"named_range":           {},
	"client_id":             {},
	"technical_account_id":  {},
	"organization_id":       {},
	"s3bucket":              {},
	"abs_connection_string": {},
	"abs_container_name":    {},
	"folder_id":             {},
	"ftp_host":              {},
	"ftp_user":              {},
	"sftp_host":             {},
	"sftp_user":             {},
	"bucket":                {},
	"prefix":                {},
	"pattern":               {},
	"archive_pattern":       {},
	"null_sequence":         {},
	"delimiter":             {},
	"escape_char":           {},
	"auth_mode":             {},
	"certificate":           {},
	"consumer_group":        {},

	"message_type":        {},
	"sync_type":           {},
	"security_protocol":   {},
	"access_key_id":       {},
	"home_folder":         {},
	"function":            {},
	"region":              {},
	"container_name":      {},
	"connection_string":   {},
	"function_app":        {},
	"function_name":       {},
	"function_key":        {},
	"merchant_id":         {},
	"api_url":             {},
	"cloud_storage_type":  {},
	"s3external_id":       {},
	"s3folder":            {},
	"gcs_bucket":          {},
	"gcs_folder":          {},
	"instance":            {},
	"aws_region_code":     {},
	"subdomain":           {},
	"host":                {},
	"user":                {},
	"customer_id":         {},
	"project_id":          {},
	"dataset_id":          {},
	"bucket_name":         {},
	"config_method":       {},
	"query_id":            {},
	"path":                {},
	"endpoint":            {},
	"identity":            {},
	"domain_name":         {},
	"resource_url":        {},
	"tunnel_host":         {},
	"tunnel_user":         {},
	"database":            {},
	"datasource":          {},
	"account":             {},
	"role":                {},
	"email":               {},
	"account_id":          {},
	"server_url":          {},
	"user_key":            {},
	"api_version":         {},
	"time_zone":           {},
	"integration_key":     {},
	"domain":              {},
	"replication_slot":    {},
	"publication_name":    {},
	"data_center":         {},
	"sub_domain":          {},
	"test_table_name":     {},
	"shop":                {},
	"sid":                 {},
	"key":                 {},
	"bucket_service":      {},
	"user_name":           {},
	"username":            {},
	"report_url":          {},
	"unique_id":           {},
	"base_url":            {},
	"entity_id":           {},
	"soap_uri":            {},
	"user_id":             {},
	"share_url":           {},
	"organization":        {},
	"access_key":          {},
	"domain_host_name":    {},
	"client_name":         {},
	"domain_type":         {},
	"connection_method":   {},
	"group_name":          {},
	"company_id":          {},
	"environment":         {},
	"csv_definition":      {},
	"export_storage_type": {},
	"uri":                 {},

	"app_ids":                     {fieldValueType: StringList},
	"servers":                     {fieldValueType: StringList},
	"report_suites":               {fieldValueType: StringList},
	"elements":                    {fieldValueType: StringList},
	"metrics":                     {fieldValueType: StringList},
	"advertisables":               {fieldValueType: StringList},
	"dimensions":                  {fieldValueType: StringList},
	"selected_exports":            {fieldValueType: StringList},
	"apps":                        {fieldValueType: StringList},
	"sales_accounts":              {fieldValueType: StringList},
	"finance_accounts":            {fieldValueType: StringList},
	"projects":                    {fieldValueType: StringList},
	"user_profiles":               {fieldValueType: StringList},
	"report_configuration_ids":    {fieldValueType: StringList},
	"accounts":                    {fieldValueType: StringList},
	"fields":                      {fieldValueType: StringList},
	"breakdowns":                  {fieldValueType: StringList},
	"action_breakdowns":           {fieldValueType: StringList},
	"pages":                       {fieldValueType: StringList},
	"repositories":                {fieldValueType: StringList},
	"dimension_attributes":        {fieldValueType: StringList},
	"columns":                     {fieldValueType: StringList},
	"manager_accounts":            {fieldValueType: StringList},
	"profiles":                    {fieldValueType: StringList},
	"site_urls":                   {fieldValueType: StringList},
	"advertisers_id":              {fieldValueType: StringList},
	"hosts":                       {fieldValueType: StringList},
	"advertisers":                 {fieldValueType: StringList},
	"organizations":               {fieldValueType: StringList},
	"account_ids":                 {fieldValueType: StringList},
	"packed_mode_tables":          {fieldValueType: StringList},
	"properties":                  {fieldValueType: StringList},
	"primary_keys":                {fieldValueType: StringList},
	"conversion_dimensions":       {fieldValueType: StringList},
	"custom_floodlight_variables": {fieldValueType: StringList},
	"partners":                    {fieldValueType: StringList},
	"per_interaction_dimensions":  {fieldValueType: StringList},
	"schema_registry_urls":        {fieldValueType: StringList},
	"segments":                    {fieldValueType: StringList},
	"topics":                      {fieldValueType: StringList},

	"short_code":         {sensitive: true},
	"passphrase":         {sensitive: true},
	"account_key":        {sensitive: true},
	"oauth_token":        {sensitive: true},
	"oauth_token_secret": {sensitive: true},
	"consumer_key":       {sensitive: true},
	"client_secret":      {sensitive: true},
	"private_key":        {sensitive: true},
	"s3role_arn":         {sensitive: true},
	"ftp_password":       {sensitive: true},
	"sftp_password":      {sensitive: true},
	"api_key":            {sensitive: true},
	"role_arn":           {sensitive: true},
	"password":           {sensitive: true},
	"secret_key":         {sensitive: true},
	"pem_certificate":    {sensitive: true},
	"access_token":       {sensitive: true},
	"api_secret":         {sensitive: true},
	"api_access_token":   {sensitive: true},
	"secret":             {sensitive: true},
	"consumer_secret":    {sensitive: true},
	"secrets":            {sensitive: true},
	"api_token":          {sensitive: true},
	"encryption_key":     {sensitive: true},
	"pat":                {sensitive: true},
	"function_trigger":   {sensitive: true},
	"token_key":          {sensitive: true},
	"token_secret":       {sensitive: true},
	"agent_password":     {sensitive: true},
	"asm_password":       {sensitive: true},
	"login_password":     {sensitive: true},
	"api_keys":           {sensitive: true, fieldValueType: StringList},

	"custom_tables": {
		fieldValueType: ObjectList,
		itemFields: map[string]configField{
			"table_name":               {},
			"config_type":              {nullable: false},
			"aggregation":              {nullable: false},
			"action_report_time":       {nullable: false},
			"click_attribution_window": {nullable: false},
			"view_attribution_window":  {nullable: false},
			"prebuilt_report_name":     {},
			"fields":                   {fieldValueType: StringList},
			"breakdowns":               {fieldValueType: StringList},
			"action_breakdowns":        {fieldValueType: StringList},
		},
	},

	"adobe_analytics_configurations": {
		fieldValueType: ObjectList,
		itemFields: map[string]configField{
			"sync_mode":          {nullable: false},
			"report_suites":      {fieldValueType: StringList},
			"elements":           {fieldValueType: StringList},
			"metrics":            {fieldValueType: StringList},
			"calculated_metrics": {fieldValueType: StringList},
			"segments":           {fieldValueType: StringList},
		},
	},

	"reports": {
		fieldValueType: ObjectList,
		itemFields: map[string]configField{
			"table":           {},
			"prebuilt_report": {},
			"filter":          {},
			"config_type":     {nullable: false},
			"report_type":     {nullable: false},
			"fields":          {fieldValueType: StringList},
			"dimensions":      {fieldValueType: StringList},
			"metrics":         {fieldValueType: StringList},
			"segments":        {fieldValueType: StringList},
		},
	},

	"secrets_list": {
		fieldValueType: ObjectList,
		itemKeyField:   "key",
		itemFields: map[string]configField{
			"key":   {},
			"value": {sensitive: true},
		},
	},

	"project_credentials": {
		fieldValueType: ObjectList,
		itemKeyField:   "project",
		itemFields: map[string]configField{
			"project":    {},
			"api_key":    {sensitive: true},
			"secret_key": {sensitive: true},
		},
	},
}

func getFieldSchema(isDataSourceSchema bool, field *configField) *schema.Schema {
	result := &schema.Schema{
		Type:      schema.TypeString,
		Optional:  !isDataSourceSchema,
		Computed:  isDataSourceSchema || !field.nullable,
		Sensitive: field.sensitive,
	}

	if field.readonly {
		if field.fieldValueType == StringList {
			result = &schema.Schema{Type: schema.TypeSet, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}}
		} else {
			result = &schema.Schema{Type: schema.TypeString, Computed: true}
		}
	} else {
		if field.fieldValueType == StringList {
			result = &schema.Schema{
				Type:      schema.TypeSet,
				Optional:  !isDataSourceSchema,
				Computed:  isDataSourceSchema,
				Sensitive: field.sensitive,
				Elem:      &schema.Schema{Type: schema.TypeString}}
		} else if field.fieldValueType == ObjectList {
			var elemSchema = map[string]*schema.Schema{}

			for k, v := range field.itemFields {
				elemSchema[k] = getFieldSchema(isDataSourceSchema, &v)
			}

			result = &schema.Schema{
				Type:     schema.TypeSet,
				Optional: !isDataSourceSchema,
				Computed: isDataSourceSchema,
				Elem: &schema.Resource{
					Schema: elemSchema,
				},
			}
		}
	}

	result.Description = field.description
	return result
}

func connectorSchemaConfig(readonly bool) *schema.Schema {
	var schemaMap = map[string]*schema.Schema{}

	for k, v := range configFields {
		schemaMap[k] = getFieldSchema(readonly, &v)
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: getMaxItems(readonly),
		Elem: &schema.Resource{
			Schema: schemaMap,
		},
	}
}

func tryCopySensitiveStringValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, name string) {
	if localConfig == nil {
		tryCopyStringValue(targetConfig, upstreamConfig, name)
	} else {
		tryCopyStringValue(targetConfig, *localConfig, name)
	}
}

func tryCopySensitiveListValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, name string) {
	if localConfig != nil {
		mapAddXInterface(targetConfig, name, (*localConfig)[name].(*schema.Set).List())
	} else {
		tryCopyList(targetConfig, upstreamConfig, name)
	}
}

// connectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func connectorReadCustomConfig(resp *fivetran.ConnectorCustomDetailsResponse, currentConfig *[]interface{}) []interface{} {
	c := make(map[string]interface{})
	var currentConfigMap *map[string]interface{} = nil

	if currentConfig != nil && len(*currentConfig) > 0 {
		vlocalConfigAsMap := (*currentConfig)[0].(map[string]interface{})
		currentConfigMap = &vlocalConfigAsMap
	}

	for k, v := range configFields {
		readFieldValueCore(k, v, currentConfigMap, c, resp.Data.Config)
	}

	return []interface{}{c}
}

func readFieldValueCore(k string, v configField, currentConfig *map[string]interface{}, c map[string]interface{}, upstream map[string]interface{}) {
	switch v.fieldValueType {
	case String:
		if v.sensitive {
			tryCopySensitiveStringValue(currentConfig, c, upstream, k)
		} else {
			tryCopyStringValue(c, upstream, k)
		}
	case Integer:
		tryCopyIntegerValue(c, upstream, k)
	case Boolean:
		tryCopyBooleanValue(c, upstream, k)
	case StringList:
		if v.sensitive {
			tryCopySensitiveListValue(currentConfig, c, upstream, k)
		} else {
			tryCopyList(c, upstream, k)
		}
	case ObjectList:
		var upstreamList = tryReadListValue(upstream, k)
		if upstreamList == nil || len(upstreamList) < 1 {
			mapAddXInterface(c, k, make([]interface{}, 0))
		} else {
			resultList := make([]interface{}, len(upstreamList))
			for i, elem := range upstreamList {
				upstreamElem := elem.(map[string]interface{})
				resultElem := make(map[string]interface{})
				var localElem *map[string]interface{}
				subKeyValue := tryReadValue(upstreamElem, v.itemKeyField)
				localElem = nil
				if currentConfig != nil && subKeyValue != nil {
					targetList := (*currentConfig)[k].(*schema.Set).List()

					var filterFunc = func(elem interface{}) bool {
						return elem.(map[string]interface{})[v.itemKeyField].(string) == subKeyValue
					}
					found := filterList(targetList, filterFunc)
					if found != nil {
						foundAsMap := (*found).(map[string]interface{})
						localElem = &foundAsMap
					}
				}

				for fn, fv := range v.itemFields {
					readFieldValueCore(fn, fv, localElem, resultElem, upstreamElem)
				}
				resultList[i] = resultElem
			}
			mapAddXInterface(c, k, resultList)
		}

	}
}

func connectorUpdateCustomConfig(c map[string]interface{}) *map[string]interface{} {
	configMap := make(map[string]interface{})
	for k, v := range c {
		if field, ok := configFields[k]; ok {
			updateConfigFieldImpl(k, field, v, configMap)
		}
	}
	return &configMap
}

func updateConfigFieldImpl(name string, field configField, v interface{}, configMap map[string]interface{}) {
	switch field.fieldValueType {
	case String:
		{
			if v.(string) != "" {
				configMap[name] = v
			}
		}
	case Integer:
		{
			if v.(string) != "" {
				configMap[name] = strToInt(v.(string))
			}
		}
	case StringList:
		{
			configMap[name] = xInterfaceStrXStr(v.(*schema.Set).List())
		}
	case Boolean:
		{
			if v.(string) != "" {
				configMap[name] = strToBool(v.(string))
			}
		}
	case ObjectList:
		{
			var list = v.(*schema.Set).List()
			result := make([]interface{}, len(list))
			for i, v := range list {
				vmap := v.(map[string]interface{})
				item := make(map[string]interface{})
				for subName, subField := range field.itemFields {
					updateConfigFieldImpl(subName, subField, vmap[subName], item)
				}
				result[i] = item
			}
			configMap[name] = result
		}
	}
}
