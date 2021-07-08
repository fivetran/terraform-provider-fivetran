package fivetran

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":         {Type: schema.TypeString, Computed: true},
			"name":       {Type: schema.TypeString, Required: true},
			"created_at": {Type: schema.TypeString, Computed: true},
			"user": {Type: schema.TypeSet, Optional: true, Set: resourceGroupHashGroupUserID,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":   {Type: schema.TypeString, Required: true},
						"role": {Type: schema.TypeString, Required: true},
					},
				},
			},
			// creator is used to store the id of the user who created the group.
			// It is important to store it because NewGroupListUsers returns all users associated
			// with the group, but the group creator is not explicit declared in the "user" set.
			"creator":      {Type: schema.TypeString, Computed: true},
			"last_updated": {Type: schema.TypeString, Optional: true, Computed: true}, // internal
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupCreate()

	resp, err := svc.Name(d.Get("name").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	groupID := resp.Data.ID
	users := d.Get("user").(*schema.Set).List()

	groupCreator, err := resourceGroupGetCreator(client, resp.Data.ID, ctx)
	if err != nil {
		// If resourceGroupGetCreator returns an error, the recently created group is deleted
		respDelete, errDelete := client.NewGroupDelete().GroupID(groupID).Do(ctx)
		if errDelete != nil {
			diags = newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, respDelete.Code, respDelete.Message))
		}

		return newDiagAppend(diags, diag.Error, "create error: groupCreator", fmt.Sprint(err))
	}
	d.Set("creator", groupCreator)

	err = resourceGroupAddUsersToGroup(client, users, groupID, ctx)
	if err != nil {
		// If resourceGroupAddUsersToGroup returns an error, the recently created group is deleted
		respDelete, errDelete := client.NewGroupDelete().GroupID(groupID).Do(ctx)
		if errDelete != nil {
			diags = newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, respDelete.Code, respDelete.Message))
		}

		return newDiagAppend(diags, diag.Error, "create error: resourceGroupAddUsersToGroup", fmt.Sprint(err))
	}

	d.SetId(resp.Data.ID)
	resourceGroupRead(ctx, d, m)

	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDetails()

	groupID := d.Get("id").(string)
	creatorID := d.Get("creator").(string)
	users := d.Get("user").(*schema.Set).List()
	svc.GroupID(groupID)

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	respUsers, err := dataSourceGroupUsersGetUsers(client, groupID, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "read error: dataSourceGroupUsersGetUsers", fmt.Sprintf("%v; code: %v; message: %v", err, respUsers.Code, respUsers.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["name"] = resp.Data.Name
	msi["user"] = resourceGroupFlattenGroupUsers(&respUsers, users, creatorID)
	msi["created_at"] = resp.Data.CreatedAt.String()
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupModify()
	var change bool

	groupID := d.Get("id").(string)
	svc.GroupID(groupID)

	if d.HasChange("name") {
		svc.Name(d.Get("name").(string))
		change = true
	}

	if change {
		resp, err := svc.Do(ctx)
		if err != nil {
			// resourceGroupRead here makes sure the state is updated after a NewGroupModify error.
			diags = resourceGroupRead(ctx, d, m)
			return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
		}

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	if d.HasChange("user") {
		respUsers, err := dataSourceGroupUsersGetUsers(client, groupID, ctx)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "read error: dataSourceGroupUsersGetUsers", fmt.Sprintf("%v; code: %v; message: %v", err, respUsers.Code, respUsers.Message))
		}

		if err := resourceGroupSyncUsers(client, &respUsers, d.Get("user").(*schema.Set).List(), d.Get("creator").(string), groupID, ctx); err != nil {
			return newDiagAppend(diags, diag.Error, "read error: resourceGroupSyncUsers", fmt.Sprint(err))
		}

		// TODO: Have to set last_updated here as well, but only if changes happened.
		// A change could have happened even if resourceGroupSyncUsers returned an error
		// d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceGroupRead(ctx, d, m)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDelete()

	resp, err := svc.GroupID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}

// resourceGroupHashGroupUserID returns an int hash of an user associated with a group.
// It is the user unique ID on that group.
func resourceGroupHashGroupUserID(v interface{}) int {
	h := fnv.New32a()
	h.Write([]byte(v.(map[string]interface{})["id"].(string)))
	return int(h.Sum32())
}

// resourceGroupGetCreator returns the id of the first user of a newly created group
func resourceGroupGetCreator(client *fivetran.Client, groupID string, ctx context.Context) (string, error) {
	resp, err := client.NewGroupListUsers().GroupID(groupID).Do(ctx)
	if err != nil {
		return "", err
	}

	return resp.Data.Items[0].ID, nil
}

// resourceGroupAddUsersToGroup ranges over a list of users and add them to a group
func resourceGroupAddUsersToGroup(client *fivetran.Client, users []interface{}, groupID string, ctx context.Context) error {
	for _, user := range users {
		id := user.(map[string]interface{})["id"].(string)
		role := user.(map[string]interface{})["role"].(string)

		// Get user email address
		respUserDetails, err := client.NewUserDetails().UserID(id).Do(ctx)
		if err != nil {
			return fmt.Errorf("%v; code: %v; message: %v", err, respUserDetails.Code, respUserDetails.Message)
		}

		// Add user to group
		respGroupAddUser, err := client.NewGroupAddUser().GroupID(groupID).Email(respUserDetails.Data.Email).Role(role).Do(ctx)
		if err != nil {
			return fmt.Errorf("%v; code: %v; message: %v", err, respGroupAddUser.Code, respGroupAddUser.Message)
		}
	}

	return nil
}

// resourceGroupFlattenGroupUsers receives a *fivetran.GroupListUsersResponse and returns a []interface{}
// containing the data type accepted by the "user" set. The group creator is ommited from the return
func resourceGroupFlattenGroupUsers(resp *fivetran.GroupListUsersResponse, localUsers []interface{}, creatorID string) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	var users []interface{}
	for _, user := range resp.Data.Items {
		if user.ID == creatorID {
			continue
		}

		u := make(map[string]interface{})
		u["id"] = user.ID
		// role is not coming from the REST API until https://fivetran.height.app/T-110695 is fixed.
		// This is a workaround and should be fixed as soon as possible. The regular outcome would
		// be to get the user role from the REST API. That means role update is not working, even
		// if Terraform returns the operation was completed successfully. For now, to change a user
		// role within a group, It is necessary to remove it from the group and add it again with
		// the new role parameter.
		//
		// TODO: Force the replacement of the user association when changing role. It is a better
		// outcome instead of taking. It can't be done with ForceNew, otherwise the whole resource
		// would be replaced.
		u["role"] = resourceGroupGetLocalUserRole(localUsers, user.ID)

		users = append(users, u)
	}

	return users
}

