package actions

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-framework/action"
	actionSchema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
)

const (
	defaultPollInterval = 30 * time.Second
	defaultTimeout      = 20 * time.Minute
	maxTimeout          = 1 * time.Hour
)

// PollIntervalOverride can be set in tests to use a shorter poll interval.
// When zero (default), defaultPollInterval is used.
var PollIntervalOverride time.Duration

func ConnectionSchemaReload() action.Action {
	interval := defaultPollInterval
	if PollIntervalOverride > 0 {
		interval = PollIntervalOverride
	}
	return &connectionSchemaReload{
		pollInterval: interval,
	}
}

type connectionSchemaReload struct {
	client       *fivetran.Client
	pollInterval time.Duration
}

type connectionSchemaReloadModel struct {
	ConnectionId types.String   `tfsdk:"connection_id"`
	ExcludeMode  types.String   `tfsdk:"exclude_mode"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

var _ action.ActionWithConfigure = &connectionSchemaReload{}
var _ action.ActionWithValidateConfig = &connectionSchemaReload{}

func (a *connectionSchemaReload) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_reload"
}

func (a *connectionSchemaReload) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = actionSchema.Schema{
		Description: "Triggers a schema reload for a Fivetran connection.",
		Attributes: map[string]actionSchema.Attribute{
			"connection_id": actionSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the connection to reload the schema for.",
			},
			"exclude_mode": actionSchema.StringAttribute{
				Optional:    true,
				Description: "The exclude mode for the schema reload. Accepted values: PRESERVE, EXCLUDE. Default: PRESERVE.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (a *connectionSchemaReload) ValidateConfig(ctx context.Context, req action.ValidateConfigRequest, resp *action.ValidateConfigResponse) {
	var data connectionSchemaReloadModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := data.Timeouts.Invoke(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if timeout > maxTimeout {
		resp.Diagnostics.AddError(
			"Invalid Timeout",
			fmt.Sprintf("Invoke timeout must not exceed %v.", maxTimeout),
		)
	}
}

func (a *connectionSchemaReload) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*fivetran.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *fivetran.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	a.client = client
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}

func (a *connectionSchemaReload) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	if a.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data connectionSchemaReloadModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	excludeMode := "PRESERVE"
	if !data.ExcludeMode.IsNull() && !data.ExcludeMode.IsUnknown() {
		excludeMode = data.ExcludeMode.ValueString()
	}

	timeout, diags := data.Timeouts.Invoke(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if timeout > maxTimeout {
		timeout = maxTimeout
	}

	connectionID := data.ConnectionId.ValueString()

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Reloading schema for connection %s...", connectionID),
		})
	}

	reloadResponse, err := a.client.NewConnectionSchemaReload().
		ConnectionID(connectionID).
		ExcludeMode(excludeMode).
		Do(ctx)
	if err == nil {
		return
	}

	if !isTimeoutError(err) {
		resp.Diagnostics.AddError(
			"Unable to Reload Connection Schema",
			fmt.Sprintf("Error reloading schema for connection %s: %v; code: %v; message: %v",
				connectionID, err, reloadResponse.Code, reloadResponse.Message),
		)
		return
	}

	// The reload request timed out. This doesn't mean it failed — the server
	// may still be processing. Poll NewConnectionSchemaDetails until we get a
	// successful response or exhaust the polling budget.
	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Schema reload request for connection %s timed out. Polling for completion...", connectionID),
		})
	}

	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			resp.Diagnostics.AddError(
				"Schema Reload Timed Out",
				fmt.Sprintf("Schema reload for connection %s did not complete within %v. "+
					"The reload may still be in progress on the server.", connectionID, timeout),
			)
			return
		}

		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError(
				"Schema Reload Cancelled",
				fmt.Sprintf("Context cancelled while waiting for schema reload on connection %s: %v",
					connectionID, ctx.Err()),
			)
			return
		case <-time.After(a.pollInterval):
		}

		if resp.SendProgress != nil {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Checking schema status for connection %s...", connectionID),
			})
		}

		detailsResponse, detailsErr := a.client.NewConnectionSchemaDetails().
			ConnectionID(connectionID).
			Do(ctx)
		if detailsErr == nil && detailsResponse.Code != "NotFound_SchemaConfig" {
			return
		}
	}
}
