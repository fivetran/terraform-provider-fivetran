package fivetran

import (
    "context"
    "fmt"

    fivetran "github.com/fivetran/go-fivetran"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeamUserMembership() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceTeamUserMembershipCreate,
        ReadContext:   resourceTeamUserMembershipRead,
        UpdateContext: resourceTeamUserMembershipUpdate,
        DeleteContext: resourceTeamUserMembershipDelete,
        Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
        Schema:        resourceTeamUserMembershipBase(false),
    }
}

func resourceTeamUserMembershipBase(datasource bool) map[string]*schema.Schema {
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
        "user": resourceTeamUserMembershipBaseUsers(datasource),
    }
}

func resourceTeamUserMembershipBaseUsers(datasource bool) *schema.Schema {
    return &schema.Schema{
        Type: schema.TypeSet, 
        Optional: true, 
        Elem: &schema.Resource{
            Schema: map[string]*schema.Schema{
                "user_id": {
                    Type:        schema.TypeString,
                    Required:    true,
                    ForceNew:    !datasource,
                    Description: "The user unique identifier",
                },
                "role": {
                    Type:        schema.TypeString,
                    Required:    !datasource,
                    Computed:    datasource,
                    Description: "The team's role that links the team and the user",
                },
            },
        },
    }
}

func resourceTeamUserMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)

    if err := resourceTeamUserMembershipSyncUsers(client, d.Get("user").(*schema.Set).List(), teamId, ctx); err != nil {
        return newDiagAppend(diags, diag.Error, "create error: resourceTeamUserMembershipSyncUsers", fmt.Sprint(err))
    }

    d.SetId(teamId)

    resourceTeamUserMembershipRead(ctx, d, m)

    return diags
}

func resourceTeamUserMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)

    resp, err  := dataSourceTeamUserMembershipsGet(client, ctx, teamId)
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

    var users []interface{}
    for _, user := range resp.Data.Items {
        if user.Role == "" {
            continue
        }
        con := make(map[string]interface{})
        con["user_id"] = user.UserId
        con["role"] = user.Role
        users = append(users, con)
    }

    msi["user"] = users

    for k, v := range msi {
        if err := d.Set(k, v); err != nil {
            return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
        }
    }

    d.SetId(teamId)

    return diags
}

func resourceTeamUserMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)

    if d.HasChange("user") {
        if err := resourceTeamUserMembershipSyncUsers(client, d.Get("user").(*schema.Set).List(), teamId, ctx); err != nil {
            return newDiagAppend(diags, diag.Error, "read error: resourceTeamUserMembershipSyncUsers", fmt.Sprint(err))
        }
    }
    
    return resourceTeamUserMembershipRead(ctx, d, m)
}

func resourceTeamUserMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    teamId := d.Get("team_id").(string)
    
    for _, v := range d.Get("user").(*schema.Set).List() {
        if resp, err := client.NewTeamUserMembershipDelete().TeamId(teamId).UserId(v.(map[string]interface{})["user_id"].(string)).Do(ctx); err != nil {
            return newDiagAppend(diags, diag.Error, "set error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
        }
    }

    return diags
}

func resourceTeamUserMembershipSyncUsers(client *fivetran.Client, users []interface{}, teamId string, ctx context.Context) error {
    responseUsers, err := dataSourceTeamUserMembershipsGet(client, ctx, teamId)
    if err != nil {
        return fmt.Errorf("read error: dataSourceTeamUserMembershipsGet %v; code: %v", err, responseUsers.Code)
    }

    localUsers := make(map[string]interface{})
    for _, v := range users {
        localUsers[v.(map[string]interface{})["user_id"].(string)] = v.(map[string]interface{})["role"].(string)
    }

    remoteUsers := make(map[string]interface{})
    for _, v := range responseUsers.Data.Items {
        remoteUsers[v.UserId] = v.Role
    }

    for remoteKey, remoteValue := range remoteUsers {
        role, found := localUsers[remoteKey]

        if !found {
            if resp, err := client.NewTeamUserMembershipDelete().TeamId(teamId).UserId(remoteKey).Do(ctx); err != nil {
                return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
            }
        } else if role.(string) != remoteValue {
            if resp, err := client.NewTeamUserMembershipModify().TeamId(teamId).UserId(remoteKey).Role(role.(string)).Do(ctx); err != nil {
                return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
            }
        }
    }


    for localKey, localValue := range localUsers {
        _, exists := remoteUsers[localKey]

        if !exists {
            if resp, err := client.NewTeamUserMembershipCreate().TeamId(teamId).UserId(localKey).Role(localValue.(string)).Do(ctx); err != nil {
                return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
            }
        }
    }

    return nil
}