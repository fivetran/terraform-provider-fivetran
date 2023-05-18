package fivetran

import (
	"strconv"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func connectorSchema(readonly bool) map[string]*schema.Schema {

	// Common for Resource and Datasource
	var result = map[string]*schema.Schema{
		// Id
		"id": {Type: schema.TypeString, Computed: !readonly, Required: readonly},

		// Computed
		"name":            {Type: schema.TypeString, Computed: true},
		"connected_by":    {Type: schema.TypeString, Computed: true},
		"created_at":      {Type: schema.TypeString, Computed: true},
		"succeeded_at":    {Type: schema.TypeString, Computed: true},
		"failed_at":       {Type: schema.TypeString, Computed: true},
		"service_version": {Type: schema.TypeString, Computed: true},
		"status":          connectorSchemaStatus(),

		// Required
		"group_id":           {Type: schema.TypeString, Required: !readonly, ForceNew: !readonly, Computed: readonly},
		"service":            {Type: schema.TypeString, Required: !readonly, ForceNew: !readonly, Computed: readonly},
		"destination_schema": connectorDestinationSchemaSchema(readonly),

		// Optional with default values in upstream
		"sync_frequency":    {Type: schema.TypeString, Optional: !readonly, Computed: true}, // Default: 360
		"schedule_type":     {Type: schema.TypeString, Optional: !readonly, Computed: true}, // Default: AUTO
		"paused":            {Type: schema.TypeString, Optional: !readonly, Computed: true}, // Default: false
		"pause_after_trial": {Type: schema.TypeString, Optional: !readonly, Computed: true}, // Default: false

		// Optional nullable in upstream
		"daily_sync_time": {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
		"config":          connectorSchemaConfig(readonly),
	}

	// Resource specific
	if !readonly {
		result["auth"] = connectorSchemaAuth()
		result["trust_certificates"] = &schema.Schema{Type: schema.TypeString, Optional: true}
		result["trust_fingerprints"] = &schema.Schema{Type: schema.TypeString, Optional: true}
		result["run_setup_tests"] = &schema.Schema{Type: schema.TypeString, Optional: true}

		// Internal resource attribute (no upstream value)
		result["last_updated"] = &schema.Schema{Type: schema.TypeString, Computed: true}
	}
	return result
}

func connectorSchemaStatus() *schema.Schema {
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

func getMaxItems(readonly bool) int {
	if readonly {
		return 0
	}
	return 1
}

func connectorDestinationSchemaSchema(readonly bool) *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList, MaxItems: getMaxItems(readonly),
		Required: !readonly, Computed: readonly,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":   {Type: schema.TypeString, Optional: !readonly, ForceNew: !readonly, Computed: readonly},
				"table":  {Type: schema.TypeString, Optional: !readonly, ForceNew: !readonly, Computed: readonly},
				"prefix": {Type: schema.TypeString, Optional: !readonly, ForceNew: !readonly, Computed: readonly},
			},
		},
	}
}

