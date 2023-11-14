package schema

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSchemaConfigNew() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: schemaConfigNewCreate,
		ReadWithoutTimeout:   schemaConfigNewRead,
		UpdateWithoutTimeout: schemaConfigNewUpdate,
		DeleteContext:        schemaConfigNewDelete,
		Importer:             &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema:               rootSchema(),
		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(2 * time.Hour), // Import operation can trigger schema reload
			Create: schema.DefaultTimeout(2 * time.Hour),
			Update: schema.DefaultTimeout(2 * time.Hour),
		},
	}
}

func schemaConfigNewCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorID := d.Get(CONNECTOR_ID).(string)
	client := m.(*fivetran.Client)
	var schemaChangeHandling = d.Get(SCHEMA_CHANGE_HANDLING).(string)

	ctx, cancel := helpers.SetContextTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		if schemaResponse.Code != "NotFound_SchemaConfig" {
			return helpers.NewDiagAppend(diags, diag.Error, "create error",
				fmt.Sprintf("Error wile retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message))
		} else {
			// Reload schema: we can't update schema if connector doesn't have it yet.
			excludeMode := "PRESERVE"

			// If we want to disable everything by default - we can do it in schema reload
			if schemaChangeHandling == BLOCK_ALL {
				excludeMode = "EXCLUDE"
			}

			schemaResponse, err = client.NewConnectorSchemaReload().ExcludeMode(excludeMode).ConnectorID(connectorID).Do(ctx)
			if err != nil {
				return helpers.NewDiagAppend(diags, diag.Error, "create error",
					fmt.Sprintf("Error wile schema reload. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message))
			}
		}
	}
	svc := client.NewConnectorSchemaUpdateService().ConnectorID(connectorID)
	if schemaChangeHandling != schemaResponse.Data.SchemaChangeHandling {
		svc.SchemaChangeHandling(schemaChangeHandling)
		schResponse, err := svc.Do(ctx)
		if err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "create error",
				fmt.Sprintf("Error wile applying schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message))
		}

		// We need to re-read schema
		schemaResponse, err = client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
		if err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "create error",
				fmt.Sprintf("Error while reading schema after schema change handling apply. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message))
		}
	}

	// read upstream config
	config := _config{}
	config.readFromResponse(schemaResponse)

	// read local config
	localConfig := _config{}
	localConfig.readFromResourceData(d)

	// apply local config, managing upstream config according to schema change handling policy
	err = config.override(&localConfig, schemaChangeHandling)

	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "create error",
			fmt.Sprintf("Error wile applying schema config patch. %v;", err))
	}

	if config.hasUpdates() {
		// applying patch
		svc = config.prepareRequest(client.NewConnectorSchemaUpdateService())
		svc.ConnectorID(connectorID)
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		schemaResponse, err = svc.Do(ctx)

		if err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "create error",
				fmt.Sprintf("Error wile applying schema config patch. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message))
		}
	}

	// set state from effective config
	for k, v := range config.toStateObject(schemaChangeHandling, localConfig) {
		if err := d.Set(k, v); err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(connectorID)

	return diags
}

func schemaConfigNewRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorID := d.Get(CONNECTOR_ID).(string)
	client := m.(*fivetran.Client)

	ctx, cancel := helpers.SetContextTimeout(ctx, d.Timeout(schema.TimeoutRead))
	defer cancel()

	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "read error",
			fmt.Sprintf("Error wile retrieving existing schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message))
	}

	// read upstream config
	config := _config{}
	config.readFromResponse(schemaResponse)

	localConfig := _config{}
	localConfig.readFromResourceData(d)

	// set state from effective config
	for k, v := range config.toStateObject(schemaResponse.Data.SchemaChangeHandling, localConfig) {
		if err := d.Set(k, v); err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func schemaConfigNewUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	connectorID := d.Get(CONNECTOR_ID).(string)
	client := m.(*fivetran.Client)

	ctx, cancel := helpers.SetContextTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	if d.HasChange(SCHEMA_CHANGE_HANDLING) {
		svc := client.NewConnectorSchemaUpdateService().ConnectorID(connectorID)
		svc.SchemaChangeHandling(d.Get(SCHEMA_CHANGE_HANDLING).(string))
		schResponse, err := svc.Do(ctx)
		if err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "update error",
				fmt.Sprintf("Error while applying schema change handling policy. %v; code: %v; message: %v", err, schResponse.Code, schResponse.Message))
		}
	}

	schemaResponse, err := client.NewConnectorSchemaDetails().ConnectorID(connectorID).Do(ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "update error",
			fmt.Sprintf("Error while reading upstream schema. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message))
	}

	// read upstream config
	config := _config{}
	config.readFromResponse(schemaResponse)

	// read local config
	localConfig := _config{}
	localConfig.readFromResourceData(d)

	// apply local config, managing upstream config according to schema change handling policy
	err = config.override(&localConfig, d.Get(SCHEMA_CHANGE_HANDLING).(string))

	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "create error",
			fmt.Sprintf("Error wile applying schema config patch. %v;", err))
	}

	if config.hasUpdates() {
		// applying patch
		svc := config.prepareRequest(client.NewConnectorSchemaUpdateService())
		svc.ConnectorID(connectorID)
		// we should not parse response here because it will contain only applied diffs, not the whole configuration
		schemaResponse, err = svc.Do(ctx)

		if err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "update error",
				fmt.Sprintf("Error wile applying schema config patch. %v; code: %v; message: %v", err, schemaResponse.Code, schemaResponse.Message))
		}
	}

	// set state from effective config
	for k, v := range config.toStateObject(schemaResponse.Data.SchemaChangeHandling, localConfig) {
		if err := d.Set(k, v); err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func schemaConfigNewDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// do nothing - we can't delete schema settings
	return diags
}
