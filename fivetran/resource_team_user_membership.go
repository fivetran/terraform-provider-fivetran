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
        Schema:        getTeamUserMembershipSchema(false),
    }
}

func getTeamUserMembershipSchema(datasource bool) map[string]*schema.Schema {
    return map[string]*schema.Schema{
        "id": {
            Type:        schema.TypeString,
            Optional:    true,
            Computed:    true,
            Description: "Fake record Id, compile from team_id and connector_id",
        },
        "team_id": {
            Type:        schema.TypeString,
            ForceNew:    !datasource,
            Required:    datasource,
            Computed:    !datasource,
            Description: "The unique identifier for the team within your account.",
        },
        "user_id": {
            Type:        schema.TypeString,
            ForceNew:    !datasource,
            Required:    datasource,
            Computed:    !datasource,
            Description: "The User unique identifier",
        },
        "role": {
            Type:        schema.TypeString,
            Required:    !datasource,
            Computed:    datasource,
            Description: "The team's role that links the team and the User",
        },
    }
}

func resourceTeamUserMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)

    svc := client.NewTeamUserMembershipCreate()
    svc.TeamId(d.Get("team_id").(string))
    svc.UserId(d.Get("user_id").(string))
    svc.Role(d.Get("role").(string))

    resp, err := svc.Do(ctx)
    if err != nil {
        return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }

    d.SetId(d.Get("team_id").(string) + "|" + d.Get("user_id").(string))

    resourceTeamUserMembershipRead(ctx, d, m)

    return diags
}

func resourceTeamUserMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    svc := client.NewTeamUserMembershipDetails()

    svc.TeamId(d.Get("team_id").(string))
    svc.UserId(d.Get("user_id").(string))

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
    msi["user_id"] = resp.Data.UserId
    msi["role"] = resp.Data.Role

    for k, v := range msi {
        if err := d.Set(k, v); err != nil {
            return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
        }
    }

    d.SetId(d.Get("team_id").(string) + "|" + d.Get("user_id").(string))

    return diags
}

func resourceTeamUserMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)

    svc := client.NewTeamUserMembershipModify()

    svc.TeamId(d.Get("team_id").(string))
    svc.UserId(d.Get("user_id").(string))

    if d.HasChange("role") {
        svc.Role(d.Get("role").(string))
    }

    resp, err := svc.Do(ctx)
    if err != nil {
        return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v", err, resp.Code))
    }       
    
    return resourceTeamUserMembershipRead(ctx, d, m)
}

func resourceTeamUserMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := m.(*fivetran.Client)
    svc := client.NewTeamUserMembershipDelete()

    svc.TeamId(d.Get("team_id").(string))
    svc.UserId(d.Get("user_id").(string))

    resp, err := svc.Do(ctx)
    if err != nil {
        return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
    }

    d.SetId("")

    return diags
}
