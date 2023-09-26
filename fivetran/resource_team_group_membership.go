package fivetran

import (
    "context"
    "fmt"

    fivetran "github.com/fivetran/go-fivetran"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeamGroupMembership() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceTeamGroupMembershipCreate,
        ReadContext:   resourceTeamGroupMembershipRead,
        UpdateContext: resourceTeamGroupMembershipUpdate,
        DeleteContext: resourceTeamGroupMembershipDelete,
        Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
        Schema:        getTeamGroupMembershipSchema(false),
    }
}

func getTeamGroupMembershipSchema(datasource bool) map[string]*schema.Schema {
    return map[string]*schema.Schema{
        "id": {
            Type:        schema.TypeString,
            Optional:    true,
            Computed:    true,
            Description: "Fake record Id, compile from team_id and connector_id",
        },
        "team_id": {
            Type:        schema.TypeString,
            Required:    true,
            ForceNew:    !datasource,
            Description: "The unique identifier for the team within your account.",
        },
        "group_id": {
            Type:        schema.TypeString,
            Required:    true,
            ForceNew:    !datasource,
            Description: "The Group unique identifier",
        },
        "role": {
            Type:        schema.TypeString,
            Required:    !datasource,
            Computed:    datasource,
            Description: "The team's role that links the team and the Group",
        },
        "created_at": {
            Type:        schema.TypeString,
            Computed:    true,
            Description: "The date and time the membership was created",
        },
    }
}

func resourceTeamGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)

    svc := client.NewTeamGroupMembershipCreate()
    svc.TeamId(d.Get("team_id").(string))
    svc.GroupId(d.Get("group_id").(string))
    svc.Role(d.Get("role").(string))

    resp, err := svc.Do(ctx)
    if err != nil {
        return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }

    d.SetId(d.Get("team_id").(string) + "|" + d.Get("group_id").(string))

    resourceTeamGroupMembershipRead(ctx, d, m)

    return diags
}

func resourceTeamGroupMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    svc := client.NewTeamGroupMembershipDetails()

    svc.TeamId(d.Get("team_id").(string))
    svc.GroupId(d.Get("group_id").(string))

    resp, err := svc.Do(ctx)
    if err != nil {
        // If the resource does not exist (404), inform Terraform. We want to immediately
        // return here to prevent further processing.
        if resp.Code == "404" {
            d.SetId("")
            return nil
        }
        return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }

    // msi stands for Map String Interface
    msi := make(map[string]interface{})
    msi["team_id"] = d.Get("team_id").(string)
    msi["group_id"] = resp.Data.GroupId
    msi["role"] = resp.Data.Role
    msi["created_at"] = resp.Data.CreatedAt

    for k, v := range msi {
        if err := d.Set(k, v); err != nil {
            return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
        }
    }

    d.SetId(d.Get("team_id").(string) + "|" + d.Get("group_id").(string))

    return diags
}

func resourceTeamGroupMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)

    svc := client.NewTeamGroupMembershipModify()

    svc.TeamId(d.Get("team_id").(string)).Do(ctx)
    svc.GroupId(d.Get("group_id").(string)).Do(ctx)

    if d.HasChange("role") {
        svc.Role(d.Get("role").(string))
    }

    resp, err := svc.Do(ctx)
    if err != nil {
        return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }       
    
    return resourceTeamGroupMembershipRead(ctx, d, m)
}

func resourceTeamGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    svc := client.NewTeamGroupMembershipDelete()

    svc.TeamId(d.Get("team_id").(string)).Do(ctx)
    svc.GroupId(d.Get("group_id").(string)).Do(ctx)

    resp, err := svc.Do(ctx)
    // for DELETE endpoint idempotence rule breaks
    if err != nil && resp.Code != "NotFound" {
        return newDiagAppend(diags, diag.Error, "delete error " + d.Get("team_id").(string) + " " + d.Get("group_id").(string), fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
    }

    d.SetId("")

    return diags
}
