package fivetran

import (
    "context"
    "fmt"

    fivetran "github.com/fivetran/go-fivetran"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeamConnectorMembership() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceTeamConnectorMembershipCreate,
        ReadContext:   resourceTeamConnectorMembershipRead,
        UpdateContext: resourceTeamConnectorMembershipUpdate,
        DeleteContext: resourceTeamConnectorMembershipDelete,
        Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
        Schema:        resourceTeamConnectorMembershipBase(false),
    }
}

func resourceTeamConnectorMembershipBase(datasource bool) map[string]*schema.Schema {
    return map[string]*schema.Schema{
        "id": {
            Type:        schema.TypeString,
            Computed:    true,
            Description: "The unique identifier for resource.",
        },
        "team_id": {
            Type:        schema.TypeString,
            Required:    true,
            Description: "The unique identifier for the team within your account.",
        },
        "connector": resourceTeamConnectorMembershipBaseConnectors(datasource),
    }
}

func resourceTeamConnectorMembershipBaseConnectors(datasource bool) *schema.Schema {
    return &schema.Schema{
        Type: schema.TypeSet, 
        Optional: true, 
        Elem: &schema.Resource{
            Schema: map[string]*schema.Schema{
                "connector_id": {
                    Type:        schema.TypeString,
                    Required:    true,
                    ForceNew:    !datasource,
                    Description: "The connector unique identifier",
                },
                "role": {
                    Type:        schema.TypeString,
                    Required:    !datasource,
                    Computed:    datasource,
                    Description: "The team's role that links the team and the connector",
                },
                "created_at": {
                    Type:        schema.TypeString,
                    Computed:    true,
                    Description: "The date and time the membership was created",
                },
            },
        },
    }
}

func resourceTeamConnectorMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)

    if err := resourceTeamConnectorMembershipSyncConnectors(client, d.Get("connector").(*schema.Set).List(), teamId, ctx); err != nil {
        return newDiagAppend(diags, diag.Error, "create error: resourceTeamConnectorMembershipSyncConnectors", fmt.Sprint(err))
    }

    d.SetId(teamId)

    resourceTeamConnectorMembershipRead(ctx, d, m)

    return diags
}

func resourceTeamConnectorMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)

    resp, err  := dataSourceTeamConnectorMembershipsGet(client, ctx, teamId)
    if err != nil {
        if resp.Code == "404" {
            d.SetId("")
            return nil
        }
        return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }

    // msi stands for Map String Interface
    msi := make(map[string]interface{})
    msi["team_id"] = teamId

    var connectors []interface{}
    for _, connector := range resp.Data.Items {
        if connector.Role == "" {
            continue
        }
        con := make(map[string]interface{})
        con["connector_id"] = connector.ConnectorId
        con["role"] = connector.Role
        con["created_at"] = connector.CreatedAt
        connectors = append(connectors, con)
    }

    msi["connector"] = connectors

    for k, v := range msi {
        if err := d.Set(k, v); err != nil {
            return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
        }
    }

    d.SetId(teamId)

    return diags
}

func resourceTeamConnectorMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)

    if d.HasChange("connector") {
        if err := resourceTeamConnectorMembershipSyncConnectors(client, d.Get("connector").(*schema.Set).List(), teamId, ctx); err != nil {
            return newDiagAppend(diags, diag.Error, "read error: resourceTeamConnectorMembershipSyncConnectors", fmt.Sprint(err))
        }
    }
    
    return resourceTeamConnectorMembershipRead(ctx, d, m)
}

func resourceTeamConnectorMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)
    
    for _, v := range d.Get("connector").(*schema.Set).List() {
        if resp, err := client.NewTeamConnectorMembershipDelete().TeamId(teamId).ConnectorId(v.(map[string]interface{})["connector_id"].(string)).Do(ctx); err != nil {
            return newDiagAppend(diags, diag.Error, "set error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
        }
    }

    return diags
}

func resourceTeamConnectorMembershipSyncConnectors(client *fivetran.Client, connectors []interface{}, teamId string, ctx context.Context) error {
    responseConnectors, err := dataSourceTeamConnectorMembershipsGet(client, ctx, teamId)
    if err != nil {
        return fmt.Errorf("read error: dataSourceTeamConnectorMembershipsGet %v; code: %v", err, responseConnectors.Code)
    }

    localConnectors := make(map[string]interface{})
    for _, v := range connectors {
        localConnectors[v.(map[string]interface{})["connector_id"].(string)] = v.(map[string]interface{})["role"].(string)
    }

    remoteConnectors := make(map[string]interface{})
    for _, v := range responseConnectors.Data.Items {
        remoteConnectors[v.ConnectorId] = v.Role
    }

    for remoteKey, remoteValue := range remoteConnectors {
        role, found := localConnectors[remoteKey]

        if !found {
            if resp, err := client.NewTeamConnectorMembershipDelete().TeamId(teamId).ConnectorId(remoteKey).Do(ctx); err != nil {
                return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
            }
        } else if role.(string) != remoteValue {
            if resp, err := client.NewTeamConnectorMembershipModify().TeamId(teamId).ConnectorId(remoteKey).Role(role.(string)).Do(ctx); err != nil {
                return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
            }
        }
    }


    for localKey, localValue := range localConnectors {
        _, exists := remoteConnectors[localKey]

        if !exists {
            if resp, err := client.NewTeamConnectorMembershipCreate().TeamId(teamId).ConnectorId(localKey).Role(localValue.(string)).Do(ctx); err != nil {
                return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
            }
        }
    }

    return nil
}