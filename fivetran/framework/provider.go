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

const Version = "1.1.6" // Current provider version

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
	return &fivetranProvider{mockClient: nil}
}

// For mocked tests
func FivetranProviderMock(client httputils.HttpClient) provider.Provider {
	common.LoadConfigFieldsMap()
	common.LoadAuthFieldsMap()
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
		resources.Connector,
		resources.ConnectorSchema,
		resources.ConnectorSchedule,
	}
}

func (p *fivetranProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.User,
		datasources.GroupSshKey,
		datasources.GroupServiceAccount,
		datasources.Connector,
	}
}
