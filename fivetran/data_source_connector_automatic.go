package fivetran

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/Jeffail/gabs/v2"
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
		path := "schemas." + service + ".properties.config.properties"
		if service == "adroll_config_V1" {
			log.Output(1, "luka")
		}
		newProperties := getDataSourceProperties(path)
		for k, v := range *newProperties {
			if k == "reports" {
				fmt.Printf("reports fields now\n")
			}
			if val, ok := properties[k]; ok {

				if k == "reports" {
					fmt.Printf("Type of val.Elem is %T\n", val.Elem)
				}

				if val.Type == schema.TypeList {
					if v2, ok := val.Elem.(*schema.Resource); ok {
						fmt.Printf("reports fields now 2\n")
						if vX1, ok := v.Elem.(*schema.Resource); ok {
							for kY, vY := range vX1.Schema {
								v2.Schema[kY] = vY
							}
							val.Elem = v2
							properties[k] = val
							continue
						}
					} else if v2, ok := val.Elem.(*schema.Schema); ok {
						if v3, ok := v2.Elem.(*schema.Resource); ok {
							fmt.Printf("reports fields now 2\n")
							if vX1, ok := v.Elem.(*schema.Resource); ok {
								for kY, vY := range vX1.Schema {
									v3.Schema[kY] = vY
								}
								val.Elem = v3
								properties[k] = val
								continue
							}
						}
					} else if v2, ok := val.Elem.(map[string]*schema.Schema); ok {
						fmt.Printf("reports fields now 2\n")
						fmt.Printf(intToStr(len(v2)))
						continue
					}
				}
			}
			properties[k] = v
		}
	}

	return &schema.Schema{Type: schema.TypeList, Optional: true, Computed: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: properties,
		},
	}

	// return &schema.Schema{Type: schema.TypeList, Computed: true,
	// 	Elem: &schema.Resource{
	// 		Schema: map[string]*schema.Schema{
	// 			"table":                      {Type: schema.TypeString, Computed: true},
	// 			"sheet_id":                   {Type: schema.TypeString, Computed: true},
	// 			"share_url":                  {Type: schema.TypeString, Computed: true},
	// 			"named_range":                {Type: schema.TypeString, Computed: true},
	// 			"client_id":                  {Type: schema.TypeString, Computed: true},
	// 			"client_secret":              {Type: schema.TypeString, Computed: true},
	// 			"technical_account_id":       {Type: schema.TypeString, Computed: true},
	// 			"organization_id":            {Type: schema.TypeString, Computed: true},
	// 			"private_key":                {Type: schema.TypeString, Computed: true},
	// 			"sync_method":                {Type: schema.TypeString, Computed: true},
	// 			"sync_mode":                  {Type: schema.TypeString, Computed: true},
	// 			"report_suites":              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"elements":                   {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"metrics":                    {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"date_granularity":           {Type: schema.TypeString, Computed: true},
	// 			"timeframe_months":           {Type: schema.TypeString, Computed: true},
	// 			"source":                     {Type: schema.TypeString, Computed: true},
	// 			"s3bucket":                   {Type: schema.TypeString, Computed: true},
	// 			"s3role_arn":                 {Type: schema.TypeString, Computed: true},
	// 			"abs_connection_string":      {Type: schema.TypeString, Computed: true},
	// 			"abs_container_name":         {Type: schema.TypeString, Computed: true},
	// 			"folder_id":                  {Type: schema.TypeString, Computed: true},
	// 			"ftp_host":                   {Type: schema.TypeString, Computed: true},
	// 			"ftp_port":                   {Type: schema.TypeString, Computed: true},
	// 			"ftp_user":                   {Type: schema.TypeString, Computed: true},
	// 			"ftp_password":               {Type: schema.TypeString, Computed: true},
	// 			"is_ftps":                    {Type: schema.TypeString, Computed: true},
	// 			"sftp_host":                  {Type: schema.TypeString, Computed: true},
	// 			"sftp_port":                  {Type: schema.TypeString, Computed: true},
	// 			"sftp_user":                  {Type: schema.TypeString, Computed: true},
	// 			"sftp_password":              {Type: schema.TypeString, Computed: true},
	// 			"sftp_is_key_pair":           {Type: schema.TypeString, Computed: true},
	// 			"is_keypair":                 {Type: schema.TypeString, Computed: true},
	// 			"is_account_level_connector": {Type: schema.TypeString, Computed: true},
	// 			"advertisables":              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"report_type":                {Type: schema.TypeString, Computed: true},
	// 			"dimensions":                 {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"api_key":                    {Type: schema.TypeString, Computed: true},
	// 			"external_id":                {Type: schema.TypeString, Computed: true},
	// 			"role_arn":                   {Type: schema.TypeString, Computed: true},
	// 			"bucket":                     {Type: schema.TypeString, Computed: true},
	// 			"prefix":                     {Type: schema.TypeString, Computed: true},
	// 			"pattern":                    {Type: schema.TypeString, Computed: true},
	// 			"file_type":                  {Type: schema.TypeString, Computed: true},
	// 			"compression":                {Type: schema.TypeString, Computed: true},
	// 			"on_error":                   {Type: schema.TypeString, Computed: true},
	// 			"append_file_option":         {Type: schema.TypeString, Computed: true},
	// 			"archive_pattern":            {Type: schema.TypeString, Computed: true},
	// 			"null_sequence":              {Type: schema.TypeString, Computed: true},
	// 			"delimiter":                  {Type: schema.TypeString, Computed: true},
	// 			"escape_char":                {Type: schema.TypeString, Computed: true},
	// 			"skip_before":                {Type: schema.TypeString, Computed: true},
	// 			"skip_after":                 {Type: schema.TypeString, Computed: true},
	// 			"project_credentials": {Type: schema.TypeList, Computed: true,
	// 				Elem: &schema.Resource{
	// 					Schema: map[string]*schema.Schema{
	// 						"project":    {Type: schema.TypeString, Computed: true},
	// 						"api_key":    {Type: schema.TypeString, Computed: true, Sensitive: true},
	// 						"secret_key": {Type: schema.TypeString, Computed: true, Sensitive: true},
	// 					},
	// 				},
	// 			},
	// 			"auth_mode":                         {Type: schema.TypeString, Computed: true},
	// 			"user_name":                         {Type: schema.TypeString, Computed: true},
	// 			"username":                          {Type: schema.TypeString, Computed: true},
	// 			"password":                          {Type: schema.TypeString, Computed: true},
	// 			"certificate":                       {Type: schema.TypeString, Computed: true},
	// 			"selected_exports":                  {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"consumer_group":                    {Type: schema.TypeString, Computed: true},
	// 			"servers":                           {Type: schema.TypeString, Computed: true},
	// 			"message_type":                      {Type: schema.TypeString, Computed: true},
	// 			"sync_type":                         {Type: schema.TypeString, Computed: true},
	// 			"security_protocol":                 {Type: schema.TypeString, Computed: true},
	// 			"apps":                              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"sales_accounts":                    {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"finance_accounts":                  {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"app_sync_mode":                     {Type: schema.TypeString, Computed: true},
	// 			"sales_account_sync_mode":           {Type: schema.TypeString, Computed: true},
	// 			"finance_account_sync_mode":         {Type: schema.TypeString, Computed: true},
	// 			"pem_certificate":                   {Type: schema.TypeString, Computed: true},
	// 			"access_key_id":                     {Type: schema.TypeString, Computed: true},
	// 			"secret_key":                        {Type: schema.TypeString, Computed: true},
	// 			"home_folder":                       {Type: schema.TypeString, Computed: true},
	// 			"sync_data_locker":                  {Type: schema.TypeString, Computed: true},
	// 			"projects":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"function":                          {Type: schema.TypeString, Computed: true},
	// 			"region":                            {Type: schema.TypeString, Computed: true},
	// 			"secrets":                           {Type: schema.TypeString, Computed: true},
	// 			"container_name":                    {Type: schema.TypeString, Computed: true},
	// 			"connection_string":                 {Type: schema.TypeString, Computed: true},
	// 			"connection_type":                   {Type: schema.TypeString, Computed: true},
	// 			"function_app":                      {Type: schema.TypeString, Computed: true},
	// 			"function_name":                     {Type: schema.TypeString, Computed: true},
	// 			"function_key":                      {Type: schema.TypeString, Computed: true},
	// 			"public_key":                        {Type: schema.TypeString, Computed: true},
	// 			"merchant_id":                       {Type: schema.TypeString, Computed: true},
	// 			"api_url":                           {Type: schema.TypeString, Computed: true},
	// 			"cloud_storage_type":                {Type: schema.TypeString, Computed: true},
	// 			"s3external_id":                     {Type: schema.TypeString, Computed: true},
	// 			"s3folder":                          {Type: schema.TypeString, Computed: true},
	// 			"gcs_bucket":                        {Type: schema.TypeString, Computed: true},
	// 			"gcs_folder":                        {Type: schema.TypeString, Computed: true},
	// 			"user_profiles":                     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"report_configuration_ids":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"enable_all_dimension_combinations": {Type: schema.TypeString, Computed: true},
	// 			"instance":                          {Type: schema.TypeString, Computed: true},
	// 			"aws_region_code":                   {Type: schema.TypeString, Computed: true},
	// 			"accounts":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"fields":                            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"breakdowns":                        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"action_breakdowns":                 {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"aggregation":                       {Type: schema.TypeString, Computed: true},
	// 			"config_type":                       {Type: schema.TypeString, Computed: true},
	// 			"prebuilt_report":                   {Type: schema.TypeString, Computed: true},
	// 			"action_report_time":                {Type: schema.TypeString, Computed: true},
	// 			"click_attribution_window":          {Type: schema.TypeString, Computed: true},
	// 			"view_attribution_window":           {Type: schema.TypeString, Computed: true},
	// 			"custom_tables": {Type: schema.TypeList, Computed: true,
	// 				Elem: &schema.Resource{
	// 					Schema: map[string]*schema.Schema{
	// 						"table_name":               {Type: schema.TypeString, Computed: true},
	// 						"config_type":              {Type: schema.TypeString, Computed: true},
	// 						"fields":                   {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"breakdowns":               {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"action_breakdowns":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"aggregation":              {Type: schema.TypeString, Computed: true},
	// 						"action_report_time":       {Type: schema.TypeString, Computed: true},
	// 						"click_attribution_window": {Type: schema.TypeString, Computed: true},
	// 						"view_attribution_window":  {Type: schema.TypeString, Computed: true},
	// 						"prebuilt_report_name":     {Type: schema.TypeString, Computed: true},
	// 					},
	// 				},
	// 			},
	// 			"pages":                {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"subdomain":            {Type: schema.TypeString, Computed: true},
	// 			"host":                 {Type: schema.TypeString, Computed: true},
	// 			"port":                 {Type: schema.TypeString, Computed: true},
	// 			"user":                 {Type: schema.TypeString, Computed: true},
	// 			"is_secure":            {Type: schema.TypeString, Computed: true},
	// 			"repositories":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"use_webhooks":         {Type: schema.TypeString, Computed: true},
	// 			"dimension_attributes": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"columns":              {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"network_code":         {Type: schema.TypeString, Computed: true},
	// 			"customer_id":          {Type: schema.TypeString, Computed: true},
	// 			"manager_accounts":     {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"reports": {Type: schema.TypeList, Computed: true,
	// 				Elem: &schema.Resource{
	// 					Schema: map[string]*schema.Schema{
	// 						"table":           {Type: schema.TypeString, Computed: true},
	// 						"config_type":     {Type: schema.TypeString, Computed: true},
	// 						"prebuilt_report": {Type: schema.TypeString, Computed: true},
	// 						"report_type":     {Type: schema.TypeString, Computed: true},
	// 						"fields":          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"dimensions":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"metrics":         {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"segments":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"filter":          {Type: schema.TypeString, Computed: true},
	// 					},
	// 				},
	// 			},
	// 			"conversion_window_size":               {Type: schema.TypeString, Computed: true},
	// 			"profiles":                             {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"project_id":                           {Type: schema.TypeString, Computed: true},
	// 			"dataset_id":                           {Type: schema.TypeString, Computed: true},
	// 			"bucket_name":                          {Type: schema.TypeString, Computed: true},
	// 			"function_trigger":                     {Type: schema.TypeString, Computed: true},
	// 			"config_method":                        {Type: schema.TypeString, Computed: true},
	// 			"query_id":                             {Type: schema.TypeString, Computed: true},
	// 			"update_config_on_each_sync":           {Type: schema.TypeString, Computed: true},
	// 			"site_urls":                            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"path":                                 {Type: schema.TypeString, Computed: true},
	// 			"on_premise":                           {Type: schema.TypeString, Computed: true},
	// 			"access_token":                         {Type: schema.TypeString, Computed: true},
	// 			"view_through_attribution_window_size": {Type: schema.TypeString, Computed: true},
	// 			"post_click_attribution_window_size":   {Type: schema.TypeString, Computed: true},
	// 			"use_api_keys":                         {Type: schema.TypeString, Computed: true},
	// 			"api_keys":                             {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"endpoint":                             {Type: schema.TypeString, Computed: true},
	// 			"identity":                             {Type: schema.TypeString, Computed: true},
	// 			"api_quota":                            {Type: schema.TypeString, Computed: true},
	// 			"domain_name":                          {Type: schema.TypeString, Computed: true},
	// 			"resource_url":                         {Type: schema.TypeString, Computed: true},
	// 			"api_secret":                           {Type: schema.TypeString, Computed: true},
	// 			"hosts":                                {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"tunnel_host":                          {Type: schema.TypeString, Computed: true},
	// 			"tunnel_port":                          {Type: schema.TypeString, Computed: true},
	// 			"tunnel_user":                          {Type: schema.TypeString, Computed: true},
	// 			"database":                             {Type: schema.TypeString, Computed: true},
	// 			"datasource":                           {Type: schema.TypeString, Computed: true},
	// 			"account":                              {Type: schema.TypeString, Computed: true},
	// 			"role":                                 {Type: schema.TypeString, Computed: true},
	// 			"email":                                {Type: schema.TypeString, Computed: true},
	// 			"account_id":                           {Type: schema.TypeString, Computed: true},
	// 			"server_url":                           {Type: schema.TypeString, Computed: true},
	// 			"user_key":                             {Type: schema.TypeString, Computed: true},
	// 			"api_version":                          {Type: schema.TypeString, Computed: true},
	// 			"daily_api_call_limit":                 {Type: schema.TypeString, Computed: true},
	// 			"time_zone":                            {Type: schema.TypeString, Computed: true},
	// 			"integration_key":                      {Type: schema.TypeString, Computed: true},
	// 			"advertisers":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"engagement_attribution_window":        {Type: schema.TypeString, Computed: true},
	// 			"conversion_report_time":               {Type: schema.TypeString, Computed: true},
	// 			"domain":                               {Type: schema.TypeString, Computed: true},
	// 			"update_method":                        {Type: schema.TypeString, Computed: true},
	// 			"replication_slot":                     {Type: schema.TypeString, Computed: true},
	// 			"publication_name":                     {Type: schema.TypeString, Computed: true},
	// 			"data_center":                          {Type: schema.TypeString, Computed: true},
	// 			"api_token":                            {Type: schema.TypeString, Computed: true},
	// 			"sub_domain":                           {Type: schema.TypeString, Computed: true},
	// 			"test_table_name":                      {Type: schema.TypeString, Computed: true},
	// 			"shop":                                 {Type: schema.TypeString, Computed: true},
	// 			"organizations":                        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"swipe_attribution_window":             {Type: schema.TypeString, Computed: true},
	// 			"api_access_token":                     {Type: schema.TypeString, Computed: true},
	// 			"account_ids":                          {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"sid":                                  {Type: schema.TypeString, Computed: true},
	// 			"secret":                               {Type: schema.TypeString, Computed: true},
	// 			"oauth_token":                          {Type: schema.TypeString, Computed: true},
	// 			"oauth_token_secret":                   {Type: schema.TypeString, Computed: true},
	// 			"consumer_key":                         {Type: schema.TypeString, Computed: true},
	// 			"consumer_secret":                      {Type: schema.TypeString, Computed: true},
	// 			"key":                                  {Type: schema.TypeString, Computed: true},
	// 			"advertisers_id":                       {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"sync_format":                          {Type: schema.TypeString, Computed: true},
	// 			"bucket_service":                       {Type: schema.TypeString, Computed: true},
	// 			"report_url":                           {Type: schema.TypeString, Computed: true},
	// 			"unique_id":                            {Type: schema.TypeString, Computed: true},
	// 			"auth_type":                            {Type: schema.TypeString, Computed: true},
	// 			"latest_version":                       {Type: schema.TypeString, Computed: true},
	// 			"authorization_method":                 {Type: schema.TypeString, Computed: true},
	// 			"service_version":                      {Type: schema.TypeString, Computed: true},
	// 			"last_synced_changes__utc_":            {Type: schema.TypeString, Computed: true},
	// 			"adobe_analytics_configurations": {Type: schema.TypeList, Computed: true,
	// 				Elem: &schema.Resource{
	// 					Schema: map[string]*schema.Schema{
	// 						"sync_mode":          {Type: schema.TypeString, Computed: true},
	// 						"report_suites":      {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"elements":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"metrics":            {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"calculated_metrics": {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 						"segments":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 					},
	// 				},
	// 			},
	// 			"is_new_package":                  {Type: schema.TypeString, Computed: true},
	// 			"is_multi_entity_feature_enabled": {Type: schema.TypeString, Computed: true},
	// 			"api_type":                        {Type: schema.TypeString, Computed: true},
	// 			"base_url":                        {Type: schema.TypeString, Computed: true},
	// 			"entity_id":                       {Type: schema.TypeString, Computed: true},
	// 			"soap_uri":                        {Type: schema.TypeString, Computed: true},
	// 			"user_id":                         {Type: schema.TypeString, Computed: true},
	// 			"encryption_key":                  {Type: schema.TypeString, Computed: true},
	// 			"always_encrypted":                {Type: schema.TypeString, Computed: true},
	// 			"eu_region":                       {Type: schema.TypeString, Computed: true},
	// 			"pat":                             {Type: schema.TypeString, Computed: true},
	// 			"token_key":                       {Type: schema.TypeString, Computed: true},
	// 			"token_secret":                    {Type: schema.TypeString, Computed: true},
	// 			"secrets_list": {Type: schema.TypeList, Computed: true,
	// 				Elem: &schema.Resource{
	// 					Schema: map[string]*schema.Schema{
	// 						"key":   {Type: schema.TypeString, Computed: true},
	// 						"value": {Type: schema.TypeString, Computed: true},
	// 					},
	// 				},
	// 			},
	// 			"pdb_name":             {Type: schema.TypeString, Computed: true},
	// 			"agent_host":           {Type: schema.TypeString, Computed: true},
	// 			"agent_port":           {Type: schema.TypeString, Computed: true},
	// 			"agent_user":           {Type: schema.TypeString, Computed: true},
	// 			"agent_password":       {Type: schema.TypeString, Computed: true},
	// 			"agent_public_cert":    {Type: schema.TypeString, Computed: true},
	// 			"agent_ora_home":       {Type: schema.TypeString, Computed: true},
	// 			"tns":                  {Type: schema.TypeString, Computed: true},
	// 			"use_oracle_rac":       {Type: schema.TypeString, Computed: true},
	// 			"asm_option":           {Type: schema.TypeString, Computed: true},
	// 			"is_single_table_mode": {Type: schema.TypeString, Computed: true},
	// 			"asm_user":             {Type: schema.TypeString, Computed: true},
	// 			"asm_password":         {Type: schema.TypeString, Computed: true},
	// 			"asm_oracle_home":      {Type: schema.TypeString, Computed: true},
	// 			"asm_tns":              {Type: schema.TypeString, Computed: true},
	// 			"sap_user":             {Type: schema.TypeString, Computed: true},
	// 			"organization":         {Type: schema.TypeString, Computed: true},
	// 			"packed_mode_tables":   {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"access_key":           {Type: schema.TypeString, Computed: true},
	// 			"domain_host_name":     {Type: schema.TypeString, Computed: true},
	// 			"client_name":          {Type: schema.TypeString, Computed: true},
	// 			"domain_type":          {Type: schema.TypeString, Computed: true},
	// 			"connection_method":    {Type: schema.TypeString, Computed: true},
	// 			"group_name":           {Type: schema.TypeString, Computed: true},
	// 			"company_id":           {Type: schema.TypeString, Computed: true},
	// 			"login_password":       {Type: schema.TypeString, Computed: true},
	// 			"environment":          {Type: schema.TypeString, Computed: true},
	// 			"properties":           {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
	// 			"is_public":            {Type: schema.TypeString, Computed: true},
	// 			"empty_header":         {Type: schema.TypeString, Computed: true},
	// 			"list_strategy":        {Type: schema.TypeString, Computed: true},
	// 		},
	// 	},
	// }
}

