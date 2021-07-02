package fivetran

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var limit = 1000 // REST API response objects limit per HTTP request

// newDiag receives a diag.Severity, a summary, a detail, and returns a diag.Diagnostic
func newDiag(severity diag.Severity, summary, detail string) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: severity,
		Summary:  summary,
		Detail:   detail,
	}
}

// newAppendDiag receives diag.Diagnostics, a diag.Severity, a summary, and a detail. It makes a new
// diag.Diagnostic, appends it to the diag.Diagnostics and returns the diag.Diagnostics.
func newDiagAppend(diags diag.Diagnostics, severity diag.Severity, summary, detail string) diag.Diagnostics {
	diags = append(diags, newDiag(severity, summary, detail))
	return diags
}

// set is a temp function. It should be removed, there is no need for it.
func set(d *schema.ResourceData, kvmap map[string]interface{}) error {
	for k, v := range kvmap {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}

// debug is a temp function. It should be improved to accept a variadic parameter
// and its name should change to logDebug
func debug(v interface{}) {
	log.Println(fmt.Sprintf("[DEBUG] FIVETRAN: %s", v))
}

// TEMP STUFF
//////////////////////////////////////////////////////////////////////////////////////////////
// // getUserEmail receives users (fivetran.UsersListRespose) and a string id, ranges over the
// // users and look for a user matching the id. If an id is found, it returns the email address
// // associated to that user id and a nil error. ............ // should return error if no
// // user is found?
// func getUserEmail(users fivetran.UsersListResponse, id string) (string, error) { // should return error?
// 	for _, user := range users.Data.Items {
// 		if user.ID == id {
// 			return user.Email, nil
// 		}
// 	}

// 	return "", fmt.Errorf("couldn't find user %v", id) // better error handling
// }

//////////////////////////////////////////////////////////////////////////
// func getUsersList(client *fivetran.Client, ctx context.Context) error {
// 	svc := client.NewUsersList()

// 	var resp fivetran.UsersListResponse
// 	var respNextCursor string
// 	limit := 1000

// 	for {
// 		var err error
// 		var respInner fivetran.UsersListResponse
// 		if respNextCursor == "" {
// 			respInner, err = svc.Limit(limit).Do(ctx)
// 		}
// 		if respNextCursor != "" {
// 			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
// 		}
// 		if err != nil {
// 			return fmt.Errorf("%v: %v", err, respInner)
// 		}

// 		for _, item := range respInner.Data.Items {
// 			resp.Data.Items = append(resp.Data.Items, item)
// 		}

// 		if respInner.Data.NextCursor == "" {
// 			break
// 		}

// 		respNextCursor = respInner.Data.NextCursor
// 	}

// 	return nil
// }
