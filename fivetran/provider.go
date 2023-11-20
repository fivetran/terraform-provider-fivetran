package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran"
	connector_schema "github.com/fivetran/terraform-provider-fivetran/modules/connector/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var limit = 1000        // REST API response objects limit per HTTP request
const Version = "1.1.4" // Current provider version

func Provider() *schema.Provider {
	var resourceMap = map[string]*schema.Resource{
		"fivetran_group":                     resourceGroup(),
		"fivetran_group_users":               resourceGroupUsers(),
		"fivetran_destination":               resourceDestination(),
		"fivetran_connector":                 resourceConnector(),
		"fivetran_connector_schedule":        resourceConnectorSchedule(),
		"fivetran_connector_schema_config":   connector_schema.ResourceSchemaConfigNew(),
		"fivetran_dbt_transformation":        resourceDbtTransformation(),
		"fivetran_dbt_project":               resourceDbtProject(),
		"fivetran_webhook":                   resourceWebhook(),
		"fivetran_external_logging":          resourceExternalLogging(),
		"fivetran_team":                      resourceTeam(),
		"fivetran_team_connector_membership": resourceTeamConnectorMembership(),
		"fivetran_team_group_membership":     resourceTeamGroupMembership(),
		"fivetran_team_user_membership":      resourceTeamUserMembership(),
		"fivetran_connector_fingerprints":    resourceFingerprints(Connector),
		"fivetran_destination_fingerprints":  resourceFingerprints(Destination),
		"fivetran_connector_certificates":    resourceCertificates(Connector),
		"fivetran_destination_certificates":  resourceCertificates(Destination),
	}

	var dataSourceMap = map[string]*schema.Resource{
		"fivetran_users":                      dataSourceUsers(),
		"fivetran_group":                      dataSourceGroup(),
		"fivetran_groups":                     dataSourceGroups(),
		"fivetran_group_connectors":           dataSourceGroupConnectors(),
		"fivetran_group_users":                dataSourceGroupUsers(),
		"fivetran_destination":                dataSourceDestination(),
		"fivetran_connectors_metadata":        dataSourceConnectorsMetadata(),
		"fivetran_connector":                  dataSourceConnector(),
		"fivetran_dbt_transformation":         dataSourceDbtTransformation(),
		"fivetran_dbt_project":                dataSourceDbtProject(),
		"fivetran_dbt_projects":               dataSourceDbtProjects(),
		"fivetran_dbt_models":                 dataSourceDbtModels(),
		"fivetran_webhook":                    dataSourceWebhook(),
		"fivetran_webhooks":                   dataSourceWebhooks(),
		"fivetran_external_logging":           dataSourceExternalLogging(),
		"fivetran_roles":                      dataSourceRoles(),
		"fivetran_metadata_schemas":           dataSourceMetadataSchemas(),
		"fivetran_metadata_tables":            dataSourceMetadataTables(),
		"fivetran_metadata_columns":           dataSourceMetadataColumns(),
		"fivetran_team":                       dataSourceTeam(),
		"fivetran_teams":                      dataSourceTeams(),
		"fivetran_team_connector_memberships": dataSourceTeamConnectorMemberships(),
		"fivetran_team_group_memberships":     dataSourceTeamGroupMemberships(),
		"fivetran_team_user_memberships":      dataSourceTeamUserMemberships(),
		"fivetran_connector_fingerprints":     dataSourceFingerprints(Connector),
		"fivetran_destination_fingerprints":   dataSourceFingerprints(Destination),
		"fivetran_connector_certificates":     dataSourceCertificates(Connector),
		"fivetran_destination_certificates":   dataSourceCertificates(Destination),
	}

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key":    {Type: schema.TypeString, Optional: true},
			"api_secret": {Type: schema.TypeString, Optional: true, Sensitive: true},
			"api_url":    {Type: schema.TypeString, Optional: true},
		},
		ResourcesMap:         resourceMap,
		DataSourcesMap:       dataSourceMap,
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	if d.Get("api_key") == "" {
		apiKey, _ := schema.EnvDefaultFunc("FIVETRAN_APIKEY", nil)()
		d.Set("api_key", apiKey)
	}
	if d.Get("api_secret") == "" {
		apiSecret, _ := schema.EnvDefaultFunc("FIVETRAN_APISECRET", nil)()
		d.Set("api_secret", apiSecret)
	}
	if d.Get("api_url") == "" {
		apiUrl, _ := schema.EnvDefaultFunc("FIVETRAN_APIURL", nil)()
		d.Set("api_url", apiUrl)
	}

	fivetranClient := fivetran.New(d.Get("api_key").(string), d.Get("api_secret").(string))
	if d.Get("api_url") != "" {
		fivetranClient.BaseURL(d.Get("api_url").(string))
	}

	fivetranClient.CustomUserAgent("terraform-provider-fivetran/" + Version)
	return fivetranClient, diag.Diagnostics{}
}
