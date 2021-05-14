package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func set(d *schema.ResourceData, kvmap map[string]interface{}) error { // better func name?
	for k, v := range kvmap {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}
