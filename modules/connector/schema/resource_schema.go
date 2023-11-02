package schema

import (
	"fmt"
	"hash/fnv"

	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func rootSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		ID: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The unique resource identifier (equals to `connector_id`).",
		},
		CONNECTOR_ID: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The unique identifier for the connector within the Fivetran system.",
		},
		SCHEMA_CHANGE_HANDLING: resourceSchemaConfigSchemaShangeHandling(),
		SCHEMA:                 resourceSchemaConfigSchema(),
	}
}

func resourceSchemaConfigSchemaShangeHandling() *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Required: true,
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			v := val.(string)
			if !(v == ALLOW_ALL || v == ALLOW_COLUMNS || v == BLOCK_ALL) {
				errs = append(errs, fmt.Errorf("%q allowed values are: %v, %v or %v, got: %v", key, ALLOW_ALL, ALLOW_COLUMNS, BLOCK_ALL, v))
			}
			return
		},
	}
}

func resourceSchemaConfigSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceSchemaConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				NAME: {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The schema name within your destination in accordance with Fivetran conventional rules.",
				},
				ENABLED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "true",
					ValidateFunc: helpers.ValidateStringBooleanValue,
					Description:  "The boolean value specifying whether the sync for the schema into the destination is enabled.",
				},
				TABLE: resourceSchemaConfigTable(),
			},
		},
	}
}

func resourceSchemaConfigTable() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceTableConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				NAME: {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The table name within your destination in accordance with Fivetran conventional rules.",
				},
				ENABLED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "true",
					ValidateFunc: helpers.ValidateStringBooleanValue,
					Description:  "The boolean value specifying whether the sync of table into the destination is enabled.",
				},
				SYNC_MODE: resourceSchemaConfigSyncMode(),
				COLUMN:    resourceSchemaConfigColumn(),
			},
		},
	}
}

func resourceSchemaConfigSyncMode() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "This field appears in the response if the connector supports switching sync modes for tables.",
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			v := val.(string)
			if !(v == HISTORY || v == SOFT_DELETE || v == LIVE) {
				errs = append(errs, fmt.Errorf("%q allowed values are: %v, %v or %v, got: %v", key, SOFT_DELETE, HISTORY, LIVE, v))
			}
			return
		},
	}
}

func resourceSchemaConfigColumn() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Optional: true, Set: resourceColumnConfigHash,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				NAME: {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The column name within your destination in accordance with Fivetran conventional rules.",
				},
				ENABLED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "true",
					ValidateFunc: helpers.ValidateStringBooleanValue,
					Description:  "The boolean value specifying whether the sync of the column into the destination is enabled.",
				},
				HASHED: {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "false",
					ValidateFunc: helpers.ValidateStringBooleanValue,
					Description:  "The boolean value specifying whether a column should be hashed",
				},
			},
		},
	}
}

func resourceSchemaConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(map[string]interface{})
	var hashKey = vmap[NAME].(string) + vmap[ENABLED].(string)

	if tables, ok := vmap[TABLE]; ok {
		tablesHash := ""
		for _, c := range tables.(*schema.Set).List() {
			tablesHash = tablesHash + helpers.IntToStr(resourceTableConfigHash(c))
		}
		hashKey = hashKey + tablesHash
	}

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}

func resourceTableConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(map[string]interface{})
	var hashKey = vmap[NAME].(string) + vmap[ENABLED].(string) + vmap[SYNC_MODE].(string)

	if columns, ok := vmap[COLUMN]; ok {
		columnsHash := ""
		for _, c := range columns.(*schema.Set).List() {
			columnsHash = columnsHash + helpers.IntToStr(resourceColumnConfigHash(c))
		}
		hashKey = hashKey + columnsHash
	}

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}

func resourceColumnConfigHash(v interface{}) int {
	h := fnv.New32a()
	vmap := v.(map[string]interface{})

	hashed := "false"
	if h, ok := vmap[HASHED].(string); ok {
		hashed = h
	}

	var hashKey = vmap[NAME].(string) + vmap[ENABLED].(string) + hashed

	h.Write([]byte(hashKey))
	return int(h.Sum32())
}