func getDataSourceProperties(path string) *map[string]*schema.Schema {
	shemasJson, err := gabs.ParseJSONFile("/Users/lukadevic/Fivetran/terraform-provider-fivetran/fivetran/schemas.json")
	if err != nil {
		panic(err)
	}

	properties := make(map[string]*schema.Schema)

	for key, child := range shemasJson.Path(path).ChildrenMap() {
		// introduce int, bool and maybe other types
		value := &schema.Schema{
			Type:     schema.TypeString,
			Computed: true}

		propertyType := child.Search("type").Data()

		switch propertyType {
		case "object":
			value = &schema.Schema{
				Type:     schema.TypeString,
				Computed: true}
		case "integer":
			value = &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true}
		case "boolean":
			value = &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true}
		case "array":
			itemType := child.Path("items.type").Data()

			childrenMap := child.Path("items.properties").ChildrenMap()

			if itemType == "object" && len(childrenMap) > 0 {
				childrenSchemaMap := make(map[string]*schema.Schema)

				for key2, child2 := range childrenMap {
					value2 := &schema.Schema{
						Type:     schema.TypeString,
						Computed: true}
					propertyType2 := child2.Search("type").Data()
					switch propertyType2 {
					case "object":
						value2 = &schema.Schema{
							Type:     schema.TypeString,
							Computed: true}
					case "integer":
						value2 = &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true}
					case "boolean":
						value2 = &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true}
					case "array":
						itemType2 := child2.Path("items.type").Data()

						if itemType2 == "string" || itemType2 == "object" {
							value2 = &schema.Schema{
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								}}
						}

						if itemType2 == "integer" {
							value2 = &schema.Schema{
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								}}
						}
					default:
						value2 = &schema.Schema{
							Type:     schema.TypeString,
							Computed: true}
					}

					childrenSchemaMap[key2] = value2
				}

				value = &schema.Schema{
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: childrenSchemaMap,
					},
				}
			} else if itemType == "string" || itemType == "object" {
				value = &schema.Schema{
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					}}
			}

			if itemType == "integer" {
				value = &schema.Schema{
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					}}
			}
		default:
			value = &schema.Schema{
				Type:     schema.TypeString,
				Computed: true}

		}
		properties[key] = value
	}

	return &properties
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

