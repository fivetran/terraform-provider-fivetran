package model

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/connections"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
)

type ConnectionConfigModel struct {
    ConnectionId            types.String                    `tfsdk:"connection_id"`
    Config                  fivetrantypes.JsonConfigValue   `tfsdk:"config"`
    Auth                    fivetrantypes.JsonConfigValue   `tfsdk:"auth"`
}

func (d *ConnectionConfigModel) ReadFromResponse(resp connections.DetailsWithCustomConfigNoTestsResponse) {
    d.ConnectionId = types.StringValue(resp.Data.ID)

    resultRawString, _ := json.Marshal(resp.Data.Config)
    d.Config = fivetrantypes.NewJsonConfigValue(string(resultRawString))
}

func (d *ConnectionConfigModel) ReadFromUpdateResponse(resp connections.DetailsWithCustomConfigResponse) {
    d.ConnectionId = types.StringValue(resp.Data.ID)

    resultRawString, _ := json.Marshal(resp.Data.Config)
    d.Config = fivetrantypes.NewJsonConfigValue(string(resultRawString))
}

func (d *ConnectionConfigModel) Validate(ctx context.Context, client *fivetran.Client) (map[string]interface{}, map[string]interface{}, error) {
    svc := client.NewConnectionDetails()
    svc.ConnectionID(d.ConnectionId.ValueString())
    connection, err := svc.Do(ctx)
    if err != nil {
        return nil, nil, err
    }

    svcMetadata := client.NewMetadataDetails()
    svcMetadata.Service(connection.Data.Service)
    response, err := svcMetadata.Do(ctx)
    if err != nil {
        return nil, nil, err
    }

    /* validation logic */
    fmt.Println(response)
    configMap := make(map[string]interface{})
    json.Unmarshal([]byte(d.Config.ValueString()), &configMap)

    authMap := make(map[string]interface{})
    json.Unmarshal([]byte(d.Auth.ValueString()), &authMap)

    return configMap, authMap, nil
}