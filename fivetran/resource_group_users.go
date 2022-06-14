package fivetran

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type userType struct {
	role string
	id   string
}

func resourceGroupUsers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupUsersCreate,
		ReadContext:   resourceGroupUsersRead,
		UpdateContext: resourceGroupUsersUpdate,
		DeleteContext: resourceGroupUsersDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":           {Type: schema.TypeString, Computed: true},
			"group_id":     {Type: schema.TypeString, Required: true},
			"user":         resourceGroupUsersSchemaUser(),
			"last_updated": {Type: schema.TypeString, Computed: true}, // internal
		},
	}
}

func resourceGroupUsersSchemaUser() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceGroupUsersHashGroupUser,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id":    {Type: schema.TypeString, Computed: true},
				"email": {Type: schema.TypeString, Required: true},
				"role":  {Type: schema.TypeString, Required: true},
			},
		},
	}
}

func resourceGroupUsersCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	var groupID = d.Get("group_id").(string)

	resp, err := client.NewGroupDetails().GroupID(groupID).Do(ctx)

	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := resourceGroupUsersSyncUsers(client, d.Get("user").(*schema.Set).List(), groupID, ctx); err != nil {
		if deleteErr := resourceGroupUsersDeleteUsersFromGroup(client, d.Get("user").(*schema.Set).List(), groupID, ctx, diags); deleteErr != nil {
			return newDiagAppend(diags, diag.Error, "cleanup after failure error: resourceGroupUsersDeleteUsersFromGroup", fmt.Sprint(deleteErr))
		}
		return newDiagAppend(diags, diag.Error, "create error: resourceGroupSyncUsers", fmt.Sprint(err))
	}

	d.SetId(resp.Data.ID)
	resourceGroupUsersRead(ctx, d, m)

	return diags
}

func resourceGroupUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	groupID := d.Get("id").(string)

	respUsers, err := dataSourceGroupUsersGetUsers(client, groupID, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "read error: dataSourceGroupUsersGetUsers", fmt.Sprintf("%v; code: %v; message: %v", err, respUsers.Code, respUsers.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["group_id"] = groupID
	msi["user"] = resourceGroupUsersFlattenGroupUsers(&respUsers)
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func resourceGroupUsersUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	groupID := d.Get("group_id").(string)

	if d.HasChange("user") {
		if err := resourceGroupUsersSyncUsers(client, d.Get("user").(*schema.Set).List(), groupID, ctx); err != nil {
			return newDiagAppend(diags, diag.Error, "read error: resourceGroupSyncUsers", fmt.Sprint(err))
		}
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceGroupUsersRead(ctx, d, m)
}

func resourceGroupUsersDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	groupID := d.Get("group_id").(string)
	users := d.Get("user").(*schema.Set).List()

	// delete all defined memberships from group
	resourceGroupUsersDeleteUsersFromGroup(client, users, groupID, ctx, diags)

	return diags
}

// resourceGroupHashGroupUserID returns an int hash of an user associated with a group.
// It is the user unique email on that group.
func resourceGroupUsersHashGroupUser(v interface{}) int {
	h := fnv.New32a()
	var hashKey = v.(map[string]interface{})["email"].(string) + v.(map[string]interface{})["role"].(string)
	h.Write([]byte(hashKey))
	return int(h.Sum32())
}

// resourceGroupSyncUsers syncs users associated with a group between the Terraform state and the REST API.
// TODO: Check if this can be simplified using d.GetChange().
func resourceGroupUsersSyncUsers(client *fivetran.Client, localUsers []interface{}, groupID string, ctx context.Context) error {

	respUsers, err := dataSourceGroupUsersGetUsers(client, groupID, ctx)
	if err != nil {
		return fmt.Errorf("read error: dataSourceGroupUsersGetUsers %v; code: %v; message: %v", err, respUsers.Code, respUsers.Message)
	}

	remoteUsers := resourceGroupUsersMapUsersWithRolesByEmails(respUsers)

	// Make a map of local users
	loUsers := make(map[string]userType)
	for _, v := range localUsers {
		loUsers[v.(map[string]interface{})["email"].(string)] = userType{
			role: v.(map[string]interface{})["role"].(string),
			id:   v.(map[string]interface{})["id"].(string),
		}
	}

	// Look for remote users not present in the Terraform state.
	for remoteKey := range remoteUsers {
		_, found := loUsers[remoteKey]

		// If user exists in group, but not found in state we delete user frou group
		if !found {
			if resp, err := client.NewGroupRemoveUser().GroupID(groupID).UserID(remoteUsers[remoteKey].id).Do(ctx); err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, resp.Code, resp.Message)
			}
		}
	}

	// Look for users in the state but not present remote or the users with updated roles.
	// If the user isn't present remote, then it is added calling NewGroupAddUser.
	for localKey, localValue := range loUsers {
		// try get corresponding user by email
		remoteUser, exists := remoteUsers[localKey]
		if exists {
			// check if role is updated in state
			if localValue.role != remoteUser.role {
				err := resourceGroupUsersUpdateUserRoleInGroup(client, groupID, remoteUser.id, localKey, localValue.role, ctx)
				if err != nil {
					return err
				}
			}
		} else {
			// add user if it isn't present in group but defined in state
			respGroupAddUser, err := client.NewGroupAddUser().GroupID(groupID).Email(localKey).Role(localValue.role).Do(ctx)
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, respGroupAddUser.Code, respGroupAddUser.Message)
			}
		}
	}

	return nil
}

