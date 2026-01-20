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
	Id                 types.String                  `tfsdk:"id"`
	ConnectionId       types.String                  `tfsdk:"connection_id"`
	Config             fivetrantypes.JsonConfigValue `tfsdk:"config"`
	Auth               fivetrantypes.JsonConfigValue `tfsdk:"auth"`
	RunSetupTests      types.Bool                    `tfsdk:"run_setup_tests"`
	TrustCertificates  types.Bool                    `tfsdk:"trust_certificates"`
	TrustFingerprints  types.Bool                    `tfsdk:"trust_fingerprints"`
}

func (d *ConnectionConfigModel) Validate(ctx context.Context, client *fivetran.Client) (map[string]interface{}, map[string]interface{}, error) {
	svc := client.NewConnectionDetails()
	svc.ConnectionID(d.ConnectionId.ValueString())
	_, err := svc.Do(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("connection %s not found: %w", d.ConnectionId.ValueString(), err)
	}

	var configMap map[string]interface{}
	hasConfig := !d.Config.IsNull() && !d.Config.IsUnknown() && d.Config.ValueString() != ""
	if hasConfig {
		if err := json.Unmarshal([]byte(d.Config.ValueString()), &configMap); err != nil {
			return nil, nil, fmt.Errorf("invalid config JSON: %w", err)
		}
	}

	var authMap map[string]interface{}
	hasAuth := !d.Auth.IsNull() && !d.Auth.IsUnknown() && d.Auth.ValueString() != ""
	if hasAuth {
		if err := json.Unmarshal([]byte(d.Auth.ValueString()), &authMap); err != nil {
			return nil, nil, fmt.Errorf("invalid auth JSON: %w", err)
		}
	}

	if !hasConfig && !hasAuth {
		return nil, nil, fmt.Errorf("at least one of 'config' or 'auth' must be specified")
	}

	return configMap, authMap, nil
}
