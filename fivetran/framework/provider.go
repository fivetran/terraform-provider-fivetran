package framework

import (
	"context"

	"os"

	"github.com/fivetran/go-fivetran"
	httputils "github.com/fivetran/go-fivetran/http_utils"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/datasources"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const Version = "1.4.3" // Current provider version

type fivetranProvider struct {
	mockClient httputils.HttpClient
}

type fivetranProviderModel struct {
	ApiKey    types.String `tfsdk:"api_key"`
	ApiSecret types.String `tfsdk:"api_secret"`
	ApiUrl    types.String `tfsdk:"api_url"`
}

func FivetranProvider() provider.Provider {
	common.LoadConfigFieldsMap()
	common.LoadAuthFieldsMap()
	common.LocaDestinationFieldsMap()
	return &fivetranProvider{mockClient: nil}
}

// For mocked tests
func FivetranProviderMock(client httputils.HttpClient) provider.Provider {
	common.LoadConfigFieldsMap()
	common.LoadAuthFieldsMap()
	common.LocaDestinationFieldsMap()
	return &fivetranProvider{mockClient: client}
}

func (p *fivetranProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fivetran"
}

func (p *fivetranProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key":    schema.StringAttribute{Optional: true},
			"api_secret": schema.StringAttribute{Optional: true, Sensitive: true},
			"api_url":    schema.StringAttribute{Optional: true},
		},
	}
}

func (p *fivetranProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Check environment variables
	apiKey := os.Getenv("FIVETRAN_APIKEY")
	apiSecret := os.Getenv("FIVETRAN_APISECRET")
	apiUrl := os.Getenv("FIVETRAN_API_URL")

	var data fivetranProviderModel

	// Read configuration data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.ApiKey.ValueString() != "" {
		apiKey = data.ApiKey.ValueString()
	}
	if data.ApiSecret.ValueString() != "" {
		apiSecret = data.ApiSecret.ValueString()
	}
	if data.ApiUrl.ValueString() != "" {
		apiUrl = data.ApiUrl.ValueString()
	}

	// Init client
	fivetranClient := fivetran.New(apiKey, apiSecret)
	if apiUrl != "" {
		fivetranClient.BaseURL(apiUrl)
	}

	// Set mocked http client for tests
	if p.mockClient != nil {
		fivetranClient.SetHttpClient(p.mockClient)
	}

	fivetranClient.CustomUserAgent("terraform-provider-fivetran/" + Version)
	resp.DataSourceData = fivetranClient
	resp.ResourceData = fivetranClient
}

func (p *fivetranProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.User,
		resources.UserConnectorMembership,
		resources.UserGroupMembership,
		resources.Webhook,
		resources.Connector,
		resources.ConnectorSchema,
		resources.ConnectorSchedule,
		resources.Destination,
		resources.Team,
		resources.TeamConnectorMembership,
		resources.TeamGroupMembership,
		resources.TeamUserMembership,
		resources.ExternalLogging,
		resources.Group,
		resources.GroupUser,
		resources.ConnectorFingerprint,
		resources.ConnectorCertificate,
		resources.DestinationFingerprint,
		resources.DestinationCertificate,
		resources.DbtProject,
		resources.DbtTransformation,
		resources.ProxyAgent,
		resources.LocalProcessingAgent,
		resources.HybridDeploymentAgent,
		resources.DbtGitProjectConfig,
		resources.PrivateLink,
		resources.TransformationProject,
		resources.Transformation,
	}
}

func (p *fivetranProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.User,
		datasources.Users,
		datasources.UserConnectorMemberships,
		datasources.UserGroupMemberships,
		datasources.Webhook,
		datasources.Webhooks,
		datasources.GroupSshKey,
		datasources.GroupServiceAccount,
		datasources.Connector,
		datasources.Destination,
		datasources.Team,
		datasources.Teams,
		datasources.TeamConnectorMemberships,
		datasources.TeamGroupMemberships,
		datasources.TeamUserMemberships,
		datasources.ExternalLogging,
		datasources.Roles,
		datasources.Group,
		datasources.Groups,
		datasources.GroupConnectors,
		datasources.GroupUsers,
		datasources.ConnectorsMetadata,
		datasources.ConnectorFingerprints,
		datasources.ConnectorCertificates,
		datasources.DestinationFingerprints,
		datasources.DestinationCertificates,
		datasources.DbtModels,
		datasources.DbtProject,
		datasources.DbtProjects,
		datasources.DbtTransformation,
		datasources.ProxyAgent,
		datasources.ProxyAgents,
		datasources.LocalProcessingAgent,
		datasources.LocalProcessingAgents,
		datasources.PrivateLink,
		datasources.PrivateLinks,
		datasources.HybridDeploymentAgent,
		datasources.HybridDeploymentAgents,
		datasources.Connectors,
		datasources.Destinations,
		datasources.ExternalLogs,
		datasources.QuickstartPackage,
		datasources.QuickstartPackages,
		datasources.TransformationProject,
		datasources.TransformationProjects,
		datasources.Transformation,
		datasources.Transformations,
	}
}