func dataSourceConnectorAutomaticReadStatusFlattenTasks(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func dataSourceConnectorAutomaticReadStatusFlattenWarnings(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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
func dataSourceConnectorAutomaticReadConfig(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	config := make([]interface{}, 1)

	configMap := make(map[string]interface{})

	c := resp.Data.CustomConfig

	m := structToMap(resp.Data.Config)
	for key, value := range m {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice && reflect.TypeOf(value).Elem().Kind() != reflect.String {
			fmt.Printf("Type of value is %T\n", value)

			var out []interface{}
			for i := 0; i < rv.Len(); i++ {
				out = append(out, rv.Index(i).Interface())
			}

			adb := structToMap(out[0])
			log.Output(1, intToStr(len(adb)))
			out[0] = adb
			c[key] = out
			continue
		}

		c[key] = value
	}

	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := "schemas." + service + ".properties.config.properties"
		newProperties := getDataSourceProperties(path)
		for k, v := range *newProperties {
			properties[k] = v
		}
	}

	for key, value := range properties {

		if key == "adobe_analytics_configurations" {
			log.Output(1, "LLL")
			fmt.Printf("Type of c[key] is %T\n", c[key])
		}
		if value.Type == schema.TypeSet || value.Type == schema.TypeList {

			if v, ok := c[key].([]string); ok {
				configMap[key] = xStrXInterface(v)
				continue
			}
			if v, ok := c[key].([]interface{}); ok {
				if v2, ok := v[0].(map[string]interface{}); ok {
					log.Output(2, intToStr(len(v2)))
					configMap[key] = v
				} else {
					configMap[key] = xInterfaceStrXStr(v)
				}
				continue
			}
		}
		if v, ok := c[key].(string); ok && v != "" {
			valueType := value.Type
			switch valueType {
			case schema.TypeBool:
				configMap[key] = strToBool(v)
			case schema.TypeInt:
				configMap[key] = strToInt(v)
			default:
				configMap[key] = v
			}
		}
	}
	config[0] = configMap

	return config
}

/*
This function will help you to convert your object from struct to map[string]interface{} based on your JSON tag in your structs.
Example how to use posted in sample_test.go file.
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

func dataSourceConnectorAutomaticReadConfigFlattenSecretsList(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	if len(resp.Data.Config.SecretsList) < 1 {
		return make([]interface{}, 0)
	}

	secretsList := make([]interface{}, len(resp.Data.Config.SecretsList))
	for i, v := range resp.Data.Config.SecretsList {
		s := make(map[string]interface{})
		mapAddStr(s, "key", v.Key)
		mapAddStr(s, "value", v.Value)
		secretsList[i] = s
	}

	return secretsList
}

func dataSourceConnectorAutomaticReadConfigFlattenProjectCredentials(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func dataSourceConnectorAutomaticReadConfigFlattenReports(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func dataSourceConnectorAutomaticReadConfigFlattenCustomTables(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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

func dataSourceConnectorAutomaticReadConfigFlattenAdobeAnalyticsConfigurations(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
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