func resourceGroupUsersUpdateUserRoleInGroup(client *fivetran.Client, groupID string, userID string, email string, role string, ctx context.Context) error {
	// TODO: update go-fivetran SDK and use updateUserMembershipInGroup endpoint instead
	// Update role
	// Remove old user membership
	respDeleteUser, err := client.NewGroupRemoveUser().GroupID(groupID).UserID(userID).Do(ctx)
	if err != nil {
		return fmt.Errorf("%v; code: %v; message: %v", err, respDeleteUser.Code, respDeleteUser.Message)
	}
	// Add user by email with new role
	respAddUser, err := client.NewGroupAddUser().GroupID(groupID).Email(email).Role(role).Do(ctx)
	if err != nil {
		return fmt.Errorf("%v; code: %v; message: %v", err, respAddUser.Code, respAddUser.Message)
	}
	return nil
}

func resourceGroupUsersDeleteUsersFromGroup(client *fivetran.Client, users []interface{}, groupID string, ctx context.Context, diags diag.Diagnostics) error {
	// Collect all emails from users to add, we don't know actual id's
	var emails []string
	for _, u := range users {
		emails = append(emails, u.(map[string]interface{})["email"].(string))
	}
	// sort emails to execute binary search
	sort.Strings(emails)

	// fetch existing users in group
	respUsers, errRead := dataSourceGroupUsersGetUsers(client, groupID, ctx)

	if errRead != nil {
		return fmt.Errorf("%v; code: %v; message: %v", errRead, respUsers.Code, respUsers.Message)
	}
	// map existing users by email
	existingUsers := resourceGroupUsersMapUsersWithRolesByEmails(respUsers)

	var errors []string
	for _, email := range emails {
		existingUser, exists := existingUsers[email]
		if exists {
			// remove user if exists
			respDeleteUser, err := client.NewGroupRemoveUser().GroupID(groupID).UserID(existingUser.id).Do(ctx)
			if err != nil {
				// we should try to remove all users if possible, so we accumulationg erorrs
				errors = append(errors, fmt.Sprintf("%v; code: %v; message: %v", errRead, respDeleteUser.Code, respDeleteUser.Message))
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}

	return nil
}

func resourceGroupUsersMapUsersWithRolesByEmails(resp fivetran.GroupListUsersResponse) map[string]userType {
	// Make a map of remoteUsers ommiting the group creator
	remoteUsers := make(map[string]userType)
	for _, v := range resp.Data.Items {
		if v.Role == "" {
			continue
		}
		remoteUsers[v.Email] = userType{
			role: v.Role,
			id:   v.ID,
		}
	}
	return remoteUsers
}

// resourceGroupFlattenGroupUsers receives a *fivetran.GroupListUsersResponse and returns a []interface{}
// containing the data type accepted by the "user" set. The group creator is ommited from the return
func resourceGroupUsersFlattenGroupUsers(resp *fivetran.GroupListUsersResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	var users []interface{}
	for _, user := range resp.Data.Items {
		if user.Role == "" {
			continue
		}
		u := make(map[string]interface{})
		u["id"] = user.ID
		u["email"] = user.Email
		u["role"] = user.Role
		users = append(users, u)
	}

	return users
}
