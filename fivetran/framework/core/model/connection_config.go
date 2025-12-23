package model

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectionConfigModel struct {
	Id           types.String                  `tfsdk:"id"`
	ConnectionId types.String                  `tfsdk:"connection_id"`
	Config       fivetrantypes.JsonConfigValue `tfsdk:"config"`
	Auth         fivetrantypes.JsonConfigValue `tfsdk:"auth"`
}

func (d *ConnectionConfigModel) Validate(ctx context.Context, client *fivetran.Client) (map[string]interface{}, map[string]interface{}, error) {
	svc := client.NewConnectionDetails()
	svc.ConnectionID(d.ConnectionId.ValueString())
	_, err := svc.Do(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("connection %s not found: %w", d.ConnectionId.ValueString(), err)
	}

	var configMap map[string]interface{}
	if !d.Config.IsNull() && !d.Config.IsUnknown() && d.Config.ValueString() != "" {
		if err := json.Unmarshal([]byte(d.Config.ValueString()), &configMap); err != nil {
			return nil, nil, fmt.Errorf("invalid config JSON: %w", err)
		}
	}

	var authMap map[string]interface{}
	if !d.Auth.IsNull() && !d.Auth.IsUnknown() && d.Auth.ValueString() != "" {
		if err := json.Unmarshal([]byte(d.Auth.ValueString()), &authMap); err != nil {
			return nil, nil, fmt.Errorf("invalid auth JSON: %w", err)
		}
	}

	return configMap, authMap, nil
}
