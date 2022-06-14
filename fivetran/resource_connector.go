package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		UpdateContext: resourceConnectorUpdate,
		DeleteContext: resourceConnectorDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":                 {Type: schema.TypeString, Computed: true},
			"group_id":           {Type: schema.TypeString, Required: true, ForceNew: true},
			"service":            {Type: schema.TypeString, Required: true, ForceNew: true},
			"service_version":    {Type: schema.TypeString, Computed: true},
			"destination_schema": resourceConnectorDestinationSchemaSchema(),
			"name":               {Type: schema.TypeString, Computed: true},
			"connected_by":       {Type: schema.TypeString, Computed: true},
			"created_at":         {Type: schema.TypeString, Computed: true},
			"succeeded_at":       {Type: schema.TypeString, Computed: true},
			"failed_at":          {Type: schema.TypeString, Computed: true},
			"sync_frequency":     {Type: schema.TypeString, Required: true},
			"daily_sync_time":    {Type: schema.TypeString, Optional: true},
			"schedule_type":      {Type: schema.TypeString, Computed: true},
			"trust_certificates": {Type: schema.TypeString, Optional: true},
			"trust_fingerprints": {Type: schema.TypeString, Optional: true},
			"run_setup_tests":    {Type: schema.TypeString, Optional: true},
			"paused":             {Type: schema.TypeString, Required: true},
			"pause_after_trial":  {Type: schema.TypeString, Required: true},
			"status":             resourceConnectorSchemaStatus(),
			"config":             resourceConnectorSchemaConfig(),
			"auth":               resourceConnectorSchemaAuth(),
			"last_updated":       {Type: schema.TypeString, Computed: true}, // internal
		},
	}
}

func resourceConnectorDestinationSchemaSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":   {Type: schema.TypeString, Optional: true, ForceNew: true},
				"table":  {Type: schema.TypeString, Optional: true, ForceNew: true},
				"prefix": {Type: schema.TypeString, Optional: true, ForceNew: true},
			},
		},
	}
}

