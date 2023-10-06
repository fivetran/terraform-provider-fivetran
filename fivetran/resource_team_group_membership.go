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
		Schema:        resourceTeamGroupMembershipBase(false),
	}
}

func resourceTeamGroupMembershipBase(datasource bool) map[string]*schema.Schema {
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
		"group": resourceTeamGroupMembershipBaseGroups(datasource),
	}
}

func resourceTeamGroupMembershipBaseGroups(datasource bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Set: func(v interface{}) int {
			return stringInt32Hash(v.(map[string]interface{})["group_id"].(string) + v.(map[string]interface{})["role"].(string))
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"group_id": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    !datasource,
					Description: "The group unique identifier",
				},
				"role": {
					Type:        schema.TypeString,
					Required:    !datasource,
					Computed:    datasource,
					Description: "The team's role that links the team and the group",
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

func resourceTeamGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	teamId := d.Get("team_id").(string)

	if err := resourceTeamGroupMembershipSyncGroups(client, d.Get("group").(*schema.Set).List(), teamId, ctx); err != nil {
		return newDiagAppend(diags, diag.Error, "create error: resourceTeamGroupMembershipSyncGroups", fmt.Sprint(err))
	}

	d.SetId(teamId)

	resourceTeamGroupMembershipRead(ctx, d, m)

	return diags
}

func resourceTeamGroupMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	teamId := d.Get("team_id").(string)

	resp, err := dataSourceTeamGroupMembershipsGet(client, ctx, teamId)
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

	var groups []interface{}
	for _, group := range resp.Data.Items {
		if group.Role == "" {
			continue
		}
		con := make(map[string]interface{})
		con["group_id"] = group.GroupId
		con["role"] = group.Role
		con["created_at"] = group.CreatedAt
		groups = append(groups, con)
	}

	msi["group"] = groups

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(teamId)

	return diags
}

func resourceTeamGroupMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	teamId := d.Get("team_id").(string)

	if d.HasChange("group") {
		if err := resourceTeamGroupMembershipSyncGroups(client, d.Get("group").(*schema.Set).List(), teamId, ctx); err != nil {
			return newDiagAppend(diags, diag.Error, "read error: resourceTeamGroupMembershipSyncGroups", fmt.Sprint(err))
		}
	}

	return resourceTeamGroupMembershipRead(ctx, d, m)
}

func resourceTeamGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	teamId := d.Get("team_id").(string)

	for _, v := range d.Get("group").(*schema.Set).List() {
		if resp, err := client.NewTeamGroupMembershipDelete().TeamId(teamId).GroupId(v.(map[string]interface{})["group_id"].(string)).Do(ctx); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
		}
	}

	return diags
}

func resourceTeamGroupMembershipSyncGroups(client *fivetran.Client, groups []interface{}, teamId string, ctx context.Context) error {
	responseGroups, err := dataSourceTeamGroupMembershipsGet(client, ctx, teamId)
	if err != nil {
		return fmt.Errorf("read error: dataSourceTeamGroupMembershipsGet %v; code: %v", err, responseGroups.Code)
	}

	localGroups := make(map[string]string)
	for _, v := range groups {
		localGroups[v.(map[string]interface{})["group_id"].(string)] = v.(map[string]interface{})["role"].(string)
	}

	remoteGroups := make(map[string]string)
	for _, v := range responseGroups.Data.Items {
		remoteGroups[v.GroupId] = v.Role
	}

	for remoteKey, remoteValue := range remoteGroups {
		role, found := localGroups[remoteKey]

		if !found {
			if resp, err := client.NewTeamGroupMembershipDelete().TeamId(teamId).GroupId(remoteKey).Do(ctx); err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		} else if role != remoteValue {
			if resp, err := client.NewTeamGroupMembershipModify().TeamId(teamId).GroupId(remoteKey).Role(role).Do(ctx); err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}
	}

	for localKey, localValue := range localGroups {
		_, exists := remoteGroups[localKey]

		if !exists {
			if resp, err := client.NewTeamGroupMembershipCreate().TeamId(teamId).GroupId(localKey).Role(localValue).Do(ctx); err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}
	}

	return nil
}