func connectorSchemaAuth() *schema.Schema {
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

func connectorSchemaConfig(readonly bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: !readonly,
		Computed: true,
		MaxItems: getMaxItems(readonly),
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Readonly config fields
				"latest_version":            {Type: schema.TypeString, Computed: true},
				"authorization_method":      {Type: schema.TypeString, Computed: true},
				"service_version":           {Type: schema.TypeString, Computed: true},
				"last_synced_changes__utc_": {Type: schema.TypeString, Computed: true},

				"public_key":  {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"external_id": {Type: schema.TypeString, Optional: !readonly, Computed: true},

				// Sensitive config fields, Fivetran returns this fields masked
				"oauth_token":        {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"oauth_token_secret": {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"consumer_key":       {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"client_secret":      {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"private_key":        {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"s3role_arn":         {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"ftp_password":       {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"sftp_password":      {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"api_key":            {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"role_arn":           {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"password":           {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"secret_key":         {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"pem_certificate":    {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"access_token":       {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"api_secret":         {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"api_access_token":   {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"secret":             {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"consumer_secret":    {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"secrets":            {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"api_token":          {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"encryption_key":     {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"pat":                {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"function_trigger":   {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"token_key":          {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"token_secret":       {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"agent_password":     {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"asm_password":       {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},
				"login_password":     {Type: schema.TypeString, Optional: !readonly, Sensitive: true, Computed: readonly},

				// Fields that always have default value in upstream (and should be marked as Computed to prevent drifting)
				// Boolean values
				"is_ftps":                           {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"sftp_is_key_pair":                  {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"sync_data_locker":                  {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"enable_all_dimension_combinations": {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"update_config_on_each_sync":        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"on_premise":                        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"use_api_keys":                      {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"is_new_package":                    {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"is_multi_entity_feature_enabled":   {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"always_encrypted":                  {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"is_secure":                         {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"use_webhooks":                      {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"eu_region":                         {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"is_keypair":                        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"is_account_level_connector":        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"use_oracle_rac":                    {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"asm_option":                        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"is_single_table_mode":              {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"is_public":                         {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"empty_header":                      {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"support_nested_columns":            {Type: schema.TypeString, Optional: !readonly, Computed: true},

				// Enum & int values
				"connection_type":                      {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"sync_method":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"sync_mode":                            {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"date_granularity":                     {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"timeframe_months":                     {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"report_type":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"aggregation":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"config_type":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"prebuilt_report":                      {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"action_report_time":                   {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"click_attribution_window":             {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"view_attribution_window":              {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"conversion_window_size":               {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"view_through_attribution_window_size": {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"post_click_attribution_window_size":   {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"update_method":                        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"swipe_attribution_window":             {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"api_type":                             {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"auth_type":                            {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"sync_format":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"app_sync_mode":                        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"sales_account_sync_mode":              {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"finance_account_sync_mode":            {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"source":                               {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"file_type":                            {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"compression":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"on_error":                             {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"append_file_option":                   {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"engagement_attribution_window":        {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"conversion_report_time":               {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"skip_before":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"skip_after":                           {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"ftp_port":                             {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"sftp_port":                            {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"port":                                 {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"tunnel_port":                          {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"daily_api_call_limit":                 {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"api_quota":                            {Type: schema.TypeString, Optional: !readonly, Computed: true},
				"agent_port":                           {Type: schema.TypeString, Optional: !readonly, Computed: true},

				// Usual fields
				"asm_oracle_home":       {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"asm_tns":               {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"pdb_name":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"agent_host":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"agent_user":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"agent_public_cert":     {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"agent_ora_home":        {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"tns":                   {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"asm_user":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"sap_user":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"sheet_id":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"named_range":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"client_id":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"technical_account_id":  {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"organization_id":       {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"s3bucket":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"abs_connection_string": {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"abs_container_name":    {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"folder_id":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"ftp_host":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"ftp_user":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"sftp_host":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"sftp_user":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"bucket":                {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"prefix":                {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"pattern":               {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"archive_pattern":       {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"null_sequence":         {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"delimiter":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"escape_char":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"auth_mode":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"certificate":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"consumer_group":        {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"servers":               {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"message_type":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"sync_type":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"security_protocol":     {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"access_key_id":         {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"home_folder":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"function":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"region":                {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"container_name":        {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"connection_string":     {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"function_app":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"function_name":         {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"function_key":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"merchant_id":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"api_url":               {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"cloud_storage_type":    {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"s3external_id":         {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"s3folder":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"gcs_bucket":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"gcs_folder":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"instance":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"aws_region_code":       {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"subdomain":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"host":                  {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"user":                  {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"network_code":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"customer_id":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"project_id":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"dataset_id":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"bucket_name":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"config_method":         {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"query_id":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"path":                  {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"endpoint":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"identity":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"domain_name":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"resource_url":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"tunnel_host":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"tunnel_user":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"database":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"datasource":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"account":               {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"role":                  {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"email":                 {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"account_id":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"server_url":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"user_key":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"api_version":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"time_zone":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"integration_key":       {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"domain":                {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"replication_slot":      {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"publication_name":      {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"data_center":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"sub_domain":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"test_table_name":       {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"shop":                  {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"sid":                   {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"key":                   {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"bucket_service":        {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"user_name":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"username":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"report_url":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"unique_id":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"base_url":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"entity_id":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"soap_uri":              {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"user_id":               {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"share_url":             {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"organization":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"access_key":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"domain_host_name":      {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"client_name":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"domain_type":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"connection_method":     {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"group_name":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"company_id":            {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"environment":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"list_strategy":         {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"csv_definition":        {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
				"export_storage_type":   {Type: schema.TypeString, Optional: !readonly, Computed: readonly},

				// String collections
				"report_suites":            {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"elements":                 {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"metrics":                  {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"advertisables":            {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"dimensions":               {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"selected_exports":         {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"apps":                     {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"sales_accounts":           {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"finance_accounts":         {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"projects":                 {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"user_profiles":            {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"report_configuration_ids": {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"accounts":                 {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"fields":                   {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"breakdowns":               {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"action_breakdowns":        {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"pages":                    {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"repositories":             {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"dimension_attributes":     {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"columns":                  {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"manager_accounts":         {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"profiles":                 {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"site_urls":                {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"api_keys":                 {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"advertisers_id":           {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"hosts":                    {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"advertisers":              {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"organizations":            {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"account_ids":              {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"packed_mode_tables":       {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"properties":               {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
				"primary_keys":             {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},

				// Objects collections
				"secrets_list": {Type: schema.TypeSet, Optional: !readonly, Computed: readonly,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"key":   {Type: schema.TypeString, Required: !readonly, Computed: readonly},
							"value": {Type: schema.TypeString, Required: !readonly, Computed: readonly, Sensitive: true},
						},
					},
				},

				"adobe_analytics_configurations": {Type: schema.TypeSet, Optional: !readonly, Computed: readonly,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sync_mode":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
							"report_suites":      {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"elements":           {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"metrics":            {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"calculated_metrics": {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"segments":           {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
						},
					},
				},
				"reports": {Type: schema.TypeSet, Optional: !readonly, Computed: readonly,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"table":           {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
							"config_type":     {Type: schema.TypeString, Optional: !readonly, Computed: true},
							"prebuilt_report": {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
							"report_type":     {Type: schema.TypeString, Optional: !readonly, Computed: true},
							"fields":          {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"dimensions":      {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"metrics":         {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"segments":        {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"filter":          {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
						},
					},
				},
				"custom_tables": {Type: schema.TypeSet, Optional: !readonly, Computed: readonly,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"table_name":               {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
							"config_type":              {Type: schema.TypeString, Optional: !readonly, Computed: true},
							"fields":                   {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"breakdowns":               {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"action_breakdowns":        {Type: schema.TypeSet, Optional: !readonly, Computed: readonly, Elem: &schema.Schema{Type: schema.TypeString}},
							"aggregation":              {Type: schema.TypeString, Optional: !readonly, Computed: true},
							"action_report_time":       {Type: schema.TypeString, Optional: !readonly, Computed: true},
							"click_attribution_window": {Type: schema.TypeString, Optional: !readonly, Computed: true},
							"view_attribution_window":  {Type: schema.TypeString, Optional: !readonly, Computed: true},
							"prebuilt_report_name":     {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
						},
					},
				},
				"project_credentials": {Type: schema.TypeSet, Optional: !readonly, Computed: readonly,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"project":    {Type: schema.TypeString, Optional: !readonly, Computed: readonly},
							"api_key":    {Type: schema.TypeString, Optional: !readonly, Computed: readonly, Sensitive: true},
							"secret_key": {Type: schema.TypeString, Optional: !readonly, Computed: readonly, Sensitive: true},
						},
					},
				},
			},
		},
	}
}

func connectorRead(currentConfig *[]interface{}, resp fivetran.ConnectorCustomMergedDetailsResponse) map[string]interface{} {
	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)
	mapAddStr(msi, "service", resp.Data.Service)
	mapAddStr(msi, "service_version", intPointerToStr(resp.Data.ServiceVersion))
	mapAddStr(msi, "name", resp.Data.Schema)
	mapAddXInterface(msi, "destination_schema", readDestinationSchema(resp.Data.Schema, resp.Data.Service))
	mapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
	mapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
	mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
	mapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
	mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
	mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
	mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))
	mapAddXInterface(msi, "status", connectorReadStatus(&resp))

	upstreamConfig := connectorReadConfig(&resp, currentConfig)

	if len(upstreamConfig) > 0 {
		mapAddXInterface(msi, "config", upstreamConfig)
	}

	return msi
}

// resourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func connectorReadStatus(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	status := make([]interface{}, 1)

	s := make(map[string]interface{})
	mapAddStr(s, "setup_state", resp.Data.Status.SetupState)
	mapAddStr(s, "sync_state", resp.Data.Status.SyncState)
	mapAddStr(s, "update_state", resp.Data.Status.UpdateState)
	mapAddStr(s, "is_historical_sync", boolPointerToStr(resp.Data.Status.IsHistoricalSync))
	mapAddXInterface(s, "tasks", connectorReadStatusFlattenTasks(resp))
	mapAddXInterface(s, "warnings", connectorReadStatusFlattenWarnings(resp))
	status[0] = s

	return status
}

func connectorReadStatusFlattenTasks(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func connectorReadStatusFlattenWarnings(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func tryReadStringValue(target, source map[string]interface{}, key string) {
	if v, ok := source[key].(string); ok {
		mapAddStr(target, key, v)
	}
}

func tryReadBooleanValue(target, source map[string]interface{}, key string) {
	if v, ok := source[key].(bool); ok {
		mapAddStr(target, key, boolToStr(v))
	}
}

func tryReadIntegerValue(target, source map[string]interface{}, key string) {
	if v, ok := source[key].(float64); ok {
		mapAddStr(target, key, strconv.Itoa((int(v))))
	}
}

func tryReadList(target, source map[string]interface{}, key string) {
	if v, ok := source[key].([]interface{}); ok {
		mapAddXInterface(target, key, v)
	}
}

// dataSourceConnectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func connectorReadConfig(resp *fivetran.ConnectorCustomMergedDetailsResponse, currentConfig *[]interface{}) []interface{} {
	config := make([]interface{}, 1)

	c := make(map[string]interface{})

	if currentConfig != nil {
		// get sensitive fields from the currentConfig to prevent drifting (Fivetran returns this values masked)
		if len(*currentConfig) > 0 {
			resourceConfig := (*currentConfig)[0].(map[string]interface{})
			mapAddStr(c, "password", resourceConfig["password"].(string))
			mapAddStr(c, "client_secret", resourceConfig["client_secret"].(string))
			mapAddStr(c, "private_key", resourceConfig["private_key"].(string))
			mapAddStr(c, "s3role_arn", resourceConfig["s3role_arn"].(string))
			mapAddStr(c, "ftp_password", resourceConfig["ftp_password"].(string))
			mapAddStr(c, "sftp_password", resourceConfig["sftp_password"].(string))
			mapAddStr(c, "api_key", resourceConfig["api_key"].(string))
			mapAddStr(c, "role_arn", resourceConfig["role_arn"].(string))
			mapAddStr(c, "secret_key", resourceConfig["secret_key"].(string))
			mapAddStr(c, "pem_certificate", resourceConfig["pem_certificate"].(string))
			mapAddStr(c, "access_token", resourceConfig["access_token"].(string))
			mapAddStr(c, "api_secret", resourceConfig["api_secret"].(string))
			mapAddStr(c, "api_access_token", resourceConfig["api_access_token"].(string))
			mapAddStr(c, "secret", resourceConfig["secret"].(string))
			mapAddStr(c, "consumer_secret", resourceConfig["consumer_secret"].(string))
			mapAddStr(c, "secrets", resourceConfig["secrets"].(string))
			mapAddStr(c, "api_token", resourceConfig["api_token"].(string))
			mapAddStr(c, "consumer_key", resourceConfig["consumer_key"].(string))
			mapAddStr(c, "encryption_key", resourceConfig["encryption_key"].(string))
			mapAddStr(c, "oauth_token", resourceConfig["oauth_token"].(string))
			mapAddStr(c, "oauth_token_secret", resourceConfig["oauth_token_secret"].(string))
			mapAddStr(c, "pat", resourceConfig["pat"].(string))
			mapAddStr(c, "function_trigger", resourceConfig["function_trigger"].(string))
			mapAddStr(c, "token_key", resourceConfig["token_key"].(string))
			mapAddStr(c, "token_secret", resourceConfig["token_secret"].(string))
			mapAddStr(c, "agent_password", resourceConfig["agent_password"].(string))
			mapAddStr(c, "asm_password", resourceConfig["asm_password"].(string))
			mapAddStr(c, "login_password", resourceConfig["login_password"].(string))

			mapAddXInterface(c, "api_keys", resourceConfig["api_keys"].(*schema.Set).List())
		}
	} else {
		// If current config is null method called from dataSource
		mapAddStr(c, "password", resp.Data.Config.Password)
		mapAddStr(c, "client_secret", resp.Data.Config.ClientSecret)
		mapAddStr(c, "private_key", resp.Data.Config.PrivateKey)
		mapAddStr(c, "s3role_arn", resp.Data.Config.S3RoleArn)
		mapAddStr(c, "ftp_password", resp.Data.Config.FTPPassword)
		mapAddStr(c, "sftp_password", resp.Data.Config.SFTPPassword)
		mapAddStr(c, "api_key", resp.Data.Config.APIKey)
		mapAddStr(c, "role_arn", resp.Data.Config.RoleArn)
		mapAddStr(c, "secret_key", resp.Data.Config.SecretKey)
		mapAddStr(c, "pem_certificate", resp.Data.Config.PEMCertificate)
		mapAddStr(c, "access_token", resp.Data.Config.AccessToken)
		mapAddStr(c, "api_secret", resp.Data.Config.APISecret)
		mapAddStr(c, "api_access_token", resp.Data.Config.APIAccessToken)
		mapAddStr(c, "secret", resp.Data.Config.Secret)
		mapAddStr(c, "consumer_secret", resp.Data.Config.ConsumerSecret)
		mapAddStr(c, "secrets", resp.Data.Config.Secrets)
		mapAddStr(c, "api_token", resp.Data.Config.APIToken)
		mapAddStr(c, "consumer_key", resp.Data.Config.ConsumerKey)
		mapAddStr(c, "encryption_key", resp.Data.Config.EncryptionKey)
		mapAddStr(c, "oauth_token", resp.Data.Config.OauthToken)
		mapAddStr(c, "oauth_token_secret", resp.Data.Config.OauthTokenSecret)
		mapAddStr(c, "pat", resp.Data.Config.PAT)
		mapAddStr(c, "function_trigger", resp.Data.Config.FunctionTrigger)
		mapAddStr(c, "token_key", resp.Data.Config.TokenKey)
		mapAddStr(c, "token_secret", resp.Data.Config.TokenSecret)

		mapAddXInterface(c, "api_keys", xStrXInterface(resp.Data.Config.APIKeys))

		tryReadStringValue(c, resp.Data.CustomConfig, "agent_password")
		tryReadStringValue(c, resp.Data.CustomConfig, "asm_password")
		tryReadStringValue(c, resp.Data.CustomConfig, "login_password")
	}

	mapAddXInterface(c, "project_credentials", connectorReadConfigFlattenProjectCredentials(resp, currentConfig))
	mapAddXInterface(c, "secrets_list", connectorReadConfigFlattenSecretsList(resp, currentConfig))

	// Collections
	mapAddXInterface(c, "report_suites", xStrXInterface(resp.Data.Config.ReportSuites))
	mapAddXInterface(c, "elements", xStrXInterface(resp.Data.Config.Elements))
	mapAddXInterface(c, "metrics", xStrXInterface(resp.Data.Config.Metrics))
	mapAddXInterface(c, "advertisables", xStrXInterface(resp.Data.Config.Advertisables))
	mapAddXInterface(c, "dimensions", xStrXInterface(resp.Data.Config.Dimensions))
	mapAddXInterface(c, "selected_exports", xStrXInterface(resp.Data.Config.SelectedExports))
	mapAddXInterface(c, "apps", xStrXInterface(resp.Data.Config.Apps))
	mapAddXInterface(c, "sales_accounts", xStrXInterface(resp.Data.Config.SalesAccounts))
	mapAddXInterface(c, "finance_accounts", xStrXInterface(resp.Data.Config.FinanceAccounts))
	mapAddXInterface(c, "projects", xStrXInterface(resp.Data.Config.Projects))
	mapAddXInterface(c, "user_profiles", xStrXInterface(resp.Data.Config.UserProfiles))
	mapAddXInterface(c, "report_configuration_ids", xStrXInterface(resp.Data.Config.ReportConfigurationIDs))
	mapAddXInterface(c, "custom_tables", connectorReadConfigFlattenCustomTables(resp))
	mapAddXInterface(c, "pages", xStrXInterface(resp.Data.Config.Pages))
	mapAddXInterface(c, "accounts", xStrXInterface(resp.Data.Config.Accounts))
	mapAddXInterface(c, "fields", xStrXInterface(resp.Data.Config.Fields))
	mapAddXInterface(c, "breakdowns", xStrXInterface(resp.Data.Config.Breakdowns))
	mapAddXInterface(c, "action_breakdowns", xStrXInterface(resp.Data.Config.ActionBreakdowns))
	mapAddXInterface(c, "repositories", xStrXInterface(resp.Data.Config.Repositories))
	mapAddXInterface(c, "dimension_attributes", xStrXInterface(resp.Data.Config.DimensionAttributes))
	mapAddXInterface(c, "columns", xStrXInterface(resp.Data.Config.Columns))
	mapAddXInterface(c, "manager_accounts", xStrXInterface(resp.Data.Config.ManagerAccounts))
	mapAddXInterface(c, "reports", connectorReadConfigFlattenReports(resp))
	mapAddXInterface(c, "site_urls", xStrXInterface(resp.Data.Config.SiteURLs))
	mapAddXInterface(c, "profiles", xStrXInterface(resp.Data.Config.Profiles))
	mapAddXInterface(c, "hosts", xStrXInterface(resp.Data.Config.Hosts))
	mapAddXInterface(c, "adobe_analytics_configurations", connectorReadConfigFlattenAdobeAnalyticsConfigurations(resp))
	mapAddXInterface(c, "advertisers", xStrXInterface(resp.Data.Config.Advertisers))
	mapAddXInterface(c, "organizations", xStrXInterface(resp.Data.Config.Organizations))
	mapAddXInterface(c, "account_ids", xStrXInterface(resp.Data.Config.AccountIDs))
	mapAddXInterface(c, "advertisers_id", xStrXInterface(resp.Data.Config.AdvertisersID))

	tryReadList(c, resp.Data.CustomConfig, "packed_mode_tables")
	tryReadList(c, resp.Data.CustomConfig, "properties")
	tryReadList(c, resp.Data.CustomConfig, "primary_keys")

	// Boolean fields
	mapAddStr(c, "is_ftps", boolPointerToStr(resp.Data.Config.IsFTPS))
	mapAddStr(c, "sftp_is_key_pair", boolPointerToStr(resp.Data.Config.SFTPIsKeyPair))
	mapAddStr(c, "is_keypair", boolPointerToStr(resp.Data.Config.IsKeypair))
	mapAddStr(c, "sync_data_locker", boolPointerToStr(resp.Data.Config.SyncDataLocker))
	mapAddStr(c, "enable_all_dimension_combinations", boolPointerToStr(resp.Data.Config.EnableAllDimensionCombinations))
	mapAddStr(c, "use_webhooks", boolPointerToStr(resp.Data.Config.UseWebhooks))
	mapAddStr(c, "update_config_on_each_sync", boolPointerToStr(resp.Data.Config.UpdateConfigOnEachSync))
	mapAddStr(c, "on_premise", boolPointerToStr(resp.Data.Config.OnPremise))
	mapAddStr(c, "always_encrypted", boolPointerToStr(resp.Data.Config.AlwaysEncrypted))
	mapAddStr(c, "is_new_package", boolPointerToStr(resp.Data.Config.IsNewPackage))
	mapAddStr(c, "is_multi_entity_feature_enabled", boolPointerToStr(resp.Data.Config.IsMultiEntityFeatureEnabled))
	mapAddStr(c, "eu_region", boolPointerToStr(resp.Data.Config.EuRegion))
	mapAddStr(c, "is_secure", boolPointerToStr(resp.Data.Config.IsSecure))
	mapAddStr(c, "use_api_keys", boolPointerToStr(resp.Data.Config.UseAPIKeys))
	tryReadBooleanValue(c, resp.Data.CustomConfig, "is_account_level_connector")
	tryReadBooleanValue(c, resp.Data.CustomConfig, "use_oracle_rac")
	tryReadBooleanValue(c, resp.Data.CustomConfig, "asm_option")
	tryReadBooleanValue(c, resp.Data.CustomConfig, "is_single_table_mode")
	tryReadBooleanValue(c, resp.Data.CustomConfig, "is_public")
	tryReadBooleanValue(c, resp.Data.CustomConfig, "empty_header")
	tryReadBooleanValue(c, resp.Data.CustomConfig, "support_nested_columns")

	// Integer fields
	mapAddStr(c, "ftp_port", intPointerToStr(resp.Data.Config.FTPPort))
	mapAddStr(c, "sftp_port", intPointerToStr(resp.Data.Config.SFTPPort))
	mapAddStr(c, "conversion_window_size", intPointerToStr(resp.Data.Config.ConversionWindowSize))
	mapAddStr(c, "port", intPointerToStr(resp.Data.Config.Port))
	mapAddStr(c, "api_quota", intPointerToStr(resp.Data.Config.APIQuota))
	mapAddStr(c, "tunnel_port", intPointerToStr(resp.Data.Config.TunnelPort))
	mapAddStr(c, "daily_api_call_limit", intPointerToStr(resp.Data.Config.DailyAPICallLimit))
	mapAddStr(c, "skip_after", intPointerToStr(resp.Data.Config.SkipAfter))
	mapAddStr(c, "skip_before", intPointerToStr(resp.Data.Config.SkipBefore))
	tryReadIntegerValue(c, resp.Data.CustomConfig, "agent_port")

	// String fields
	mapAddStr(c, "sheet_id", resp.Data.Config.SheetID)
	mapAddStr(c, "share_url", resp.Data.Config.ShareURL)
	mapAddStr(c, "named_range", resp.Data.Config.NamedRange)
	mapAddStr(c, "client_id", resp.Data.Config.ClientID)
	mapAddStr(c, "technical_account_id", resp.Data.Config.TechnicalAccountID)
	mapAddStr(c, "organization_id", resp.Data.Config.OrganizationID)

	tryReadStringValue(c, resp.Data.CustomConfig, "sync_method")
	tryReadStringValue(c, resp.Data.CustomConfig, "group_name")
	tryReadStringValue(c, resp.Data.CustomConfig, "pdb_name")
	tryReadStringValue(c, resp.Data.CustomConfig, "agent_host")
	tryReadStringValue(c, resp.Data.CustomConfig, "agent_user")
	tryReadStringValue(c, resp.Data.CustomConfig, "agent_public_cert")
	tryReadStringValue(c, resp.Data.CustomConfig, "agent_ora_home")
	tryReadStringValue(c, resp.Data.CustomConfig, "tns")
	tryReadStringValue(c, resp.Data.CustomConfig, "asm_user")
	tryReadStringValue(c, resp.Data.CustomConfig, "asm_oracle_home")
	tryReadStringValue(c, resp.Data.CustomConfig, "asm_tns")
	tryReadStringValue(c, resp.Data.CustomConfig, "sap_user")
	tryReadStringValue(c, resp.Data.CustomConfig, "organization")
	tryReadStringValue(c, resp.Data.CustomConfig, "access_key")
	tryReadStringValue(c, resp.Data.CustomConfig, "domain_host_name")
	tryReadStringValue(c, resp.Data.CustomConfig, "client_name")
	tryReadStringValue(c, resp.Data.CustomConfig, "domain_type")
	tryReadStringValue(c, resp.Data.CustomConfig, "connection_method")
	tryReadStringValue(c, resp.Data.CustomConfig, "company_id")
	tryReadStringValue(c, resp.Data.CustomConfig, "environment")
	tryReadStringValue(c, resp.Data.CustomConfig, "list_strategy")
	tryReadStringValue(c, resp.Data.CustomConfig, "csv_definition")
	tryReadStringValue(c, resp.Data.CustomConfig, "export_storage_type")

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
	mapAddStr(c, "network_code", resp.Data.Config.NetworkCode)
	mapAddStr(c, "customer_id", resp.Data.Config.CustomerID)
	mapAddStr(c, "project_id", resp.Data.Config.ProjectID)
	mapAddStr(c, "dataset_id", resp.Data.Config.DatasetID)
	mapAddStr(c, "bucket_name", resp.Data.Config.BucketName)
	mapAddStr(c, "config_method", resp.Data.Config.ConfigMethod)
	mapAddStr(c, "query_id", resp.Data.Config.QueryID)
	mapAddStr(c, "path", resp.Data.Config.Path)
	mapAddStr(c, "view_through_attribution_window_size", resp.Data.Config.ViewThroughAttributionWindowSize)
	mapAddStr(c, "post_click_attribution_window_size", resp.Data.Config.PostClickAttributionWindowSize)
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

func connectorReadConfigFlattenProjectCredentials(resp *fivetran.ConnectorCustomMergedDetailsResponse, currentConfig *[]interface{}) []interface{} {
	if len(resp.Data.Config.ProjectCredentials) < 1 {
		return make([]interface{}, 0)
	}

	projectCredentials := make([]interface{}, len(resp.Data.Config.ProjectCredentials))
	for i, v := range resp.Data.Config.ProjectCredentials {
		pc := make(map[string]interface{})
		mapAddStr(pc, "project", v.Project)
		if currentConfig != nil && len(*currentConfig) > 0 {
			// The REST API sends the fields "api_key" and "secret_key" masked. We use the state stored config here.
			mapAddStr(pc, "api_key", connectorReadConfigFlattenProjectCredentialsGetStateValue(v.Project, "api_key", *currentConfig))
			mapAddStr(pc, "secret_key", connectorReadConfigFlattenProjectCredentialsGetStateValue(v.Project, "secret_key", *currentConfig))
		} else {
			// On Import these values will be masked, but we can't rely on state
			mapAddStr(pc, "api_key", v.APIKey)
			mapAddStr(pc, "secret_key", v.SecretKey)
		}
		projectCredentials[i] = pc
	}

	return projectCredentials
}

func connectorReadConfigFlattenSecretsList(resp *fivetran.ConnectorCustomMergedDetailsResponse, currentConfig *[]interface{}) []interface{} {
	if len(resp.Data.Config.SecretsList) < 1 {
		return make([]interface{}, 0)
	}
	secretsList := make([]interface{}, len(resp.Data.Config.SecretsList))
	for i, v := range resp.Data.Config.SecretsList {
		s := make(map[string]interface{})
		mapAddStr(s, "key", v.Key)
		if currentConfig != nil && len(*currentConfig) > 0 {
			mapAddStr(s, "value", connectorReadConfigFlattenSecretsListGetStateValue(v.Key, *currentConfig))
		} else {
			mapAddStr(s, "value", v.Value)
		}
		secretsList[i] = s
	}

	return secretsList
}

func connectorReadConfigFlattenProjectCredentialsGetStateValue(project, key string, currentConfig []interface{}) string {
	result := getSubcollectionElementValue("project_credentials", "project", project, key, currentConfig)

	if result == nil {
		return ""
	}

	return result.(string)
}

func connectorReadConfigFlattenSecretsListGetStateValue(key string, currentConfig []interface{}) string {
	result := getSubcollectionElementValue("secrets_list", "key", key, "value", currentConfig)

	if result == nil {
		return ""
	}

	return result.(string)
}

func connectorReadConfigFlattenReports(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func connectorReadConfigFlattenAdobeAnalyticsConfigurations(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func connectorReadConfigFlattenCustomTables(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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
