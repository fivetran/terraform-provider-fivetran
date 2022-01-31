package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectorRead,
		Schema: map[string]*schema.Schema{
			"id":                {Type: schema.TypeString, Required: true},
			"group_id":          {Type: schema.TypeString, Computed: true},
			"service":           {Type: schema.TypeString, Computed: true},
			"service_version":   {Type: schema.TypeString, Computed: true},
			"schema":            {Type: schema.TypeString, Computed: true},
			"connected_by":      {Type: schema.TypeString, Computed: true},
			"created_at":        {Type: schema.TypeString, Computed: true},
			"succeeded_at":      {Type: schema.TypeString, Computed: true},
			"failed_at":         {Type: schema.TypeString, Computed: true},
			"sync_frequency":    {Type: schema.TypeString, Computed: true},
			"schedule_type":     {Type: schema.TypeString, Computed: true},
			"paused":            {Type: schema.TypeString, Computed: true},
			"pause_after_trial": {Type: schema.TypeString, Computed: true},
			"status":            dataSourceConnectorSchemaStatus(),
			"config":            dataSourceConnectorSchemaConfig(),
		},
	}
}

func dataSourceConnectorSchemaStatus() *schema.Schema {
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

func dataSourceConnectorSchemaConfig() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"schema":                {Type: schema.TypeString, Computed: true},
				"table":                 {Type: schema.TypeString, Computed: true},
				"sheet_id":              {Type: schema.TypeString, Computed: true},
				"named_range":           {Type: schema.TypeString, Computed: true},
				"client_id":             {Type: schema.TypeString, Computed: true},
				"client_secret":         {Type: schema.TypeString, Computed: true},
				"technical_account_id":  {Type: schema.TypeString, Computed: true},
				"organization_id":       {Type: schema.TypeString, Computed: true},
				"private_key":           {Type: schema.TypeString, Computed: true},
				"sync_mode":             {Type: schema.TypeString, Computed: true},
				"report_suites":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"elements":              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"metrics":               {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"date_granularity":      {Type: schema.TypeString, Computed: true},
				"timeframe_months":      {Type: schema.TypeString, Computed: true},
				"source":                {Type: schema.TypeString, Computed: true},
				"s3bucket":              {Type: schema.TypeString, Computed: true},
				"s3role_arn":            {Type: schema.TypeString, Computed: true},
				"abs_connection_string": {Type: schema.TypeString, Computed: true},
				"abs_container_name":    {Type: schema.TypeString, Computed: true},
				"ftp_host":              {Type: schema.TypeString, Computed: true},
				"ftp_port":              {Type: schema.TypeString, Computed: true},
				"ftp_user":              {Type: schema.TypeString, Computed: true},
				"ftp_password":          {Type: schema.TypeString, Computed: true},
				"is_ftps":               {Type: schema.TypeString, Computed: true},
				"sftp_host":             {Type: schema.TypeString, Computed: true},
				"sftp_port":             {Type: schema.TypeString, Computed: true},
				"sftp_user":             {Type: schema.TypeString, Computed: true},
				"sftp_password":         {Type: schema.TypeString, Computed: true},
				"sftp_is_key_pair":      {Type: schema.TypeString, Computed: true},
				"advertisables":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"report_type":           {Type: schema.TypeString, Computed: true},
				"dimensions":            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"schema_prefix":         {Type: schema.TypeString, Computed: true},
				"api_key":               {Type: schema.TypeString, Computed: true},
				"external_id":           {Type: schema.TypeString, Computed: true},
				"role_arn":              {Type: schema.TypeString, Computed: true},
				"bucket":                {Type: schema.TypeString, Computed: true},
				"prefix":                {Type: schema.TypeString, Computed: true},
				"pattern":               {Type: schema.TypeString, Computed: true},
				"file_type":             {Type: schema.TypeString, Computed: true},
				"compression":           {Type: schema.TypeString, Computed: true},
				"on_error":              {Type: schema.TypeString, Computed: true},
				"append_file_option":    {Type: schema.TypeString, Computed: true},
				"archive_pattern":       {Type: schema.TypeString, Computed: true},
				"null_sequence":         {Type: schema.TypeString, Computed: true},
				"delimiter":             {Type: schema.TypeString, Computed: true},
				"escape_char":           {Type: schema.TypeString, Computed: true},
				"skip_before":           {Type: schema.TypeString, Computed: true},
				"skip_after":            {Type: schema.TypeString, Computed: true},
				"project_credentials": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"project":    {Type: schema.TypeString, Computed: true},
							"api_key":    {Type: schema.TypeString, Computed: true, Sensitive: true},
							"secret_key": {Type: schema.TypeString, Computed: true, Sensitive: true},
						},
					},
				},
				"auth_mode":                         {Type: schema.TypeString, Computed: true},
				"username":                          {Type: schema.TypeString, Computed: true},
				"password":                          {Type: schema.TypeString, Computed: true},
				"certificate":                       {Type: schema.TypeString, Computed: true},
				"selected_exports":                  {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"consumer_group":                    {Type: schema.TypeString, Computed: true},
				"servers":                           {Type: schema.TypeString, Computed: true},
				"message_type":                      {Type: schema.TypeString, Computed: true},
				"sync_type":                         {Type: schema.TypeString, Computed: true},
				"security_protocol":                 {Type: schema.TypeString, Computed: true},
				"apps":                              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"sales_accounts":                    {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"finance_accounts":                  {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"app_sync_mode":                     {Type: schema.TypeString, Computed: true},
				"sales_account_sync_mode":           {Type: schema.TypeString, Computed: true},
				"finance_account_sync_mode":         {Type: schema.TypeString, Computed: true},
				"pem_certificate":                   {Type: schema.TypeString, Computed: true},
				"access_key_id":                     {Type: schema.TypeString, Computed: true},
				"secret_key":                        {Type: schema.TypeString, Computed: true},
				"home_folder":                       {Type: schema.TypeString, Computed: true},
				"sync_data_locker":                  {Type: schema.TypeString, Computed: true},
				"projects":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"function":                          {Type: schema.TypeString, Computed: true},
				"region":                            {Type: schema.TypeString, Computed: true},
				"secrets":                           {Type: schema.TypeString, Computed: true},
				"container_name":                    {Type: schema.TypeString, Computed: true},
				"connection_string":                 {Type: schema.TypeString, Computed: true},
				"connection_type":                   {Type: schema.TypeString, Computed: true},
				"function_app":                      {Type: schema.TypeString, Computed: true},
				"function_name":                     {Type: schema.TypeString, Computed: true},
				"function_key":                      {Type: schema.TypeString, Computed: true},
				"public_key":                        {Type: schema.TypeString, Computed: true},
				"merchant_id":                       {Type: schema.TypeString, Computed: true},
				"api_url":                           {Type: schema.TypeString, Computed: true},
				"cloud_storage_type":                {Type: schema.TypeString, Computed: true},
				"s3external_id":                     {Type: schema.TypeString, Computed: true},
				"s3folder":                          {Type: schema.TypeString, Computed: true},
				"gcs_bucket":                        {Type: schema.TypeString, Computed: true},
				"gcs_folder":                        {Type: schema.TypeString, Computed: true},
				"user_profiles":                     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"report_configuration_ids":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"enable_all_dimension_combinations": {Type: schema.TypeString, Computed: true},
				"instance":                          {Type: schema.TypeString, Computed: true},
				"aws_region_code":                   {Type: schema.TypeString, Computed: true},
				"accounts":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"fields":                            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"breakdowns":                        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"action_breakdowns":                 {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"aggregation":                       {Type: schema.TypeString, Computed: true},
				"config_type":                       {Type: schema.TypeString, Computed: true},
				"prebuilt_report":                   {Type: schema.TypeString, Computed: true},
				"action_report_time":                {Type: schema.TypeString, Computed: true},
				"click_attribution_window":          {Type: schema.TypeString, Computed: true},
				"view_attribution_window":           {Type: schema.TypeString, Computed: true},
				"custom_tables": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"table_name":               {Type: schema.TypeString, Computed: true},
							"config_type":              {Type: schema.TypeString, Computed: true},
							"fields":                   {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"breakdowns":               {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"action_breakdowns":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"aggregation":              {Type: schema.TypeString, Computed: true},
							"action_report_time":       {Type: schema.TypeString, Computed: true},
							"click_attribution_window": {Type: schema.TypeString, Computed: true},
							"view_attribution_window":  {Type: schema.TypeString, Computed: true},
							"prebuilt_report_name":     {Type: schema.TypeString, Computed: true},
						},
					},
				},
				"pages":                {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"subdomain":            {Type: schema.TypeString, Computed: true},
				"host":                 {Type: schema.TypeString, Computed: true},
				"port":                 {Type: schema.TypeString, Computed: true},
				"user":                 {Type: schema.TypeString, Computed: true},
				"is_secure":            {Type: schema.TypeString, Computed: true},
				"repositories":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"use_webhooks":         {Type: schema.TypeString, Computed: true},
				"dimension_attributes": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"columns":              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"network_code":         {Type: schema.TypeString, Computed: true},
				"customer_id":          {Type: schema.TypeString, Computed: true},
				"manager_accounts":     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"reports": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"table":           {Type: schema.TypeString, Computed: true},
							"config_type":     {Type: schema.TypeString, Computed: true},
							"prebuilt_report": {Type: schema.TypeString, Computed: true},
							"report_type":     {Type: schema.TypeString, Computed: true},
							"fields":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"dimensions":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"metrics":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"segments":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"filter":          {Type: schema.TypeString, Computed: true},
						},
					},
				},
				"conversion_window_size":               {Type: schema.TypeString, Computed: true},
				"profiles":                             {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"project_id":                           {Type: schema.TypeString, Computed: true},
				"dataset_id":                           {Type: schema.TypeString, Computed: true},
				"bucket_name":                          {Type: schema.TypeString, Computed: true},
				"function_trigger":                     {Type: schema.TypeString, Computed: true},
				"config_method":                        {Type: schema.TypeString, Computed: true},
				"query_id":                             {Type: schema.TypeString, Computed: true},
				"update_config_on_each_sync":           {Type: schema.TypeString, Computed: true},
				"site_urls":                            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"path":                                 {Type: schema.TypeString, Computed: true},
				"on_premise":                           {Type: schema.TypeString, Computed: true},
				"access_token":                         {Type: schema.TypeString, Computed: true},
				"view_through_attribution_window_size": {Type: schema.TypeString, Computed: true},
				"post_click_attribution_window_size":   {Type: schema.TypeString, Computed: true},
				"use_api_keys":                         {Type: schema.TypeString, Computed: true},
				"api_keys":                             {Type: schema.TypeString, Computed: true},
				"endpoint":                             {Type: schema.TypeString, Computed: true},
				"identity":                             {Type: schema.TypeString, Computed: true},
				"api_quota":                            {Type: schema.TypeString, Computed: true},
				"domain_name":                          {Type: schema.TypeString, Computed: true},
				"resource_url":                         {Type: schema.TypeString, Computed: true},
				"api_secret":                           {Type: schema.TypeString, Computed: true},
				"hosts":                                {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"tunnel_host":                          {Type: schema.TypeString, Computed: true},
				"tunnel_port":                          {Type: schema.TypeString, Computed: true},
				"tunnel_user":                          {Type: schema.TypeString, Computed: true},
				"database":                             {Type: schema.TypeString, Computed: true},
				"datasource":                           {Type: schema.TypeString, Computed: true},
				"account":                              {Type: schema.TypeString, Computed: true},
				"role":                                 {Type: schema.TypeString, Computed: true},
				"email":                                {Type: schema.TypeString, Computed: true},
				"account_id":                           {Type: schema.TypeString, Computed: true},
				"server_url":                           {Type: schema.TypeString, Computed: true},
				"user_key":                             {Type: schema.TypeString, Computed: true},
				"api_version":                          {Type: schema.TypeString, Computed: true},
				"daily_api_call_limit":                 {Type: schema.TypeString, Computed: true},
				"time_zone":                            {Type: schema.TypeString, Computed: true},
				"integration_key":                      {Type: schema.TypeString, Computed: true},
				"advertisers":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"engagement_attribution_window":        {Type: schema.TypeString, Computed: true},
				"conversion_report_time":               {Type: schema.TypeString, Computed: true},
				"domain":                               {Type: schema.TypeString, Computed: true},
				"update_method":                        {Type: schema.TypeString, Computed: true},
				"replication_slot":                     {Type: schema.TypeString, Computed: true},
				"data_center":                          {Type: schema.TypeString, Computed: true},
				"api_token":                            {Type: schema.TypeString, Computed: true},
				"sub_domain":                           {Type: schema.TypeString, Computed: true},
				"test_table_name":                      {Type: schema.TypeString, Computed: true},
				"shop":                                 {Type: schema.TypeString, Computed: true},
				"organizations":                        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"swipe_attribution_window":             {Type: schema.TypeString, Computed: true},
				"api_access_token":                     {Type: schema.TypeString, Computed: true},
				"account_ids":                          {Type: schema.TypeString, Computed: true},
				"sid":                                  {Type: schema.TypeString, Computed: true},
				"secret":                               {Type: schema.TypeString, Computed: true},
				"oauth_token":                          {Type: schema.TypeString, Computed: true},
				"oauth_token_secret":                   {Type: schema.TypeString, Computed: true},
				"consumer_key":                         {Type: schema.TypeString, Computed: true},
				"consumer_secret":                      {Type: schema.TypeString, Computed: true},
				"key":                                  {Type: schema.TypeString, Computed: true},
				"advertisers_id":                       {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				"sync_format":                          {Type: schema.TypeString, Computed: true},
				"bucket_service":                       {Type: schema.TypeString, Computed: true},
				"user_name":                            {Type: schema.TypeString, Computed: true},
				"report_url":                           {Type: schema.TypeString, Computed: true},
				"unique_id":                            {Type: schema.TypeString, Computed: true},
				"auth_type":                            {Type: schema.TypeString, Computed: true},
				"latest_version":                       {Type: schema.TypeString, Computed: true},
				"authorization_method":                 {Type: schema.TypeString, Computed: true},
				"service_version":                      {Type: schema.TypeString, Computed: true},
				"last_synced_changes__utc_":            {Type: schema.TypeString, Computed: true},
				"adobe_analytics_configurations": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sync_mode":          {Type: schema.TypeString, Computed: true},
							"report_suites":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"elements":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"metrics":            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"calculated_metrics": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
							"segments":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
						},
					},
				},
				"is_new_package": {Type: schema.TypeString, Computed: true},
			},
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	mapAddStr(msi, "service", resp.Data.Service)
	mapAddStr(msi, "service_version", intPointerToStr(resp.Data.ServiceVersion))
	mapAddStr(msi, "schema", resp.Data.Schema)
	mapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
	mapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
	mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
	mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
	mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
	mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))
	mapAddXInterface(msi, "status", dataSourceConnectorReadStatus(&resp))
	mapAddXInterface(msi, "config", dataSourceConnectorReadConfig(&resp))
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

// dataSourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func dataSourceConnectorReadStatus(resp *fivetran.ConnectorDetailsResponse) []interface{} {
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

func dataSourceConnectorReadStatusFlattenTasks(resp *fivetran.ConnectorDetailsResponse) []interface{} {
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

func dataSourceConnectorReadStatusFlattenWarnings(resp *fivetran.ConnectorDetailsResponse) []interface{} {
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
func dataSourceConnectorReadConfig(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	config := make([]interface{}, 1)

	c := make(map[string]interface{})
	mapAddStr(c, "schema", resp.Data.Config.Schema)
	mapAddStr(c, "table", resp.Data.Config.Table)
	mapAddStr(c, "sheet_id", resp.Data.Config.SheetID)
	mapAddStr(c, "named_range", resp.Data.Config.NamedRange)
	mapAddStr(c, "client_id", resp.Data.Config.ClientID)
	mapAddStr(c, "client_secret", resp.Data.Config.ClientSecret)
	mapAddStr(c, "technical_account_id", resp.Data.Config.TechnicalAccountID)
	mapAddStr(c, "organization_id", resp.Data.Config.OrganizationID)
	mapAddStr(c, "private_key", resp.Data.Config.PrivateKey)
	mapAddStr(c, "sync_mode", resp.Data.Config.SyncMode)
	mapAddXInterface(c, "report_suites", xStrXInterface(resp.Data.Config.ReportSuites))
	mapAddXInterface(c, "elements", xStrXInterface(resp.Data.Config.Elements))
	mapAddXInterface(c, "metrics", xStrXInterface(resp.Data.Config.Metrics))
	mapAddStr(c, "date_granularity", resp.Data.Config.DateGranularity)
	mapAddStr(c, "timeframe_months", resp.Data.Config.TimeframeMonths)
	mapAddStr(c, "source", resp.Data.Config.Source)
	mapAddStr(c, "s3bucket", resp.Data.Config.S3Bucket)
	mapAddStr(c, "s3role_arn", resp.Data.Config.S3RoleArn)
	mapAddStr(c, "abs_connection_string", resp.Data.Config.ABSConnectionString)
	mapAddStr(c, "abs_container_name", resp.Data.Config.ABSContainerName)
	mapAddStr(c, "ftp_host", resp.Data.Config.FTPHost)
	mapAddStr(c, "ftp_port", intPointerToStr(resp.Data.Config.FTPPort))
	mapAddStr(c, "ftp_user", resp.Data.Config.FTPUser)
	mapAddStr(c, "ftp_password", resp.Data.Config.FTPPassword)
	mapAddStr(c, "is_ftps", boolPointerToStr(resp.Data.Config.IsFTPS))
	mapAddStr(c, "sftp_host", resp.Data.Config.SFTPHost)
	mapAddStr(c, "sftp_port", intPointerToStr(resp.Data.Config.SFTPPort))
	mapAddStr(c, "sftp_user", resp.Data.Config.SFTPUser)
	mapAddStr(c, "sftp_password", resp.Data.Config.SFTPPassword)
	mapAddStr(c, "sftp_is_key_pair", boolPointerToStr(resp.Data.Config.SFTPIsKeyPair))
	mapAddXInterface(c, "advertisables", xStrXInterface(resp.Data.Config.Advertisables))
	mapAddStr(c, "report_type", resp.Data.Config.ReportType)
	mapAddXInterface(c, "dimensions", xStrXInterface(resp.Data.Config.Dimensions))
	mapAddStr(c, "schema_prefix", resp.Data.Config.SchemaPrefix)
	mapAddStr(c, "api_key", resp.Data.Config.APIKey)
	mapAddStr(c, "external_id", resp.Data.Config.ExternalID)
	mapAddStr(c, "role_arn", resp.Data.Config.RoleArn)
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
	mapAddStr(c, "skip_before", resp.Data.Config.SkipBefore)
	mapAddStr(c, "skip_after", resp.Data.Config.SkipAfter)
	mapAddXInterface(c, "project_credentials", dataSourceConnectorReadConfigFlattenProjectCredentials(resp))
	mapAddStr(c, "auth_mode", resp.Data.Config.AuthMode)
	mapAddStr(c, "username", resp.Data.Config.UserName)
	mapAddStr(c, "password", resp.Data.Config.Password)
	mapAddStr(c, "certificate", resp.Data.Config.Certificate)
	mapAddXInterface(c, "selected_exports", xStrXInterface(resp.Data.Config.SelectedExports))
	mapAddStr(c, "consumer_group", resp.Data.Config.ConsumerGroup)
	mapAddStr(c, "servers", resp.Data.Config.Servers)
	mapAddStr(c, "message_type", resp.Data.Config.MessageType)
	mapAddStr(c, "sync_type", resp.Data.Config.SyncType)
	mapAddStr(c, "security_protocol", resp.Data.Config.SecurityProtocol)
	mapAddXInterface(c, "apps", xStrXInterface(resp.Data.Config.Apps))
	mapAddXInterface(c, "sales_accounts", xStrXInterface(resp.Data.Config.SalesAccounts))
	mapAddXInterface(c, "finance_accounts", xStrXInterface(resp.Data.Config.FinanceAccounts))
	mapAddStr(c, "app_sync_mode", resp.Data.Config.AppSyncMode)
	mapAddStr(c, "sales_account_sync_mode", resp.Data.Config.SalesAccountSyncMode)
	mapAddStr(c, "finance_account_sync_mode", resp.Data.Config.FinanceAccountSyncMode)
	mapAddStr(c, "pem_certificate", resp.Data.Config.PEMCertificate)
	mapAddStr(c, "access_key_id", resp.Data.Config.AccessKeyID)
	mapAddStr(c, "secret_key", resp.Data.Config.SecretKey)
	mapAddStr(c, "home_folder", resp.Data.Config.HomeFolder)
	mapAddStr(c, "sync_data_locker", boolPointerToStr(resp.Data.Config.SyncDataLocker))
	mapAddXInterface(c, "projects", xStrXInterface(resp.Data.Config.Projects))
	mapAddStr(c, "function", resp.Data.Config.Function)
	mapAddStr(c, "region", resp.Data.Config.Region)
	mapAddStr(c, "secrets", resp.Data.Config.Secrets)
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
	mapAddXInterface(c, "user_profiles", xStrXInterface(resp.Data.Config.UserProfiles))
	mapAddXInterface(c, "report_configuration_ids", xStrXInterface(resp.Data.Config.ReportConfigurationIDs))
	mapAddStr(c, "enable_all_dimension_combinations", boolPointerToStr(resp.Data.Config.EnableAllDimensionCombinations))
	mapAddStr(c, "instance", resp.Data.Config.Instance)
	mapAddStr(c, "aws_region_code", resp.Data.Config.AWSRegionCode)
	mapAddXInterface(c, "accounts", xStrXInterface(resp.Data.Config.Accounts))
	mapAddXInterface(c, "fields", xStrXInterface(resp.Data.Config.Fields))
	mapAddXInterface(c, "breakdowns", xStrXInterface(resp.Data.Config.Breakdowns))
	mapAddXInterface(c, "action_breakdowns", xStrXInterface(resp.Data.Config.ActionBreakdowns))
	mapAddStr(c, "aggregation", resp.Data.Config.Aggregation)
	mapAddStr(c, "config_type", resp.Data.Config.ConfigType)
	mapAddStr(c, "prebuilt_report", resp.Data.Config.PrebuiltReport)
	mapAddStr(c, "action_report_time", resp.Data.Config.ActionReportTime)
	mapAddStr(c, "click_attribution_window", resp.Data.Config.ClickAttributionWindow)
	mapAddStr(c, "view_attribution_window", resp.Data.Config.ViewAttributionWindow)
	mapAddXInterface(c, "custom_tables", dataSourceConnectorReadConfigFlattenCustomTables(resp))
	mapAddXInterface(c, "pages", xStrXInterface(resp.Data.Config.Pages))
	mapAddStr(c, "subdomain", resp.Data.Config.Subdomain)
	mapAddStr(c, "host", resp.Data.Config.Host)
	mapAddStr(c, "port", intPointerToStr(resp.Data.Config.Port))
	mapAddStr(c, "user", resp.Data.Config.User)
	mapAddStr(c, "is_secure", resp.Data.Config.IsSecure)
	mapAddXInterface(c, "repositories", xStrXInterface(resp.Data.Config.Repositories))
	mapAddStr(c, "use_webhooks", boolPointerToStr(resp.Data.Config.UseWebhooks))
	mapAddXInterface(c, "dimension_attributes", xStrXInterface(resp.Data.Config.DimensionAttributes))
	mapAddXInterface(c, "columns", xStrXInterface(resp.Data.Config.Columns))
	mapAddStr(c, "network_code", resp.Data.Config.NetworkCode)
	mapAddStr(c, "customer_id", resp.Data.Config.CustomerID)
	mapAddXInterface(c, "manager_accounts", xStrXInterface(resp.Data.Config.ManagerAccounts))
	mapAddXInterface(c, "reports", dataSourceConnectorReadConfigFlattenReports(resp))
	mapAddStr(c, "conversion_window_size", intPointerToStr(resp.Data.Config.ConversionWindowSize))
	mapAddXInterface(c, "profiles", xStrXInterface(resp.Data.Config.Profiles))
	mapAddStr(c, "project_id", resp.Data.Config.ProjectID)
	mapAddStr(c, "dataset_id", resp.Data.Config.DatasetID)
	mapAddStr(c, "bucket_name", resp.Data.Config.BucketName)
	mapAddStr(c, "function_trigger", resp.Data.Config.FunctionTrigger)
	mapAddStr(c, "config_method", resp.Data.Config.ConfigMethod)
	mapAddStr(c, "query_id", resp.Data.Config.QueryID)
	mapAddStr(c, "update_config_on_each_sync", boolPointerToStr(resp.Data.Config.UpdateConfigOnEachSync))
	mapAddXInterface(c, "site_urls", xStrXInterface(resp.Data.Config.SiteURLs))
	mapAddStr(c, "path", resp.Data.Config.Path)
	mapAddStr(c, "on_premise", boolPointerToStr(resp.Data.Config.OnPremise))
	mapAddStr(c, "access_token", resp.Data.Config.AccessToken)
	mapAddStr(c, "view_through_attribution_window_size", resp.Data.Config.ViewThroughAttributionWindowSize)
	mapAddStr(c, "post_click_attribution_window_size", resp.Data.Config.PostClickAttributionWindowSize)
	mapAddStr(c, "use_api_keys", resp.Data.Config.UseAPIKeys)
	mapAddStr(c, "api_keys", resp.Data.Config.APIKeys)
	mapAddStr(c, "endpoint", resp.Data.Config.Endpoint)
	mapAddStr(c, "identity", resp.Data.Config.Identity)
	mapAddStr(c, "api_quota", intPointerToStr(resp.Data.Config.APIQuota))
	mapAddStr(c, "domain_name", resp.Data.Config.DomainName)
	mapAddStr(c, "resource_url", resp.Data.Config.ResourceURL)
	mapAddStr(c, "api_secret", resp.Data.Config.APISecret)
	mapAddXInterface(c, "hosts", xStrXInterface(resp.Data.Config.Hosts))
	mapAddStr(c, "tunnel_host", resp.Data.Config.TunnelHost)
	mapAddStr(c, "tunnel_port", intPointerToStr(resp.Data.Config.TunnelPort))
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
	mapAddStr(c, "daily_api_call_limit", intPointerToStr(resp.Data.Config.DailyAPICallLimit))
	mapAddStr(c, "time_zone", resp.Data.Config.TimeZone)
	mapAddStr(c, "integration_key", resp.Data.Config.IntegrationKey)
	mapAddXInterface(c, "advertisers", xStrXInterface(resp.Data.Config.Advertisers))
	mapAddStr(c, "engagement_attribution_window", resp.Data.Config.EngagementAttributionWindow)
	mapAddStr(c, "conversion_report_time", resp.Data.Config.ConversionReportTime)
	mapAddStr(c, "domain", resp.Data.Config.Domain)
	mapAddStr(c, "update_method", resp.Data.Config.UpdateMethod)
	mapAddStr(c, "replication_slot", resp.Data.Config.ReplicationSlot)
	mapAddStr(c, "data_center", resp.Data.Config.DataCenter)
	mapAddStr(c, "api_token", resp.Data.Config.APIToken)
	mapAddStr(c, "sub_domain", resp.Data.Config.SubDomain)
	mapAddStr(c, "test_table_name", resp.Data.Config.TestTableName)
	mapAddStr(c, "shop", resp.Data.Config.Shop)
	mapAddXInterface(c, "organizations", xStrXInterface(resp.Data.Config.Organizations))
	mapAddStr(c, "swipe_attribution_window", resp.Data.Config.SwipeAttributionWindow)
	mapAddStr(c, "api_access_token", resp.Data.Config.APIAccessToken)
	mapAddStr(c, "account_ids", resp.Data.Config.AccountIDs)
	mapAddStr(c, "sid", resp.Data.Config.SID)
	mapAddStr(c, "secret", resp.Data.Config.Secret)
	mapAddStr(c, "oauth_token", resp.Data.Config.OauthToken)
	mapAddStr(c, "oauth_token_secret", resp.Data.Config.OauthTokenSecret)
	mapAddStr(c, "consumer_key", resp.Data.Config.ConsumerKey)
	mapAddStr(c, "consumer_secret", resp.Data.Config.ConsumerSecret)
	mapAddStr(c, "key", resp.Data.Config.Key)
	mapAddXInterface(c, "advertisers_id", xStrXInterface(resp.Data.Config.AdvertisersID))
	mapAddStr(c, "sync_format", resp.Data.Config.SyncFormat)
	mapAddStr(c, "bucket_service", resp.Data.Config.BucketService)
	mapAddStr(c, "user_name", resp.Data.Config.UserName)
	mapAddStr(c, "report_url", resp.Data.Config.ReportURL)
	mapAddStr(c, "unique_id", resp.Data.Config.UniqueID)
	mapAddStr(c, "auth_type", resp.Data.Config.AuthType)
	mapAddStr(c, "latest_version", resp.Data.Config.LatestVersion)
	mapAddStr(c, "authorization_method", resp.Data.Config.AuthorizationMethod)
	mapAddStr(c, "service_version", resp.Data.Config.ServiceVersion)
	mapAddStr(c, "last_synced_changes__utc_", resp.Data.Config.LastSyncedChangesUtc)
	mapAddStr(c, "is_new_package", boolPointerToStr(resp.Data.Config.IsNewPackage))
	mapAddXInterface(c, "adobe_analytics_configurations", dataSourceConnectorReadConfigFlattenAdobeAnalyticsConfigurations(resp))
	config[0] = c

	return config
}

func dataSourceConnectorReadConfigFlattenProjectCredentials(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Config.ProjectCredentials) < 1 {
		return make([]interface{}, 0)
	}

	projectCredentials := make([]interface{}, len(resp.Data.Config.ProjectCredentials))
	for i, v := range resp.Data.Config.ProjectCredentials {
		pc := make(map[string]interface{})
		mapAddStr(pc, "project", v.Project)
		mapAddStr(pc, "api_key", v.APIKey)
		mapAddStr(pc, "secret_key", v.SecretKey)
		projectCredentials[i] = pc
	}

	return projectCredentials
}

func dataSourceConnectorReadConfigFlattenReports(resp *fivetran.ConnectorDetailsResponse) []interface{} {
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

func dataSourceConnectorReadConfigFlattenCustomTables(resp *fivetran.ConnectorDetailsResponse) []interface{} {
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

func dataSourceConnectorReadConfigFlattenAdobeAnalyticsConfigurations(resp *fivetran.ConnectorDetailsResponse) []interface{} {
	if len(resp.Data.Config.AdobeAnalyticsConfigurations) < 1 {
		return make([]interface{}, 0)
	}

	adobeAnalyticsConfigurations := make([]interface{}, len(resp.Data.Config.AdobeAnalyticsConfigurations))
	for i, v := range resp.Data.Config.AdobeAnalyticsConfigurations {
		aac := make(map[string]interface{})
		mapAddStr(aac, "sync_mode", v.SyncMode)
		mapAddXInterface(aac, "report_suites", xStrXInterface(v.ReportSuites))
		mapAddXInterface(aac, "elements", xStrXInterface(v.Elements))
		mapAddXInterface(aac, "metrics", xStrXInterface(v.Metrics))
		mapAddXInterface(aac, "calculated_metrics", xStrXInterface(v.CalculatedMetrics))
		mapAddXInterface(aac, "segments", xStrXInterface(v.Segments))
		adobeAnalyticsConfigurations[i] = aac
	}

	return adobeAnalyticsConfigurations
}
