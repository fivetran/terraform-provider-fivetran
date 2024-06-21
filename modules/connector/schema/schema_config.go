package schema

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type SchemaConfig struct {
	schemas map[string]*_schema
}

func (c SchemaConfig) ValidateSchemas(connectorId string, schemas map[string]*connectors.ConnectorSchemaConfigSchemaResponse, client fivetran.Client, ctx context.Context) (error, bool) {
	for sName, schema := range c.schemas {
		if responseSchema, ok := schemas[sName]; ok {
			err, needReload := schema.validateTables(connectorId, sName, responseSchema, client, ctx)
			if err != nil {
				return err, needReload
			}
		} else {
			return fmt.Errorf("Schema with name `%s` not found in source.", sName), true
		}
	}
	return nil, false
}

func (c SchemaConfig) HasUpdates() bool {
	result := false
	for _, v := range c.schemas {
		result = result || v.updated
	}
	return result
}

func (c SchemaConfig) PrepareRequest(svc *connectors.ConnectorSchemaConfigUpdateService) *connectors.ConnectorSchemaConfigUpdateService {
	for k, v := range c.schemas {
		if v.updated {
			svc.Schema(k, v.prepareRequest())
		}
	}
	return svc
}
func (c *SchemaConfig) Override(local *SchemaConfig, sch string) error {
	if local != nil {
		for sName, s := range c.schemas {
			if lSchema, ok := local.schemas[sName]; ok {
				err := s.override(lSchema, sch)
				if err != nil {
					return err
				}
			} else {
				// Schema not configured
				err := s.override(nil, sch)
				if err != nil {
					return err
				}
			}
		}
	} else {
		// Align not configured schemas to policy
		for _, s := range c.schemas {
			err := s.override(nil, sch)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *SchemaConfig) ReadFromRawSourceData(d []interface{}, sch string) {
	c.schemas = make(map[string]*_schema)
	for _, schema := range d {
		if sMap, ok := schema.(map[string]interface{}); ok {
			s := &_schema{}
			s.readFromResourceData(sMap, sch)
			c.schemas[sMap[NAME].(string)] = s
		}
	}
}

func (c *SchemaConfig) ReadFromResponse(response connectors.ConnectorSchemaDetailsResponse) {
	c.schemas = make(map[string]*_schema)
	for k, v := range response.Data.Schemas {
		s := &_schema{}
		s.readFromResponse(k, v)
		c.schemas[k] = s
	}
}

func (c SchemaConfig) GetSchemas(sch string, local SchemaConfig, diag *diag.Diagnostics) []interface{} {
	schemas := make([]interface{}, 0)

	for k, v := range c.schemas {
		var schemaState map[string]interface{}
		var include bool
		if ls, ok := local.schemas[k]; ok {
			schemaState, include = v.toStateObject(sch, ls, diag)
		} else {
			schemaState, include = v.toStateObject(sch, nil, diag)
		}
		if include {
			schemas = append(schemas, schemaState)
		}
	}

	// Include locally configured, but not represented in upstream
	for k, v := range local.schemas {
		if _, ok := c.schemas[k]; !ok {
			diag.AddWarning(
				"Schema might be missconfigured.",
				fmt.Sprintf(
					"Schema with name `%v`, defined in your config, not found in upstream source config.\n"+
						"Schema might be deleted from source or renamed.\n "+
						"Please remove it from your configuration, or align its name with source schema.", k),
			)
			schemaState, include := v.toStateObject(sch, nil, diag)
			if include {
				schemas = append(schemas, schemaState)
			}
		}
	}

	return schemas
}