func resourceConnectorSchemaStatus() *schema.Schema {
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

func resourceConnectorSchemaConfig() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Required: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Readonly config fields
				"latest_version":            {Type: schema.TypeString, Computed: true},
				"authorization_method":      {Type: schema.TypeString, Computed: true},
				"service_version":           {Type: schema.TypeString, Computed: true},
				"last_synced_changes__utc_": {Type: schema.TypeString, Computed: true},

				// Sensitive config fields, Fivetran returns this fields masked
				"oauth_token":        {Type: schema.TypeString, Optional: true, Sensitive: true},
				"oauth_token_secret": {Type: schema.TypeString, Optional: true, Sensitive: true},
				"consumer_key":       {Type: schema.TypeString, Optional: true, Sensitive: true},
				"client_secret":      {Type: schema.TypeString, Optional: true, Sensitive: true},
				"private_key":        {Type: schema.TypeString, Optional: true, Sensitive: true},
				"s3role_arn":         {Type: schema.TypeString, Optional: true, Sensitive: true},
				"ftp_password":       {Type: schema.TypeString, Optional: true, Sensitive: true},
				"sftp_password":      {Type: schema.TypeString, Optional: true, Sensitive: true},
				"api_key":            {Type: schema.TypeString, Optional: true, Sensitive: true},
				"role_arn":           {Type: schema.TypeString, Optional: true, Sensitive: true},
				"password":           {Type: schema.TypeString, Optional: true, Sensitive: true},
				"secret_key":         {Type: schema.TypeString, Optional: true, Sensitive: true},
				"pem_certificate":    {Type: schema.TypeString, Optional: true, Sensitive: true},
				"access_token":       {Type: schema.TypeString, Optional: true, Sensitive: true},
				"api_secret":         {Type: schema.TypeString, Optional: true, Sensitive: true},
				"api_access_token":   {Type: schema.TypeString, Optional: true, Sensitive: true},
				"secret":             {Type: schema.TypeString, Optional: true, Sensitive: true},
				"consumer_secret":    {Type: schema.TypeString, Optional: true, Sensitive: true},
				"secrets":            {Type: schema.TypeString, Optional: true, Sensitive: true},
				"api_token":          {Type: schema.TypeString, Optional: true, Sensitive: true},
				"encryption_key":     {Type: schema.TypeString, Optional: true, Sensitive: true},

				// Fields that are always have default value (and should be marked as Computed to prevent drifting)
				// Boolean values
				"is_ftps":                           {Type: schema.TypeString, Optional: true, Computed: true},
				"sftp_is_key_pair":                  {Type: schema.TypeString, Optional: true, Computed: true},
				"sync_data_locker":                  {Type: schema.TypeString, Optional: true, Computed: true},
				"enable_all_dimension_combinations": {Type: schema.TypeString, Optional: true, Computed: true},
				"update_config_on_each_sync":        {Type: schema.TypeString, Optional: true, Computed: true},
				"on_premise":                        {Type: schema.TypeString, Optional: true, Computed: true},
				"use_api_keys":                      {Type: schema.TypeString, Optional: true, Computed: true},
				"is_new_package":                    {Type: schema.TypeString, Optional: true, Computed: true},
				"is_multi_entity_feature_enabled":   {Type: schema.TypeString, Optional: true, Computed: true},
				"always_encrypted":                  {Type: schema.TypeString, Optional: true, Computed: true},
				"is_secure":                         {Type: schema.TypeString, Optional: true, Computed: true},
				"use_webhooks":                      {Type: schema.TypeString, Optional: true, Computed: true},
				"eu_region":                      	 {Type: schema.TypeString, Optional: true, Computed: true},
				// Enum & int values
				"connection_type":                      {Type: schema.TypeString, Optional: true, Computed: true},
				"sync_mode":                            {Type: schema.TypeString, Optional: true, Computed: true},
				"date_granularity":                     {Type: schema.TypeString, Optional: true, Computed: true},
				"timeframe_months":                     {Type: schema.TypeString, Optional: true, Computed: true},
				"report_type":                          {Type: schema.TypeString, Optional: true, Computed: true},
				"aggregation":                          {Type: schema.TypeString, Optional: true, Computed: true},
				"config_type":                          {Type: schema.TypeString, Optional: true, Computed: true},
				"prebuilt_report":                      {Type: schema.TypeString, Optional: true, Computed: true},
				"action_report_time":                   {Type: schema.TypeString, Optional: true, Computed: true},
				"click_attribution_window":             {Type: schema.TypeString, Optional: true, Computed: true},
				"view_attribution_window":              {Type: schema.TypeString, Optional: true, Computed: true},
				"conversion_window_size":               {Type: schema.TypeString, Optional: true, Computed: true},
				"view_through_attribution_window_size": {Type: schema.TypeString, Optional: true, Computed: true},
				"post_click_attribution_window_size":   {Type: schema.TypeString, Optional: true, Computed: true},
				"update_method":                        {Type: schema.TypeString, Optional: true, Computed: true},
				"swipe_attribution_window":             {Type: schema.TypeString, Optional: true, Computed: true},
				"api_type":                             {Type: schema.TypeString, Optional: true, Computed: true},
				"auth_type":                            {Type: schema.TypeString, Optional: true, Computed: true},
				"sync_format":                          {Type: schema.TypeString, Optional: true, Computed: true},
				"app_sync_mode":                        {Type: schema.TypeString, Optional: true, Computed: true},
				"sales_account_sync_mode":              {Type: schema.TypeString, Optional: true, Computed: true},
				"finance_account_sync_mode":            {Type: schema.TypeString, Optional: true, Computed: true},
				"source":                               {Type: schema.TypeString, Optional: true, Computed: true},
				"file_type":                            {Type: schema.TypeString, Optional: true, Computed: true},
				"compression":                          {Type: schema.TypeString, Optional: true, Computed: true},
				"on_error":                             {Type: schema.TypeString, Optional: true, Computed: true},
				"append_file_option":                   {Type: schema.TypeString, Optional: true, Computed: true},
				"engagement_attribution_window":        {Type: schema.TypeString, Optional: true, Computed: true},
				"conversion_report_time":               {Type: schema.TypeString, Optional: true, Computed: true},
				"skip_before":                          {Type: schema.TypeString, Optional: true, Computed: true},
				"skip_after":                           {Type: schema.TypeString, Optional: true, Computed: true},

				// For db-like connectors it's a readonly field, but it also used in Braintree connector as public field
				"public_key": {Type: schema.TypeString, Optional: true, Computed: true},

				// external_id is used among AWS connectors and correcponds to group_id. For some connectors it is computed, some expect it as a parameter.
				"external_id": {Type: schema.TypeString, Optional: true, Computed: true},

				"sheet_id":              {Type: schema.TypeString, Optional: true},
				"named_range":           {Type: schema.TypeString, Optional: true},
				"client_id":             {Type: schema.TypeString, Optional: true},
				"technical_account_id":  {Type: schema.TypeString, Optional: true},
				"organization_id":       {Type: schema.TypeString, Optional: true},
				"s3bucket":              {Type: schema.TypeString, Optional: true},
				"abs_connection_string": {Type: schema.TypeString, Optional: true},
				"abs_container_name":    {Type: schema.TypeString, Optional: true},
				"folder_id":             {Type: schema.TypeString, Optional: true},
				"ftp_host":              {Type: schema.TypeString, Optional: true},
				"ftp_port":              {Type: schema.TypeString, Optional: true},
				"ftp_user":              {Type: schema.TypeString, Optional: true},
				"sftp_host":             {Type: schema.TypeString, Optional: true},
				"sftp_port":             {Type: schema.TypeString, Optional: true},
				"sftp_user":             {Type: schema.TypeString, Optional: true},
				"bucket":                {Type: schema.TypeString, Optional: true},
				"prefix":                {Type: schema.TypeString, Optional: true},
				"pattern":               {Type: schema.TypeString, Optional: true},
				"archive_pattern":       {Type: schema.TypeString, Optional: true},
				"null_sequence":         {Type: schema.TypeString, Optional: true},
				"delimiter":             {Type: schema.TypeString, Optional: true},
				"escape_char":           {Type: schema.TypeString, Optional: true},
				"auth_mode":             {Type: schema.TypeString, Optional: true},

				"certificate":          {Type: schema.TypeString, Optional: true},
				"consumer_group":       {Type: schema.TypeString, Optional: true},
				"servers":              {Type: schema.TypeString, Optional: true},
				"message_type":         {Type: schema.TypeString, Optional: true},
				"sync_type":            {Type: schema.TypeString, Optional: true},
				"security_protocol":    {Type: schema.TypeString, Optional: true},
				"access_key_id":        {Type: schema.TypeString, Optional: true},
				"home_folder":          {Type: schema.TypeString, Optional: true},
				"function":             {Type: schema.TypeString, Optional: true},
				"region":               {Type: schema.TypeString, Optional: true},
				"container_name":       {Type: schema.TypeString, Optional: true},
				"connection_string":    {Type: schema.TypeString, Optional: true},
				"function_app":         {Type: schema.TypeString, Optional: true},
				"function_name":        {Type: schema.TypeString, Optional: true},
				"function_key":         {Type: schema.TypeString, Optional: true},
				"merchant_id":          {Type: schema.TypeString, Optional: true},
				"api_url":              {Type: schema.TypeString, Optional: true},
				"cloud_storage_type":   {Type: schema.TypeString, Optional: true},
				"s3external_id":        {Type: schema.TypeString, Optional: true},
				"s3folder":             {Type: schema.TypeString, Optional: true},
				"gcs_bucket":           {Type: schema.TypeString, Optional: true},
				"gcs_folder":           {Type: schema.TypeString, Optional: true},
				"instance":             {Type: schema.TypeString, Optional: true},
				"aws_region_code":      {Type: schema.TypeString, Optional: true},
				"subdomain":            {Type: schema.TypeString, Optional: true},
				"host":                 {Type: schema.TypeString, Optional: true},
				"port":                 {Type: schema.TypeString, Optional: true},
				"user":                 {Type: schema.TypeString, Optional: true},
				"network_code":         {Type: schema.TypeString, Optional: true},
				"customer_id":          {Type: schema.TypeString, Optional: true},
				"project_id":           {Type: schema.TypeString, Optional: true},
				"dataset_id":           {Type: schema.TypeString, Optional: true},
				"bucket_name":          {Type: schema.TypeString, Optional: true},
				"function_trigger":     {Type: schema.TypeString, Optional: true},
				"config_method":        {Type: schema.TypeString, Optional: true},
				"query_id":             {Type: schema.TypeString, Optional: true},
				"path":                 {Type: schema.TypeString, Optional: true},
				"endpoint":             {Type: schema.TypeString, Optional: true},
				"identity":             {Type: schema.TypeString, Optional: true},
				"api_quota":            {Type: schema.TypeString, Optional: true},
				"domain_name":          {Type: schema.TypeString, Optional: true},
				"resource_url":         {Type: schema.TypeString, Optional: true},
				"tunnel_host":          {Type: schema.TypeString, Optional: true},
				"tunnel_port":          {Type: schema.TypeString, Optional: true},
				"tunnel_user":          {Type: schema.TypeString, Optional: true},
				"database":             {Type: schema.TypeString, Optional: true},
				"datasource":           {Type: schema.TypeString, Optional: true},
				"account":              {Type: schema.TypeString, Optional: true},
				"role":                 {Type: schema.TypeString, Optional: true},
				"email":                {Type: schema.TypeString, Optional: true},
				"account_id":           {Type: schema.TypeString, Optional: true},
				"server_url":           {Type: schema.TypeString, Optional: true},
				"user_key":             {Type: schema.TypeString, Optional: true},
				"api_version":          {Type: schema.TypeString, Optional: true},
				"daily_api_call_limit": {Type: schema.TypeString, Optional: true},
				"time_zone":            {Type: schema.TypeString, Optional: true},
				"integration_key":      {Type: schema.TypeString, Optional: true},
				"domain":               {Type: schema.TypeString, Optional: true},
				"replication_slot":     {Type: schema.TypeString, Optional: true},
				"publication_name":     {Type: schema.TypeString, Optional: true},
				"data_center":          {Type: schema.TypeString, Optional: true},
				"sub_domain":           {Type: schema.TypeString, Optional: true},
				"test_table_name":      {Type: schema.TypeString, Optional: true},
				"shop":                 {Type: schema.TypeString, Optional: true},
				"sid":                  {Type: schema.TypeString, Optional: true},
				"key":                  {Type: schema.TypeString, Optional: true},
				"bucket_service":       {Type: schema.TypeString, Optional: true},
				"user_name":            {Type: schema.TypeString, Optional: true},
				"username":             {Type: schema.TypeString, Optional: true},
				"report_url":           {Type: schema.TypeString, Optional: true},
				"unique_id":            {Type: schema.TypeString, Optional: true},
				"base_url":             {Type: schema.TypeString, Optional: true},
				"entity_id":            {Type: schema.TypeString, Optional: true},
				"soap_uri":             {Type: schema.TypeString, Optional: true},
				"user_id":              {Type: schema.TypeString, Optional: true},

				// Collections
				"report_suites":            {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"elements":                 {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"metrics":                  {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"advertisables":            {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"dimensions":               {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"selected_exports":         {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"apps":                     {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"sales_accounts":           {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"finance_accounts":         {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"projects":                 {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"user_profiles":            {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"report_configuration_ids": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"accounts":                 {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"fields":                   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"breakdowns":               {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"action_breakdowns":        {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"pages":                    {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"repositories":             {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"dimension_attributes":     {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"columns":                  {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"manager_accounts":         {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"profiles":                 {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"site_urls":                {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"api_keys":                 {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"advertisers_id":           {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"hosts":                    {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"advertisers":              {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"organizations":            {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"account_ids":              {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},

				"adobe_analytics_configurations": {Type: schema.TypeList, Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sync_mode":          {Type: schema.TypeString, Optional: true},
							"report_suites":      {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"elements":           {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"metrics":            {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"calculated_metrics": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"segments":           {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
						},
					},
				},
				"reports": {Type: schema.TypeList, Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"table":           {Type: schema.TypeString, Optional: true},
							"config_type":     {Type: schema.TypeString, Optional: true, Computed: true},
							"prebuilt_report": {Type: schema.TypeString, Optional: true},
							"report_type":     {Type: schema.TypeString, Optional: true, Computed: true},
							"fields":          {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"dimensions":      {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"metrics":         {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"segments":        {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"filter":          {Type: schema.TypeString, Optional: true},
						},
					},
				},
				"custom_tables": {Type: schema.TypeList, Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"table_name":               {Type: schema.TypeString, Optional: true},
							"config_type":              {Type: schema.TypeString, Optional: true, Computed: true},
							"fields":                   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"breakdowns":               {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"action_breakdowns":        {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"aggregation":              {Type: schema.TypeString, Optional: true, Computed: true},
							"action_report_time":       {Type: schema.TypeString, Optional: true, Computed: true},
							"click_attribution_window": {Type: schema.TypeString, Optional: true, Computed: true},
							"view_attribution_window":  {Type: schema.TypeString, Optional: true, Computed: true},
							"prebuilt_report_name":     {Type: schema.TypeString, Optional: true},
						},
					},
				},
				"project_credentials": {Type: schema.TypeList, Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"project":    {Type: schema.TypeString, Optional: true},
							"api_key":    {Type: schema.TypeString, Optional: true, Sensitive: true},
							"secret_key": {Type: schema.TypeString, Optional: true, Sensitive: true},
						},
					},
				},
			},
		},
	}
}

func resourceConnectorSchemaAuth() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Optional: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_access": {Type: schema.TypeList, Optional: true, MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_id":       {Type: schema.TypeString, Optional: true},
							"client_secret":   {Type: schema.TypeString, Optional: true, Sensitive: true},
							"user_agent":      {Type: schema.TypeString, Optional: true},
							"developer_token": {Type: schema.TypeString, Optional: true, Sensitive: true},
						},
					},
				},
				"refresh_token": {Type: schema.TypeString, Optional: true, Sensitive: true},
				"access_token":  {Type: schema.TypeString, Optional: true, Sensitive: true},
				"realm_id":      {Type: schema.TypeString, Optional: true, Sensitive: true},
			},
		},
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorCreate()

	svc.GroupID(d.Get("group_id").(string))

	currentService := d.Get("service").(string)

	if currentService == "adwords" {
		return newDiagAppend(diags, diag.Error, "create error", "service `adwords` has been deprecated, use `google_ads` instead")
	}

	svc.Service(currentService)
	svc.TrustCertificates(strToBool(d.Get("trust_certificates").(string)))
	svc.TrustFingerprints(strToBool(d.Get("trust_fingerprints").(string)))
	svc.RunSetupTests(strToBool(d.Get("run_setup_tests").(string)))
	svc.Paused(strToBool(d.Get("paused").(string)))
	svc.PauseAfterTrial(strToBool(d.Get("pause_after_trial").(string)))
	svc.SyncFrequency(strToInt(d.Get("sync_frequency").(string)))
	if d.Get("sync_frequency") == "1440" {
		svc.DailySyncTime(d.Get("daily_sync_time").(string))
	}

	fivetranConfig := resourceConnectorUpdateConfig(d, true)

	svc.Config(resourceConnectorCreateConfig(fivetranConfig, d.Get("destination_schema").([]interface{})))

	svc.Auth(resourceConnectorCreateAuth(d.Get("auth").([]interface{})))

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceConnectorRead(ctx, d, m)

	return diags
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)

	currentService := d.Get("service").(string)

	// ignore service change for migrated `adwords connectors
	if currentService == "adwords" && resp.Data.Service == "google_ads" {
		mapAddStr(msi, "service", "adwords")
		diags = newDiagAppend(diags, diag.Warning, "Google Ads service migration detected", "service update supressed to prevent resource re-creation.")
	} else {
		mapAddStr(msi, "service", resp.Data.Service)
	}

	mapAddStr(msi, "service_version", intPointerToStr(resp.Data.ServiceVersion))
	mapAddStr(msi, "name", resp.Data.Schema)
	mapAddXInterface(msi, "destination_schema", resourceConnectorReadDestinationSchema(resp.Data.Schema, resp.Data.Service))
	mapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
	mapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
	mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
	if *resp.Data.SyncFrequency == 1440 {
		mapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
	} else {
		mapAddStr(msi, "daily_sync_time", d.Get("daily_sync_time").(string))
	}
	mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
	mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
	mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))
	mapAddXInterface(msi, "status", resourceConnectorReadStatus(&resp))
	mapAddXInterface(msi, "config", resourceConnectorReadConfig(&resp, d.Get("config").([]interface{})))
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorModify()

	svc.ConnectorID(d.Get("id").(string))

	if d.HasChange("sync_frequency") {
		svc.SyncFrequency(strToInt(d.Get("sync_frequency").(string)))
	}
	if d.HasChange("trust_certificates") {
		svc.TrustCertificates(strToBool(d.Get("trust_certificates").(string)))
	}
	if d.HasChange("trust_fingerprints") {
		svc.TrustFingerprints(strToBool(d.Get("trust_fingerprints").(string)))
	}
	if d.HasChange("run_setup_tests") {
		svc.RunSetupTests(strToBool(d.Get("run_setup_tests").(string)))
	}
	if d.HasChange("paused") {
		svc.Paused(strToBool(d.Get("paused").(string)))
	}
	if d.HasChange("pause_after_trial") {
		svc.PauseAfterTrial(strToBool(d.Get("pause_after_trial").(string)))
	}

	svc.Config(resourceConnectorUpdateConfig(d, false))
	svc.Auth(resourceConnectorCreateAuth(d.Get("auth").([]interface{})))

	resp, err := svc.Do(ctx)
	if err != nil {
		// resourceConnectorRead here makes sure the state is updated after a NewConnectorModify error.
		diags = resourceConnectorRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceConnectorRead(ctx, d, m)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorDelete()

	resp, err := svc.ConnectorID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}

func resourceConnectorUpdateConfig(d *schema.ResourceData, creation bool) *fivetran.ConnectorConfig {
	fivetranConfig := fivetran.NewConnectorConfig()
	var config = d.Get("config").([]interface{})

	if len(config) < 1 {
		return fivetranConfig
	}
	if config[0] == nil {
		return fivetranConfig
	}

	c := config[0].(map[string]interface{})

	if v := c["sheet_id"].(string); v != "" {
		fivetranConfig.SheetID(v)
	}
	if v := c["named_range"].(string); v != "" {
		fivetranConfig.NamedRange(v)
	}
	if v := c["client_id"].(string); v != "" {
		fivetranConfig.ClientID(v)
	}
	if v := c["client_secret"].(string); v != "" {
		fivetranConfig.ClientSecret(v)
	}
	if v := c["technical_account_id"].(string); v != "" {
		fivetranConfig.TechnicalAccountID(v)
	}
	if v := c["organization_id"].(string); v != "" {
		fivetranConfig.OrganizationID(v)
	}
	if v := c["private_key"].(string); v != "" {
		fivetranConfig.PrivateKey(v)
	}
	if v := c["sync_mode"].(string); v != "" {
		fivetranConfig.SyncMode(v)
	}
	if v := c["report_suites"].([]interface{}); len(v) > 0 {
		fivetranConfig.ReportSuites(xInterfaceStrXStr(v))
	}
	if v := c["elements"].([]interface{}); len(v) > 0 {
		fivetranConfig.Elements(xInterfaceStrXStr(v))
	}
	if v := c["metrics"].([]interface{}); len(v) > 0 {
		fivetranConfig.Metrics(xInterfaceStrXStr(v))
	}
	if v := c["date_granularity"].(string); v != "" {
		fivetranConfig.DateGranularity(v)
	}
	if v := c["timeframe_months"].(string); v != "" {
		fivetranConfig.TimeframeMonths(v)
	}
	if v := c["source"].(string); v != "" {
		fivetranConfig.Source(v)
	}
	if v := c["s3bucket"].(string); v != "" {
		fivetranConfig.S3Bucket(v)
	}
	if v := c["s3role_arn"].(string); v != "" {
		fivetranConfig.S3RoleArn(v)
	}
	if v := c["abs_connection_string"].(string); v != "" {
		fivetranConfig.ABSConnectionString(v)
	}
	if v := c["abs_container_name"].(string); v != "" {
		fivetranConfig.ABSContainerName(v)
	}
	if v := c["folder_id"].(string); v != "" {
		fivetranConfig.FolderId(v)
	}
	if v := c["ftp_host"].(string); v != "" {
		fivetranConfig.FTPHost(v)
	}
	if v := c["ftp_port"].(string); v != "" {
		fivetranConfig.FTPPort(strToInt(v))
	}
	if v := c["ftp_user"].(string); v != "" {
		fivetranConfig.FTPUser(v)
	}
	if v := c["ftp_password"].(string); v != "" {
		fivetranConfig.FTPPassword(v)
	}
	if v := c["is_ftps"].(string); v != "" {
		fivetranConfig.IsFTPS(strToBool(v))
	}
	if v := c["sftp_host"].(string); v != "" {
		fivetranConfig.SFTPHost(v)
	}
	if v := c["sftp_port"].(string); v != "" {
		fivetranConfig.SFTPPort(strToInt(v))
	}
	if v := c["sftp_user"].(string); v != "" {
		fivetranConfig.SFTPUser(v)
	}
	if v := c["sftp_password"].(string); v != "" {
		fivetranConfig.SFTPPassword(v)
	}
	if v := c["sftp_is_key_pair"].(string); v != "" {
		fivetranConfig.SFTPIsKeyPair(strToBool(v))
	}
	if v := c["advertisables"].([]interface{}); len(v) > 0 {
		fivetranConfig.Advertisables(xInterfaceStrXStr(v))
	}
	if v := c["report_type"].(string); v != "" {
		fivetranConfig.ReportType(v)
	}
	if v := c["dimensions"].([]interface{}); len(v) > 0 {
		fivetranConfig.Dimensions(xInterfaceStrXStr(v))
	}
	if v := c["api_key"].(string); v != "" {
		fivetranConfig.APIKey(v)
	}
	if v := c["external_id"].(string); v != "" {
		fivetranConfig.ExternalID(v)
	}
	if v := c["role_arn"].(string); v != "" {
		fivetranConfig.RoleArn(v)
	}
	if v := c["bucket"].(string); v != "" {
		fivetranConfig.Bucket(v)
	}
	if v := c["prefix"].(string); v != "" {
		fivetranConfig.Prefix(v)
	}
	if v := c["pattern"].(string); v != "" {
		fivetranConfig.Pattern(v)
	}
	if v := c["file_type"].(string); v != "" {
		fivetranConfig.FileType(v)
	}
	if v := c["compression"].(string); v != "" {
		fivetranConfig.Compression(v)
	}
	if v := c["on_error"].(string); v != "" {
		fivetranConfig.OnError(v)
	}
	if v := c["append_file_option"].(string); v != "" {
		fivetranConfig.AppendFileOption(v)
	}
	if v := c["archive_pattern"].(string); v != "" {
		fivetranConfig.ArchivePattern(v)
	}
	if v := c["null_sequence"].(string); v != "" {
		fivetranConfig.NullSequence(v)
	}
	if v := c["delimiter"].(string); v != "" {
		fivetranConfig.Delimiter(v)
	}
	if v := c["escape_char"].(string); v != "" {
		fivetranConfig.EscapeChar(v)
	}
	if v := c["skip_before"].(string); v != "" {
		fivetranConfig.SkipBefore(strToInt(v))
	}
	if v := c["skip_after"].(string); v != "" {
		fivetranConfig.SkipAfter(strToInt(v))
	}
	if v := c["project_credentials"].([]interface{}); len(v) > 0 {
		fivetranConfig.ProjectCredentials(resourceConnectorCreateConfigProjectCredentials(v))
	}
	if v := c["auth_mode"].(string); v != "" {
		fivetranConfig.AuthMode(v)
	}
	if v := c["username"].(string); v != "" {
		fivetranConfig.Username(v)
	}
	if v := c["user_name"].(string); v != "" {
		fivetranConfig.UserName(v)
	}
	if v := c["password"].(string); v != "" {
		fivetranConfig.Password(v)
	}
	if v := c["certificate"].(string); v != "" {
		fivetranConfig.Certificate(v)
	}
	if v := c["selected_exports"].([]interface{}); len(v) > 0 {
		fivetranConfig.SelectedExports(xInterfaceStrXStr(v))
	}
	if v := c["consumer_group"].(string); v != "" {
		fivetranConfig.ConsumerGroup(v)
	}
	if v := c["servers"].(string); v != "" {
		fivetranConfig.Servers(v)
	}
	if v := c["message_type"].(string); v != "" {
		fivetranConfig.MessageType(v)
	}
	if v := c["sync_type"].(string); v != "" {
		fivetranConfig.SyncType(v)
	}
	if v := c["security_protocol"].(string); v != "" {
		fivetranConfig.SecurityProtocol(v)
	}
	if v := c["apps"].([]interface{}); len(v) > 0 {
		fivetranConfig.Apps(xInterfaceStrXStr(v))
	}
	if v := c["sales_accounts"].([]interface{}); len(v) > 0 {
		fivetranConfig.SalesAccounts(xInterfaceStrXStr(v))
	}
	if v := c["finance_accounts"].([]interface{}); len(v) > 0 {
		fivetranConfig.FinanceAccounts(xInterfaceStrXStr(v))
	}
	if v := c["app_sync_mode"].(string); v != "" {
		fivetranConfig.AppSyncMode(v)
	}
	if v := c["sales_account_sync_mode"].(string); v != "" {
		fivetranConfig.SalesAccountSyncMode(v)
	}
	if v := c["finance_account_sync_mode"].(string); v != "" {
		fivetranConfig.FinanceAccountSyncMode(v)
	}
	if v := c["pem_certificate"].(string); v != "" {
		fivetranConfig.PEMCertificate(v)
	}
	if v := c["access_key_id"].(string); v != "" {
		fivetranConfig.AccessKeyID(v)
	}
	if v := c["secret_key"].(string); v != "" {
		fivetranConfig.SecretKey(v)
	}
	if v := c["home_folder"].(string); v != "" {
		fivetranConfig.HomeFolder(v)
	}
	if v := c["sync_data_locker"].(string); v != "" {
		fivetranConfig.SyncDataLocker(strToBool(v))
	}
	if v := c["projects"].([]interface{}); len(v) > 0 {
		fivetranConfig.Projects(xInterfaceStrXStr(v))
	}
	if v := c["function"].(string); v != "" {
		fivetranConfig.Function(v)
	}
	if v := c["region"].(string); v != "" {
		fivetranConfig.Region(v)
	}
	if v := c["secrets"].(string); v != "" {
		fivetranConfig.Secrets(v)
	}
	if v := c["container_name"].(string); v != "" {
		fivetranConfig.ContainerName(v)
	}
	if v := c["connection_string"].(string); v != "" {
		fivetranConfig.ConnectionString(v)
	}
	if v := c["function_app"].(string); v != "" {
		fivetranConfig.FunctionApp(v)
	}
	if v := c["function_name"].(string); v != "" {
		fivetranConfig.FunctionName(v)
	}
	if v := c["function_key"].(string); v != "" {
		fivetranConfig.FunctionKey(v)
	}
	if v := c["merchant_id"].(string); v != "" {
		fivetranConfig.MerchantID(v)
	}
	if v := c["api_url"].(string); v != "" {
		fivetranConfig.APIURL(v)
	}
	if v := c["cloud_storage_type"].(string); v != "" {
		fivetranConfig.CloudStorageType(v)
	}
	if v := c["s3external_id"].(string); v != "" {
		fivetranConfig.S3ExternalID(v)
	}
	if v := c["s3folder"].(string); v != "" {
		fivetranConfig.S3Folder(v)
	}
	if v := c["gcs_bucket"].(string); v != "" {
		fivetranConfig.GCSBucket(v)
	}
	if v := c["gcs_folder"].(string); v != "" {
		fivetranConfig.GCSFolder(v)
	}
	if v := c["user_profiles"].([]interface{}); len(v) > 0 {
		fivetranConfig.UserProfiles(xInterfaceStrXStr(v))
	}
	if v := c["report_configuration_ids"].([]interface{}); len(v) > 0 {
		fivetranConfig.ReportConfigurationIDs(xInterfaceStrXStr(v))
	}
	if v := c["enable_all_dimension_combinations"].(string); v != "" {
		fivetranConfig.EnableAllDimensionCombinations(strToBool(v))
	}
	if v := c["instance"].(string); v != "" {
		fivetranConfig.Instance(v)
	}
	if v := c["aws_region_code"].(string); v != "" {
		fivetranConfig.AWSRegionCode(v)
	}
	if v := c["accounts"].([]interface{}); len(v) > 0 {
		fivetranConfig.Accounts(xInterfaceStrXStr(v))
	}
	if v := c["fields"].([]interface{}); len(v) > 0 {
		fivetranConfig.Fields(xInterfaceStrXStr(v))
	}
	if v := c["breakdowns"].([]interface{}); len(v) > 0 {
		fivetranConfig.Breakdowns(xInterfaceStrXStr(v))
	}
	if v := c["action_breakdowns"].([]interface{}); len(v) > 0 {
		fivetranConfig.ActionBreakdowns(xInterfaceStrXStr(v))
	}
	if v := c["aggregation"].(string); v != "" {
		fivetranConfig.Aggregation(v)
	}
	if v := c["config_type"].(string); v != "" {
		fivetranConfig.ConfigType(v)
	}
	if v := c["prebuilt_report"].(string); v != "" {
		fivetranConfig.PrebuiltReport(v)
	}
	if v := c["action_report_time"].(string); v != "" {
		fivetranConfig.ActionReportTime(v)
	}
	if v := c["click_attribution_window"].(string); v != "" {
		fivetranConfig.ClickAttributionWindow(v)
	}
	if v := c["view_attribution_window"].(string); v != "" {
		fivetranConfig.ViewAttributionWindow(v)
	}
	if v := c["custom_tables"].([]interface{}); len(v) > 0 {
		fivetranConfig.CustomTables(resourceConnectorCreateConfigCustomTables(v))
	}
	if v := c["pages"].([]interface{}); len(v) > 0 {
		fivetranConfig.Pages(xInterfaceStrXStr(v))
	}
	if v := c["subdomain"].(string); v != "" {
		fivetranConfig.SubDomain(v)
	}
	if v := c["port"].(string); v != "" {
		fivetranConfig.Port(strToInt(v))
	}
	if v := c["user"].(string); v != "" {
		fivetranConfig.User(v)
	}
	if v := c["is_secure"].(string); v != "" {
		fivetranConfig.IsSecure(v)
	}
	if v := c["repositories"].([]interface{}); len(v) > 0 {
		fivetranConfig.Repositories(xInterfaceStrXStr(v))
	}
	if v := c["use_webhooks"].(string); v != "" {
		fivetranConfig.UseWebhooks(strToBool(v))
	}
	if v := c["dimension_attributes"].([]interface{}); len(v) > 0 {
		fivetranConfig.DimensionAttributes(xInterfaceStrXStr(v))
	}
	if v := c["columns"].([]interface{}); len(v) > 0 {
		fivetranConfig.Columns(xInterfaceStrXStr(v))
	}
	if v := c["network_code"].(string); v != "" {
		fivetranConfig.NetworkCode(v)
	}
	if v := c["customer_id"].(string); v != "" {
		fivetranConfig.CustomerID(v)
	}
	if v := c["manager_accounts"].([]interface{}); len(v) > 0 {
		fivetranConfig.ManagerAccounts(xInterfaceStrXStr(v))
	}
	if v := c["reports"].([]interface{}); len(v) > 0 {
		fivetranConfig.Reports(resourceConnectorCreateConfigReports(v))
	}
	if v := c["conversion_window_size"].(string); v != "" {
		fivetranConfig.ConversionWindowSize(strToInt(v))
	}
	if v := c["profiles"].([]interface{}); len(v) > 0 {
		fivetranConfig.Profiles(xInterfaceStrXStr(v))
	}
	if v := c["project_id"].(string); v != "" {
		fivetranConfig.ProjectID(v)
	}
	if v := c["dataset_id"].(string); v != "" {
		fivetranConfig.DatasetID(v)
	}
	if v := c["bucket_name"].(string); v != "" {
		fivetranConfig.BucketName(v)
	}
	if v := c["function_trigger"].(string); v != "" {
		fivetranConfig.FunctionTrigger(v)
	}
	if v := c["config_method"].(string); v != "" {
		fivetranConfig.ConfigMethod(v)
	}
	if v := c["query_id"].(string); v != "" {
		fivetranConfig.QueryID(v)
	}
	if v := c["update_config_on_each_sync"].(string); v != "" {
		fivetranConfig.UpdateConfigOnEachSync(strToBool(v))
	}
	if v := c["site_urls"].([]interface{}); len(v) > 0 {
		fivetranConfig.SiteURLs(xInterfaceStrXStr(v))
	}
	if v := c["path"].(string); v != "" {
		fivetranConfig.Path(v)
	}
	if v := c["on_premise"].(string); v != "" {
		fivetranConfig.OnPremise(strToBool(v))
	}
	if v := c["access_token"].(string); v != "" {
		fivetranConfig.AccessToken(v)
	}
	if v := c["view_through_attribution_window_size"].(string); v != "" {
		fivetranConfig.ViewThroughAttributionWindowSize(v)
	}
	if v := c["post_click_attribution_window_size"].(string); v != "" {
		fivetranConfig.PostClickAttributionWindowSize(v)
	}
	if v := c["use_api_keys"].(string); v != "" {
		fivetranConfig.UseAPIKeys(v)
	}
	if v := c["api_keys"].([]interface{}); len(v) > 0 {
		fivetranConfig.APIKeys(xInterfaceStrXStr(v))
	}
	if v := c["endpoint"].(string); v != "" {
		fivetranConfig.Endpoint(v)
	}
	if v := c["identity"].(string); v != "" {
		fivetranConfig.Identity(v)
	}
	if v := c["api_quota"].(string); v != "" {
		fivetranConfig.APIQuota(strToInt(v))
	}
	if v := c["domain_name"].(string); v != "" {
		fivetranConfig.DomainName(v)
	}
	if v := c["resource_url"].(string); v != "" {
		fivetranConfig.ResourceURL(v)
	}
	if v := c["api_secret"].(string); v != "" {
		fivetranConfig.APISecret(v)
	}
	if v := c["host"].(string); v != "" {
		fivetranConfig.Host(v)
	}
	if v := c["hosts"].([]interface{}); len(v) > 0 {
		fivetranConfig.Hosts(xInterfaceStrXStr(v))
	}
	if v := c["tunnel_host"].(string); v != "" {
		fivetranConfig.TunnelHost(v)
	}
	if v := c["tunnel_port"].(string); v != "" {
		fivetranConfig.TunnelPort(strToInt(v))
	}
	if v := c["tunnel_user"].(string); v != "" {
		fivetranConfig.TunnelUser(v)
	}
	if v := c["database"].(string); v != "" {
		fivetranConfig.Database(v)
	}
	if v := c["datasource"].(string); v != "" {
		fivetranConfig.Datasource(v)
	}
	if v := c["account"].(string); v != "" {
		fivetranConfig.Account(v)
	}
	if v := c["role"].(string); v != "" {
		fivetranConfig.Role(v)
	}
	if v := c["email"].(string); v != "" {
		fivetranConfig.Email(v)
	}
	if v := c["account_id"].(string); v != "" {
		fivetranConfig.AccountID(v)
	}
	if v := c["server_url"].(string); v != "" {
		fivetranConfig.ServerURL(v)
	}
	if v := c["user_key"].(string); v != "" {
		fivetranConfig.UserKey(v)
	}
	if v := c["api_version"].(string); v != "" {
		fivetranConfig.APIVersion(v)
	}
	if v := c["daily_api_call_limit"].(string); v != "" {
		fivetranConfig.DailyAPICallLimit(strToInt(v))
	}
	if v := c["time_zone"].(string); v != "" {
		fivetranConfig.TimeZone(v)
	}
	if v := c["integration_key"].(string); v != "" {
		fivetranConfig.IntegrationKey(v)
	}
	if v := c["advertisers"].([]interface{}); len(v) > 0 {
		fivetranConfig.Advertisers(xInterfaceStrXStr(v))
	}
	if v := c["engagement_attribution_window"].(string); v != "" {
		fivetranConfig.EngagementAttributionWindow(v)
	}
	if v := c["conversion_report_time"].(string); v != "" {
		fivetranConfig.ConversionReportTime(v)
	}
	if v := c["domain"].(string); v != "" {
		fivetranConfig.Domain(v)
	}
	if v := c["update_method"].(string); v != "" {
		fivetranConfig.UpdateMethod(v)
	}
	if v := c["connection_type"].(string); v != "" {
		fivetranConfig.ConnectionType(v)
	}
	if v := c["replication_slot"].(string); v != "" {
		fivetranConfig.ReplicationSlot(v)
	}
	if v := c["publication_name"].(string); v != "" {
		fivetranConfig.PublicationName(v)
	}
	if v := c["data_center"].(string); v != "" {
		fivetranConfig.DataCenter(v)
	}
	if v := c["api_token"].(string); v != "" {
		fivetranConfig.APIToken(v)
	}
	if v := c["sub_domain"].(string); v != "" {
		fivetranConfig.SubDomain(v)
	}
	if v := c["test_table_name"].(string); v != "" {
		fivetranConfig.TestTableName(v)
	}
	if v := c["shop"].(string); v != "" {
		fivetranConfig.Shop(v)
	}
	if v := c["organizations"].([]interface{}); len(v) > 0 {
		fivetranConfig.Organizations(xInterfaceStrXStr(v))
	}
	if v := c["swipe_attribution_window"].(string); v != "" {
		fivetranConfig.SwipeAttributionWindow(v)
	}
	if v := c["api_access_token"].(string); v != "" {
		fivetranConfig.APIAccessToken(v)
	}
	if v := c["account_ids"].([]interface{}); len(v) > 0 {
		fivetranConfig.AccountIDs(xInterfaceStrXStr(v))
	}
	if v := c["sid"].(string); v != "" {
		fivetranConfig.SID(v)
	}
	if v := c["secret"].(string); v != "" {
		fivetranConfig.Secret(v)
	}
	if v := c["oauth_token"].(string); v != "" {
		fivetranConfig.OauthToken(v)
	}
	if v := c["oauth_token_secret"].(string); v != "" {
		fivetranConfig.OauthTokenSecret(v)
	}
	if v := c["consumer_key"].(string); v != "" {
		fivetranConfig.ConsumerKey(v)
	}
	if v := c["consumer_secret"].(string); v != "" {
		fivetranConfig.ConsumerSecret(v)
	}
	if v := c["key"].(string); v != "" {
		fivetranConfig.Key(v)
	}
	if v := c["advertisers_id"].([]interface{}); len(v) > 0 {
		fivetranConfig.AdvertisersID(xInterfaceStrXStr(v))
	}
	if v := c["sync_format"].(string); v != "" {
		fivetranConfig.SyncFormat(v)
	}
	if v := c["bucket_service"].(string); v != "" {
		fivetranConfig.BucketService(v)
	}

	if v := c["report_url"].(string); v != "" {
		fivetranConfig.ReportURL(v)
	}
	if v := c["unique_id"].(string); v != "" {
		fivetranConfig.UniqueID(v)
	}
	if v := c["auth_type"].(string); v != "" {
		fivetranConfig.AuthType(v)
	}
	if v := c["is_new_package"].(string); v != "" {
		fivetranConfig.IsNewPackage(strToBool(v))
	}
	if v := c["adobe_analytics_configurations"].([]interface{}); len(v) > 0 {
		fivetranConfig.AdobeAnalyticsConfigurations(resourceConnectorCreateConfigAdobeAnalyticsConfigurations(v))
	}
	if v := c["is_multi_entity_feature_enabled"].(string); v != "" {
		fivetranConfig.IsMultiEntityFeatureEnabled(strToBool(v))
	}
	if v := c["api_type"].(string); v != "" {
		fivetranConfig.ApiType(v)
	}
	if v := c["base_url"].(string); v != "" {
		fivetranConfig.BaseUrl(v)
	}
	if v := c["entity_id"].(string); v != "" {
		fivetranConfig.EntityId(v)
	}
	if v := c["soap_uri"].(string); v != "" {
		fivetranConfig.SoapUri(v)
	}
	if v := c["user_id"].(string); v != "" {
		fivetranConfig.UserId(v)
	}
	if v := c["encryption_key"].(string); v != "" {
		fivetranConfig.EncryptionKey(v)
	}
	if v := c["always_encrypted"].(string); v != "" {
		fivetranConfig.AlwaysEncrypted(strToBool(v))
	}

	return fivetranConfig
}

func resourceConnectorCreateConfig(fivetranConfig *fivetran.ConnectorConfig, destination_schema []interface{}) *fivetran.ConnectorConfig {
	d := destination_schema[0].(map[string]interface{})
	if v := d["name"].(string); v != "" {
		fivetranConfig.Schema(v)
	}
	if v := d["table"].(string); v != "" {
		fivetranConfig.Table(v)
	}
	if v := d["prefix"].(string); v != "" {
		fivetranConfig.SchemaPrefix(v)
	}

	return fivetranConfig
}

func resourceConnectorCreateConfigProjectCredentials(xi []interface{}) []*fivetran.ConnectorConfigProjectCredentials {
	projectCredentials := make([]*fivetran.ConnectorConfigProjectCredentials, len(xi))
	for i, v := range xi {
		pc := fivetran.NewConnectorConfigProjectCredentials()
		if project, ok := v.(map[string]interface{})["project"].(string); ok && project != "" {
			pc.Project(project)
		}
		if apiKey, ok := v.(map[string]interface{})["api_key"].(string); ok && apiKey != "" {
			pc.APIKey(apiKey)
		}
		if secretKey, ok := v.(map[string]interface{})["secret_key"].(string); ok && secretKey != "" {
			pc.SecretKey(secretKey)
		}
		projectCredentials[i] = pc
	}

	return projectCredentials
}

func resourceConnectorCreateConfigCustomTables(xi []interface{}) []*fivetran.ConnectorConfigCustomTables {
	customTables := make([]*fivetran.ConnectorConfigCustomTables, len(xi))
	for i, v := range xi {
		ct := fivetran.NewConnectorConfigCustomTables()
		if tableName, ok := v.(map[string]interface{})["table_name"].(string); ok && tableName != "" {
			ct.TableName(tableName)
		}
		if configType, ok := v.(map[string]interface{})["config_type"].(string); ok && configType != "" {
			ct.ConfigType(configType)
		}
		if fields, ok := v.(map[string]interface{})["fields"].([]interface{}); ok && len(fields) > 0 {
			ct.Fields(xInterfaceStrXStr(fields))
		}
		if breakdowns, ok := v.(map[string]interface{})["breakdowns"].([]interface{}); ok && len(breakdowns) > 0 {
			ct.Breakdowns(xInterfaceStrXStr(breakdowns))
		}
		if actionBreakdowns, ok := v.(map[string]interface{})["action_breakdowns"].([]interface{}); ok && len(actionBreakdowns) > 0 {
			ct.ActionBreakdowns(xInterfaceStrXStr(actionBreakdowns))
		}
		if aggregation, ok := v.(map[string]interface{})["aggregation"].(string); ok && aggregation != "" {
			ct.Aggregation(aggregation)
		}
		if actionReportTime, ok := v.(map[string]interface{})["action_report_time"].(string); ok && actionReportTime != "" {
			ct.ActionReportTime(actionReportTime)
		}
		if clickAttributionWindow, ok := v.(map[string]interface{})["click_attribution_window"].(string); ok && clickAttributionWindow != "" {
			ct.ClickAttributionWindow(clickAttributionWindow)
		}
		if viewAttributionWindow, ok := v.(map[string]interface{})["view_attribution_window"].(string); ok && viewAttributionWindow != "" {
			ct.ViewAttributionWindow(viewAttributionWindow)
		}
		if prebuiltReportName, ok := v.(map[string]interface{})["prebuilt_report_name"].(string); ok && prebuiltReportName != "" {
			ct.PrebuiltReportName(prebuiltReportName)
		}
		customTables[i] = ct
	}

	return customTables
}

func resourceConnectorCreateConfigAdobeAnalyticsConfigurations(xi []interface{}) []*fivetran.ConnectorConfigAdobeAnalyticsConfiguration {
	configurations := make([]*fivetran.ConnectorConfigAdobeAnalyticsConfiguration, len(xi))
	for i, v := range xi {
		c := fivetran.NewConnectorConfigAdobeAnalyticsConfiguration()

		if syncMode, ok := v.(map[string]interface{})["sync_mode"].(string); ok && syncMode != "" {
			c.SyncMode(syncMode)
		}
		if metrics, ok := v.(map[string]interface{})["metrics"].([]interface{}); ok && len(metrics) > 0 {
			c.Metrics(xInterfaceStrXStr(metrics))
		}
		if reportSuites, ok := v.(map[string]interface{})["report_suites"].([]interface{}); ok && len(reportSuites) > 0 {
			c.ReportSuites(xInterfaceStrXStr(reportSuites))
		}
		if segments, ok := v.(map[string]interface{})["segments"].([]interface{}); ok && len(segments) > 0 {
			c.Segments(xInterfaceStrXStr(segments))
		}
		if elements, ok := v.(map[string]interface{})["elements"].([]interface{}); ok && len(elements) > 0 {
			c.Elements(xInterfaceStrXStr(elements))
		}
		if calculatedMetrics, ok := v.(map[string]interface{})["calculated_metrics"].([]interface{}); ok && len(calculatedMetrics) > 0 {
			c.CalculatedMetrics(xInterfaceStrXStr(calculatedMetrics))
		}

		configurations[i] = c
	}

	return configurations
}

func resourceConnectorCreateConfigReports(xi []interface{}) []*fivetran.ConnectorConfigReports {
	reports := make([]*fivetran.ConnectorConfigReports, len(xi))
	for i, v := range xi {
		r := fivetran.NewConnectorConfigReports()
		if table, ok := v.(map[string]interface{})["table"].(string); ok && table != "" {
			r.Table(table)
		}
		if configType, ok := v.(map[string]interface{})["config_type"].(string); ok && configType != "" {
			r.ConfigType(configType)
		}
		if prebuiltReport, ok := v.(map[string]interface{})["prebuilt_report"].(string); ok && prebuiltReport != "" {
			r.PrebuiltReport(prebuiltReport)
		}
		if reportType, ok := v.(map[string]interface{})["report_type"].(string); ok && reportType != "" {
			r.ReportType(reportType)
		}
		if fields, ok := v.(map[string]interface{})["fields"].([]interface{}); ok && len(fields) > 0 {
			r.Fields(xInterfaceStrXStr(fields))
		}
		if dimensions, ok := v.(map[string]interface{})["dimensions"].([]interface{}); ok && len(dimensions) > 0 {
			r.Dimensions(xInterfaceStrXStr(dimensions))
		}
		if metrics, ok := v.(map[string]interface{})["metrics"].([]interface{}); ok && len(metrics) > 0 {
			r.Metrics(xInterfaceStrXStr(metrics))
		}
		if segments, ok := v.(map[string]interface{})["segments"].([]interface{}); ok && len(segments) > 0 {
			r.Segments(xInterfaceStrXStr(segments))
		}
		if filter, ok := v.(map[string]interface{})["filter"].(string); ok && filter != "" {
			r.Filter(filter)
		}
		reports[i] = r
	}

	return reports
}

func resourceConnectorCreateAuth(auth []interface{}) *fivetran.ConnectorAuth {
	fivetranAuth := fivetran.NewConnectorAuth()

	if len(auth) < 1 {
		return fivetranAuth
	}
	if auth[0] == nil {
		return fivetranAuth
	}

	a := auth[0].(map[string]interface{})

	if v := a["client_access"].([]interface{}); len(v) > 0 {
		fivetranAuth.ClientAccess(resourceConnectorCreateAuthClientAccess(v))
	}
	if v := a["refresh_token"].(string); v != "" {
		fivetranAuth.RefreshToken(v)
	}
	if v := a["access_token"].(string); v != "" {
		fivetranAuth.AccessToken(v)
	}
	if v := a["realm_id"].(string); v != "" {
		fivetranAuth.RealmID(v)
	}

	return fivetranAuth
}

func resourceConnectorCreateAuthClientAccess(clientAccess []interface{}) *fivetran.ConnectorAuthClientAccess {
	fivetranAuthClientAccess := fivetran.NewConnectorAuthClientAccess()

	if len(clientAccess) < 1 {
		return fivetranAuthClientAccess
	}
	if clientAccess[0] == nil {
		return fivetranAuthClientAccess
	}

	ca := clientAccess[0].(map[string]interface{})
	if v := ca["client_id"].(string); v != "" {
		fivetranAuthClientAccess.ClientID(v)
	}
	if v := ca["client_secret"].(string); v != "" {
		fivetranAuthClientAccess.ClientSecret(v)
	}
	if v := ca["user_agent"].(string); v != "" {
		fivetranAuthClientAccess.UserAgent(v)
	}
	if v := ca["developer_token"].(string); v != "" {
		fivetranAuthClientAccess.DeveloperToken(v)
	}

	return fivetranAuthClientAccess
}

func resourceConnectorReadDestinationSchema(schema string, service string) []interface{} {
	return readDestinationSchema(schema, service)
}

// resourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func resourceConnectorReadStatus(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	status := make([]interface{}, 1)

	s := make(map[string]interface{})
	mapAddStr(s, "setup_state", resp.Data.Status.SetupState)
	mapAddStr(s, "sync_state", resp.Data.Status.SyncState)
	mapAddStr(s, "update_state", resp.Data.Status.UpdateState)
	mapAddStr(s, "is_historical_sync", boolPointerToStr(resp.Data.Status.IsHistoricalSync))
	mapAddXInterface(s, "tasks", resourceConnectorReadStatusFlattenTasks(resp))
	mapAddXInterface(s, "warnings", resourceConnectorReadStatusFlattenWarnings(resp))
	status[0] = s

	return status
}

func resourceConnectorReadStatusFlattenTasks(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Status.Tasks) < 1 {
		return make([]interface{}, 0)
	}

	tasks := make([]interface{}, len(resp.Data.Status.Tasks))
	for i, v := range resp.Data.Status.Tasks {
		task := make(map[string]interface{})
		mapAddStr(task, "code", v.Code)
		mapAddStr(task, "message", v.Message)

		tasks[i] = task
	}

	return tasks
}

func resourceConnectorReadStatusFlattenWarnings(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Status.Warnings) < 1 {
		return make([]interface{}, 0)
	}

	warnings := make([]interface{}, len(resp.Data.Status.Warnings))
	for i, v := range resp.Data.Status.Warnings {
		warning := make(map[string]interface{})
		mapAddStr(warning, "code", v.Code)
		mapAddStr(warning, "message", v.Message)

		warnings[i] = warning
	}

	return warnings
}

// dataSourceConnectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func resourceConnectorReadConfig(resp *fivetran.ConnectorDetailsResponse, currentConfig []interface{}) []interface{} {
	config := make([]interface{}, 1)

	c := make(map[string]interface{})

	// get sensitive fields from the currentConfig to prevent drifting (Fivetran returns this values masked)
	if len(currentConfig) > 0 {
		mapAddStr(c, "password", currentConfig[0].(map[string]interface{})["password"].(string))
		mapAddStr(c, "client_secret", currentConfig[0].(map[string]interface{})["client_secret"].(string))
		mapAddStr(c, "private_key", currentConfig[0].(map[string]interface{})["private_key"].(string))
		mapAddStr(c, "s3role_arn", currentConfig[0].(map[string]interface{})["s3role_arn"].(string))
		mapAddStr(c, "ftp_password", currentConfig[0].(map[string]interface{})["ftp_password"].(string))
		mapAddStr(c, "sftp_password", currentConfig[0].(map[string]interface{})["sftp_password"].(string))
		mapAddStr(c, "api_key", currentConfig[0].(map[string]interface{})["api_key"].(string))
		mapAddStr(c, "role_arn", currentConfig[0].(map[string]interface{})["role_arn"].(string))
		mapAddStr(c, "secret_key", currentConfig[0].(map[string]interface{})["secret_key"].(string))
		mapAddStr(c, "pem_certificate", currentConfig[0].(map[string]interface{})["pem_certificate"].(string))
		mapAddStr(c, "access_token", currentConfig[0].(map[string]interface{})["access_token"].(string))
		mapAddStr(c, "api_secret", currentConfig[0].(map[string]interface{})["api_secret"].(string))
		mapAddStr(c, "api_access_token", currentConfig[0].(map[string]interface{})["api_access_token"].(string))
		mapAddStr(c, "secret", currentConfig[0].(map[string]interface{})["secret"].(string))
		mapAddStr(c, "consumer_secret", currentConfig[0].(map[string]interface{})["consumer_secret"].(string))
		mapAddStr(c, "secrets", currentConfig[0].(map[string]interface{})["secrets"].(string))
		mapAddStr(c, "api_token", currentConfig[0].(map[string]interface{})["api_token"].(string))
		mapAddStr(c, "consumer_key", currentConfig[0].(map[string]interface{})["consumer_key"].(string))
		mapAddStr(c, "encryption_key", currentConfig[0].(map[string]interface{})["encryption_key"].(string))
		mapAddStr(c, "oauth_token", currentConfig[0].(map[string]interface{})["oauth_token"].(string))
		mapAddStr(c, "oauth_token_secret", currentConfig[0].(map[string]interface{})["oauth_token_secret"].(string))
	}

	// Collections
	mapAddXInterface(c, "report_suites", xStrXInterface(resp.Data.Config.ReportSuites))
	mapAddXInterface(c, "elements", xStrXInterface(resp.Data.Config.Elements))
	mapAddXInterface(c, "metrics", xStrXInterface(resp.Data.Config.Metrics))
	mapAddXInterface(c, "advertisables", xStrXInterface(resp.Data.Config.Advertisables))
	mapAddXInterface(c, "dimensions", xStrXInterface(resp.Data.Config.Dimensions))
	mapAddXInterface(c, "project_credentials", resourceConnectorReadConfigFlattenProjectCredentials(resp, currentConfig))
	mapAddXInterface(c, "selected_exports", xStrXInterface(resp.Data.Config.SelectedExports))
	mapAddXInterface(c, "apps", xStrXInterface(resp.Data.Config.Apps))
	mapAddXInterface(c, "sales_accounts", xStrXInterface(resp.Data.Config.SalesAccounts))
	mapAddXInterface(c, "finance_accounts", xStrXInterface(resp.Data.Config.FinanceAccounts))
	mapAddXInterface(c, "projects", xStrXInterface(resp.Data.Config.Projects))
	mapAddXInterface(c, "user_profiles", xStrXInterface(resp.Data.Config.UserProfiles))
	mapAddXInterface(c, "report_configuration_ids", xStrXInterface(resp.Data.Config.ReportConfigurationIDs))
	mapAddXInterface(c, "custom_tables", resourceConnectorReadConfigFlattenCustomTables(resp))
	mapAddXInterface(c, "pages", xStrXInterface(resp.Data.Config.Pages))
	mapAddXInterface(c, "accounts", xStrXInterface(resp.Data.Config.Accounts))
	mapAddXInterface(c, "fields", xStrXInterface(resp.Data.Config.Fields))
	mapAddXInterface(c, "breakdowns", xStrXInterface(resp.Data.Config.Breakdowns))
	mapAddXInterface(c, "action_breakdowns", xStrXInterface(resp.Data.Config.ActionBreakdowns))
	mapAddXInterface(c, "repositories", xStrXInterface(resp.Data.Config.Repositories))
	mapAddXInterface(c, "dimension_attributes", xStrXInterface(resp.Data.Config.DimensionAttributes))
	mapAddXInterface(c, "columns", xStrXInterface(resp.Data.Config.Columns))
	mapAddXInterface(c, "manager_accounts", xStrXInterface(resp.Data.Config.ManagerAccounts))
	mapAddXInterface(c, "reports", resourceConnectorReadConfigFlattenReports(resp))
	mapAddXInterface(c, "site_urls", xStrXInterface(resp.Data.Config.SiteURLs))
	mapAddXInterface(c, "profiles", xStrXInterface(resp.Data.Config.Profiles))
	mapAddXInterface(c, "hosts", xStrXInterface(resp.Data.Config.Hosts))
	mapAddXInterface(c, "api_keys", xStrXInterface(resp.Data.Config.APIKeys))
	mapAddXInterface(c, "adobe_analytics_configurations", resourceConnectorReadConfigFlattenAdobeAnalyticsConfigurations(resp))
	mapAddXInterface(c, "advertisers", xStrXInterface(resp.Data.Config.Advertisers))
	mapAddXInterface(c, "organizations", xStrXInterface(resp.Data.Config.Organizations))
	mapAddXInterface(c, "account_ids", xStrXInterface(resp.Data.Config.AccountIDs))
	mapAddXInterface(c, "advertisers_id", xStrXInterface(resp.Data.Config.AdvertisersID))

	// Boolean fields
	mapAddStr(c, "is_ftps", boolPointerToStr(resp.Data.Config.IsFTPS))
	mapAddStr(c, "sftp_is_key_pair", boolPointerToStr(resp.Data.Config.SFTPIsKeyPair))
	mapAddStr(c, "sync_data_locker", boolPointerToStr(resp.Data.Config.SyncDataLocker))
	mapAddStr(c, "enable_all_dimension_combinations", boolPointerToStr(resp.Data.Config.EnableAllDimensionCombinations))
	mapAddStr(c, "use_webhooks", boolPointerToStr(resp.Data.Config.UseWebhooks))
	mapAddStr(c, "update_config_on_each_sync", boolPointerToStr(resp.Data.Config.UpdateConfigOnEachSync))
	mapAddStr(c, "on_premise", boolPointerToStr(resp.Data.Config.OnPremise))
	mapAddStr(c, "always_encrypted", boolPointerToStr(resp.Data.Config.AlwaysEncrypted))
	mapAddStr(c, "is_new_package", boolPointerToStr(resp.Data.Config.IsNewPackage))
	mapAddStr(c, "is_multi_entity_feature_enabled", boolPointerToStr(resp.Data.Config.IsMultiEntityFeatureEnabled))

	// Integer fields
	mapAddStr(c, "ftp_port", intPointerToStr(resp.Data.Config.FTPPort))
	mapAddStr(c, "sftp_port", intPointerToStr(resp.Data.Config.SFTPPort))
	mapAddStr(c, "conversion_window_size", intPointerToStr(resp.Data.Config.ConversionWindowSize))
	mapAddStr(c, "port", intPointerToStr(resp.Data.Config.Port))
	mapAddStr(c, "api_quota", intPointerToStr(resp.Data.Config.APIQuota))
	mapAddStr(c, "tunnel_port", intPointerToStr(resp.Data.Config.TunnelPort))
	mapAddStr(c, "daily_api_call_limit", intPointerToStr(resp.Data.Config.DailyAPICallLimit))
	mapAddStr(c, "skip_before", intPointerToStr(&resp.Data.Config.SkipBefore))
	mapAddStr(c, "skip_after", intPointerToStr(&resp.Data.Config.SkipAfter))

	// String fields
	mapAddStr(c, "sheet_id", resp.Data.Config.SheetID)
	mapAddStr(c, "named_range", resp.Data.Config.NamedRange)
	mapAddStr(c, "client_id", resp.Data.Config.ClientID)
	mapAddStr(c, "technical_account_id", resp.Data.Config.TechnicalAccountID)
	mapAddStr(c, "organization_id", resp.Data.Config.OrganizationID)
	mapAddStr(c, "sync_mode", resp.Data.Config.SyncMode)
	mapAddStr(c, "date_granularity", resp.Data.Config.DateGranularity)
	mapAddStr(c, "timeframe_months", resp.Data.Config.TimeframeMonths)
	mapAddStr(c, "source", resp.Data.Config.Source)
	mapAddStr(c, "s3bucket", resp.Data.Config.S3Bucket)
	mapAddStr(c, "abs_connection_string", resp.Data.Config.ABSConnectionString)
	mapAddStr(c, "abs_container_name", resp.Data.Config.ABSContainerName)
	mapAddStr(c, "folder_id", resp.Data.Config.FolderId)
	mapAddStr(c, "ftp_host", resp.Data.Config.FTPHost)
	mapAddStr(c, "ftp_user", resp.Data.Config.FTPUser)
	mapAddStr(c, "sftp_host", resp.Data.Config.SFTPHost)
	mapAddStr(c, "sftp_user", resp.Data.Config.SFTPUser)
	mapAddStr(c, "report_type", resp.Data.Config.ReportType)
	mapAddStr(c, "external_id", resp.Data.Config.ExternalID)
	mapAddStr(c, "bucket", resp.Data.Config.Bucket)
	mapAddStr(c, "prefix", resp.Data.Config.Prefix)
	mapAddStr(c, "pattern", resp.Data.Config.Pattern)
	mapAddStr(c, "file_type", resp.Data.Config.FileType)
	mapAddStr(c, "compression", resp.Data.Config.Compression)
	mapAddStr(c, "on_error", resp.Data.Config.OnError)
	mapAddStr(c, "append_file_option", resp.Data.Config.AppendFileOption)
	mapAddStr(c, "archive_pattern", resp.Data.Config.ArchivePattern)
	mapAddStr(c, "null_sequence", resp.Data.Config.NullSequence)
	mapAddStr(c, "delimiter", resp.Data.Config.Delimiter)
	mapAddStr(c, "escape_char", resp.Data.Config.EscapeChar)
	mapAddStr(c, "auth_mode", resp.Data.Config.AuthMode)
	mapAddStr(c, "user_name", resp.Data.Config.UserName)
	mapAddStr(c, "username", resp.Data.Config.Username)
	mapAddStr(c, "certificate", resp.Data.Config.Certificate)
	mapAddStr(c, "consumer_group", resp.Data.Config.ConsumerGroup)
	mapAddStr(c, "servers", resp.Data.Config.Servers)
	mapAddStr(c, "message_type", resp.Data.Config.MessageType)
	mapAddStr(c, "sync_type", resp.Data.Config.SyncType)
	mapAddStr(c, "security_protocol", resp.Data.Config.SecurityProtocol)
	mapAddStr(c, "app_sync_mode", resp.Data.Config.AppSyncMode)
	mapAddStr(c, "sales_account_sync_mode", resp.Data.Config.SalesAccountSyncMode)
	mapAddStr(c, "finance_account_sync_mode", resp.Data.Config.FinanceAccountSyncMode)
	mapAddStr(c, "access_key_id", resp.Data.Config.AccessKeyID)
	mapAddStr(c, "home_folder", resp.Data.Config.HomeFolder)
	mapAddStr(c, "function", resp.Data.Config.Function)
	mapAddStr(c, "region", resp.Data.Config.Region)
	mapAddStr(c, "container_name", resp.Data.Config.ContainerName)
	mapAddStr(c, "connection_string", resp.Data.Config.ConnectionString)
	mapAddStr(c, "connection_type", resp.Data.Config.ConnectionType)
	mapAddStr(c, "function_app", resp.Data.Config.FunctionApp)
	mapAddStr(c, "function_name", resp.Data.Config.FunctionName)
	mapAddStr(c, "function_key", resp.Data.Config.FunctionKey)
	mapAddStr(c, "public_key", resp.Data.Config.PublicKey)
	mapAddStr(c, "merchant_id", resp.Data.Config.MerchantID)
	mapAddStr(c, "api_url", resp.Data.Config.APIURL)
	mapAddStr(c, "cloud_storage_type", resp.Data.Config.CloudStorageType)
	mapAddStr(c, "s3external_id", resp.Data.Config.S3ExternalID)
	mapAddStr(c, "s3folder", resp.Data.Config.S3Folder)
	mapAddStr(c, "gcs_bucket", resp.Data.Config.GCSBucket)
	mapAddStr(c, "gcs_folder", resp.Data.Config.GCSFolder)
	mapAddStr(c, "instance", resp.Data.Config.Instance)
	mapAddStr(c, "aws_region_code", resp.Data.Config.AWSRegionCode)
	mapAddStr(c, "aggregation", resp.Data.Config.Aggregation)
	mapAddStr(c, "config_type", resp.Data.Config.ConfigType)
	mapAddStr(c, "prebuilt_report", resp.Data.Config.PrebuiltReport)
	mapAddStr(c, "action_report_time", resp.Data.Config.ActionReportTime)
	mapAddStr(c, "click_attribution_window", resp.Data.Config.ClickAttributionWindow)
	mapAddStr(c, "view_attribution_window", resp.Data.Config.ViewAttributionWindow)
	mapAddStr(c, "subdomain", resp.Data.Config.Subdomain)
	mapAddStr(c, "host", resp.Data.Config.Host)
	mapAddStr(c, "user", resp.Data.Config.User)
	mapAddStr(c, "is_secure", resp.Data.Config.IsSecure)
	mapAddStr(c, "network_code", resp.Data.Config.NetworkCode)
	mapAddStr(c, "customer_id", resp.Data.Config.CustomerID)
	mapAddStr(c, "project_id", resp.Data.Config.ProjectID)
	mapAddStr(c, "dataset_id", resp.Data.Config.DatasetID)
	mapAddStr(c, "bucket_name", resp.Data.Config.BucketName)
	mapAddStr(c, "function_trigger", resp.Data.Config.FunctionTrigger)
	mapAddStr(c, "config_method", resp.Data.Config.ConfigMethod)
	mapAddStr(c, "query_id", resp.Data.Config.QueryID)
	mapAddStr(c, "path", resp.Data.Config.Path)
	mapAddStr(c, "view_through_attribution_window_size", resp.Data.Config.ViewThroughAttributionWindowSize)
	mapAddStr(c, "post_click_attribution_window_size", resp.Data.Config.PostClickAttributionWindowSize)
	mapAddStr(c, "use_api_keys", resp.Data.Config.UseAPIKeys)
	mapAddStr(c, "endpoint", resp.Data.Config.Endpoint)
	mapAddStr(c, "identity", resp.Data.Config.Identity)
	mapAddStr(c, "tunnel_host", resp.Data.Config.TunnelHost)
	mapAddStr(c, "domain_name", resp.Data.Config.DomainName)
	mapAddStr(c, "resource_url", resp.Data.Config.ResourceURL)
	mapAddStr(c, "tunnel_user", resp.Data.Config.TunnelUser)
	mapAddStr(c, "database", resp.Data.Config.Database)
	mapAddStr(c, "datasource", resp.Data.Config.Datasource)
	mapAddStr(c, "account", resp.Data.Config.Account)
	mapAddStr(c, "role", resp.Data.Config.Role)
	mapAddStr(c, "email", resp.Data.Config.Email)
	mapAddStr(c, "account_id", resp.Data.Config.AccountID)
	mapAddStr(c, "server_url", resp.Data.Config.ServerURL)
	mapAddStr(c, "user_key", resp.Data.Config.UserKey)
	mapAddStr(c, "api_version", resp.Data.Config.APIVersion)
	mapAddStr(c, "time_zone", resp.Data.Config.TimeZone)
	mapAddStr(c, "integration_key", resp.Data.Config.IntegrationKey)
	mapAddStr(c, "engagement_attribution_window", resp.Data.Config.EngagementAttributionWindow)
	mapAddStr(c, "conversion_report_time", resp.Data.Config.ConversionReportTime)
	mapAddStr(c, "domain", resp.Data.Config.Domain)
	mapAddStr(c, "update_method", resp.Data.Config.UpdateMethod)
	mapAddStr(c, "replication_slot", resp.Data.Config.ReplicationSlot)
	mapAddStr(c, "publication_name", resp.Data.Config.PublicationName)
	mapAddStr(c, "data_center", resp.Data.Config.DataCenter)
	mapAddStr(c, "sub_domain", resp.Data.Config.SubDomain)
	mapAddStr(c, "test_table_name", resp.Data.Config.TestTableName)
	mapAddStr(c, "shop", resp.Data.Config.Shop)
	mapAddStr(c, "swipe_attribution_window", resp.Data.Config.SwipeAttributionWindow)
	mapAddStr(c, "sid", resp.Data.Config.SID)
	mapAddStr(c, "key", resp.Data.Config.Key)
	mapAddStr(c, "sync_format", resp.Data.Config.SyncFormat)
	mapAddStr(c, "bucket_service", resp.Data.Config.BucketService)
	mapAddStr(c, "report_url", resp.Data.Config.ReportURL)
	mapAddStr(c, "unique_id", resp.Data.Config.UniqueID)
	mapAddStr(c, "auth_type", resp.Data.Config.AuthType)
	mapAddStr(c, "latest_version", resp.Data.Config.LatestVersion)
	mapAddStr(c, "authorization_method", resp.Data.Config.AuthorizationMethod)
	mapAddStr(c, "service_version", resp.Data.Config.ServiceVersion)
	mapAddStr(c, "last_synced_changes__utc_", resp.Data.Config.LastSyncedChangesUtc)
	mapAddStr(c, "api_type", resp.Data.Config.ApiType)
	mapAddStr(c, "base_url", resp.Data.Config.BaseUrl)
	mapAddStr(c, "entity_id", resp.Data.Config.EntityId)
	mapAddStr(c, "soap_uri", resp.Data.Config.SoapUri)
	mapAddStr(c, "user_id", resp.Data.Config.UserId)

	config[0] = c

	return config
}

func resourceConnectorReadConfigFlattenProjectCredentials(resp *fivetran.ConnectorDetailsResponse, currentConfig []interface{}) []interface{} {
	if len(resp.Data.Config.ProjectCredentials) < 1 {
		return make([]interface{}, 0)
	}

	projectCredentials := make([]interface{}, len(resp.Data.Config.ProjectCredentials))
	for i, v := range resp.Data.Config.ProjectCredentials {
		pc := make(map[string]interface{})
		mapAddStr(pc, "project", v.Project)
		// The REST API sends the fields "api_key" and "secret_key" masked. We use the state stored config here.
		mapAddStr(pc, "api_key", resourceConnectorReadConfigFlattenProjectCredentialsGetStateValue(v.Project, "api_key", currentConfig))
		mapAddStr(pc, "secret_key", resourceConnectorReadConfigFlattenProjectCredentialsGetStateValue(v.Project, "secret_key", currentConfig))
		projectCredentials[i] = pc
	}

	return projectCredentials
}

func resourceConnectorReadConfigFlattenProjectCredentialsGetStateValue(project, key string, currentConfig []interface{}) string {
	projectCredentials := currentConfig[0].(map[string]interface{})["project_credentials"].([]interface{})
	for _, v := range projectCredentials {
		if v.(map[string]interface{})["project"].(string) == project {
			return v.(map[string]interface{})[key].(string)
		}
	}

	return ""
}

func resourceConnectorReadConfigFlattenReports(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Config.Reports) < 1 {
		return make([]interface{}, 0)
	}

	reports := make([]interface{}, len(resp.Data.Config.Reports))
	for i, v := range resp.Data.Config.Reports {
		r := make(map[string]interface{})
		mapAddStr(r, "table", v.Table)
		mapAddStr(r, "config_type", v.ConfigType)
		mapAddStr(r, "prebuilt_report", v.PrebuiltReport)
		mapAddStr(r, "report_type", v.ReportType)
		mapAddXInterface(r, "fields", xStrXInterface(v.Fields))
		mapAddXInterface(r, "dimensions", xStrXInterface(v.Dimensions))
		mapAddXInterface(r, "metrics", xStrXInterface(v.Metrics))
		mapAddXInterface(r, "segments", xStrXInterface(v.Segments))
		mapAddStr(r, "filter", v.Filter)
		reports[i] = r
	}

	return reports
}

func resourceConnectorReadConfigFlattenAdobeAnalyticsConfigurations(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Config.AdobeAnalyticsConfigurations) < 1 {
		return make([]interface{}, 0)
	}

	configurations := make([]interface{}, len(resp.Data.Config.AdobeAnalyticsConfigurations))
	for i, v := range resp.Data.Config.AdobeAnalyticsConfigurations {
		c := make(map[string]interface{})
		mapAddStr(c, "sync_mode", v.SyncMode)
		mapAddXInterface(c, "metrics", xStrXInterface(v.Metrics))
		mapAddXInterface(c, "calculated_metrics", xStrXInterface(v.CalculatedMetrics))
		mapAddXInterface(c, "elements", xStrXInterface(v.Elements))
		mapAddXInterface(c, "segments", xStrXInterface(v.Segments))
		mapAddXInterface(c, "report_suites", xStrXInterface(v.ReportSuites))
		configurations[i] = c
	}

	return configurations
}

func resourceConnectorReadConfigFlattenCustomTables(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Config.CustomTables) < 1 {
		return make([]interface{}, 0)
	}

	customTables := make([]interface{}, len(resp.Data.Config.CustomTables))
	for i, v := range resp.Data.Config.CustomTables {
		ct := make(map[string]interface{})
		mapAddStr(ct, "table_name", v.TableName)
		mapAddStr(ct, "config_type", v.ConfigType)
		mapAddXInterface(ct, "fields", xStrXInterface(v.Fields))
		mapAddXInterface(ct, "breakdowns", xStrXInterface(v.Breakdowns))
		mapAddXInterface(ct, "action_breakdowns", xStrXInterface(v.ActionBreakdowns))
		mapAddStr(ct, "aggregation", v.Aggregation)
		mapAddStr(ct, "action_report_time", v.ActionReportTime)
		mapAddStr(ct, "click_attribution_window", v.ClickAttributionWindow)
		mapAddStr(ct, "view_attribution_window", v.ViewAttributionWindow)
		mapAddStr(ct, "prebuilt_report_name", v.PrebuiltReportName)
		customTables[i] = ct
	}

	return customTables
}
