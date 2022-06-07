package fivetran

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceGroupUsersHashGroupUserEmail,
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

	if err := resourceGroupUsersAddUsersToGroup(client, d.Get("user").(*schema.Set).List(), groupID, ctx, diags); err != nil {
		if deleteErr := deleteUsersFromGroup(client, d.Get("user").(*schema.Set).List(), groupID, ctx, diags); deleteErr != nil {
			return newDiagAppend(diags, diag.Error, "create error: resourceGroupUsersAddUsersToGroup", fmt.Sprint(deleteErr))
		}
		return newDiagAppend(diags, diag.Error, "create error: resourceGroupUsersAddUsersToGroup", fmt.Sprint(err))
	}

	d.SetId(resp.Data.ID)
	resourceGroupUsersRead(ctx, d, m)

	return diags
}

// resourceGroupAddUsersToGroup ranges over a list of users and add them to a group
func resourceGroupUsersAddUsersToGroup(client *fivetran.Client, users []interface{}, groupID string, ctx context.Context, diags diag.Diagnostics) error {
	for _, user := range users {
		email := user.(map[string]interface{})["email"].(string)
		role := user.(map[string]interface{})["role"].(string)

		// Add user to group
		respGroupAddUser, err := client.NewGroupAddUser().GroupID(groupID).Email(email).Role(role).Do(ctx)

		if err != nil {
			// If something went wrong we shoud remove users already added
			return fmt.Errorf("%v; code: %v; message: %v", err, respGroupAddUser.Code, respGroupAddUser.Message)
		}
	}

	return nil
}

func deleteUsersFromGroup(client *fivetran.Client, users []interface{}, groupID string, ctx context.Context, diags diag.Diagnostics) error {
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
	} else {
		// try find all users added before failure to rollback changes
		for _, user := range respUsers.Data.Items {
			var i = sort.SearchStrings(emails, user.Email)
			if i < len(emails) && emails[i] == user.Email {
				respDeleteUser, err := client.NewGroupRemoveUser().GroupID(groupID).UserID(user.ID).Do(ctx)
				if err != nil {
					return fmt.Errorf("%v; code: %v; message: %v", errRead, respDeleteUser.Code, respDeleteUser.Message)
				}
			}
		}
	}
	return nil
}

func resourceGroupUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDetails()

	groupID := d.Get("id").(string)

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
	msi["user"] = resourceGroupUsersFlattenGroupUsers(&respUsers, users)
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
		respUsers, err := dataSourceGroupUsersGetUsers(client, groupID, ctx)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "read error: dataSourceGroupUsersGetUsers", fmt.Sprintf("%v; code: %v; message: %v", err, respUsers.Code, respUsers.Message))
		}

		if err := resourceGroupUsersSyncUsers(client, &respUsers, d.Get("user").(*schema.Set).List(), groupID, ctx); err != nil {
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
	// TODO: remove user memberships, not group itself
	groupID := d.Get("group_id").(string)
	users := d.Get("user").(*schema.Set).List()
	deleteUsersFromGroup(client, users, groupID, ctx, diags)
	return diags
}

// resourceGroupHashGroupUserID returns an int hash of an user associated with a group.
// It is the user unique ID on that group.
func resourceGroupUsersHashGroupUserEmail(v interface{}) int {
	h := fnv.New32a()
	var hashKey = v.(map[string]interface{})["email"].(string) + v.(map[string]interface{})["role"].(string)
	h.Write([]byte(hashKey))
	return int(h.Sum32())
}

// resourceGroupFlattenGroupUsers receives a *fivetran.GroupListUsersResponse and returns a []interface{}
// containing the data type accepted by the "user" set. The group creator is ommited from the return
func resourceGroupUsersFlattenGroupUsers(resp *fivetran.GroupListUsersResponse, localUsers []interface{}) []interface{} {
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

// resourceGroupSyncUsers syncs users associated with a group between the Terraform state and the REST API.
// TODO: Check if this can be simplified using d.GetChange().
func resourceGroupUsersSyncUsers(client *fivetran.Client, resp *fivetran.GroupListUsersResponse, localUsers []interface{}, groupID string, ctx context.Context) error {
	type userType struct {
		role string
		id   string
	}

	// Make a map of remoteUsers ommiting the group creator
	remoteUsers := make(map[string]userType)
	for _, v := range resp.Data.Items {
		remoteUsers[v.Email] = userType{
			role: v.Role,
			id:   v.ID,
		}
	}

	// Make a map of local users
	loUsers := make(map[string]userType)
	for _, v := range localUsers {
		loUsers[v.(map[string]interface{})["email"].(string)] = userType{
			role: v.(map[string]interface{})["role"].(string),
			id:   v.(map[string]interface{})["id"].(string),
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
			if resp, err := client.NewGroupRemoveUser().GroupID(groupID).UserID(remoteUsers[remoteKey].id).Do(ctx); err != nil {
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
				if localValue.role != remoteUsers[remoteKey].role {
					// Update role
					// Remove old user membership
					respDeleteUser, err := client.NewGroupRemoveUser().GroupID(groupID).UserID(remoteUsers[remoteKey].id).Do(ctx)
					if err != nil {
						return fmt.Errorf("%v; code: %v; message: %v", err, respDeleteUser.Code, respDeleteUser.Message)
					}
					// Add user with new role
					respAddUser, err := client.NewGroupAddUser().GroupID(groupID).Email(localKey).Role(localValue.role).Do(ctx)
					if err != nil {
						return fmt.Errorf("%v; code: %v; message: %v", err, respAddUser.Code, respAddUser.Message)
					}
				}
			}
		}

		if !found {
			// Get user email address
			respGroupAddUser, err := client.NewGroupAddUser().GroupID(groupID).Email(localKey).Role(localValue.role).Do(ctx)
			if err != nil {
				return fmt.Errorf("%v; code: %v; message: %v", err, respGroupAddUser.Code, respGroupAddUser.Message)
			}
		}
	}

	return nil
}