// resourceGroupGetLocalUserRole is a workaround while the REST API doesn't returns `role` correctly
// (https://fivetran.height.app/T-110695). It receives a []interface{} containing users, and an user ID.
// It ranges over the []interface{} and return the user's role stored in the state.
func resourceGroupGetLocalUserRole(localUsers []interface{}, userID string) string {
	for _, user := range localUsers {
		if user.(map[string]interface{})["id"].(string) == userID {
			return user.(map[string]interface{})["role"].(string)
		}
	}

	return ""
}

// resourceGroupSyncUsers syncs users associated with a group between the Terraform state and the REST API.
// TODO: Check if this can be simplified using d.GetChange().
func resourceGroupSyncUsers(client *fivetran.Client, resp *fivetran.GroupListUsersResponse, localUsers []interface{}, creatorID string, groupID string, ctx context.Context) error {
	type userType struct {
		role string
	}

	// Make a map of remoteUsers ommiting the group creator
	remoteUsers := make(map[string]userType)
	for _, v := range resp.Data.Items {
		if v.ID == creatorID {
			continue
		}

		remoteUsers[v.ID] = userType{
			// No real effect, "role" is coming from the local state. See
			// func resourceGroupFlattenGroupUsers(), and
			// func resourceGroupGetLocalUserRole() comments.
			role: resourceGroupGetLocalUserRole(localUsers, v.ID),
		}
	}

	// Make a map of local users
	loUsers := make(map[string]userType)
	for _, v := range localUsers {
		loUsers[v.(map[string]interface{})["id"].(string)] = userType{
			role: v.(map[string]interface{})["role"].(string),
		}
	}

	// Look for remote users not present in the Terraform state. If a user
	// isn't present, then it is removed calling NewGroupRemoveUser.
	for remoteKey := range remoteUsers {
		var found bool
		for localKey := range loUsers {
			if remoteKey == localKey {
				found = true
			}
		}

		if !found {
			if resp, err := client.NewGroupRemoveUser().GroupID(groupID).UserID(remoteKey).Do(ctx); err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}
	}

	// Look for users in the state but not present remote. If a user isn't
	// present remote, then it is added calling NewGroupAddUser.
	for localKey, localValue := range loUsers {
		var found bool
		for remoteKey := range remoteUsers {
			if localKey == remoteKey {
				found = true
			}
		}

		if !found {
			// Get user email address
			respUserDetails, err := client.NewUserDetails().UserID(localKey).Do(ctx)
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, respUserDetails.Code, respUserDetails.Message)
			}

			respGroupAddUser, err := client.NewGroupAddUser().GroupID(groupID).Email(respUserDetails.Data.Email).Role(localValue.role).Do(ctx)
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, respGroupAddUser.Code, respGroupAddUser.Message)
			}
		}
	}

	return nil
}
