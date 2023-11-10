package schema

import (
	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type _config struct {
	schemas map[string]*_schema
}

func (c _config) hasUpdates() bool {
	result := false
	for _, v := range c.schemas {
		result = result || v.updated
	}
	return result
}

func (c _config) prepareRequest(svc *fivetran.ConnectorSchemaConfigUpdateService) *fivetran.ConnectorSchemaConfigUpdateService {
	for k, v := range c.schemas {
		if v.updated {
			svc.Schema(k, v.prepareRequest())
		}
	}
	return svc
}

func (c *_config) override(local *_config, sch string) error {
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

func (c *_config) readFromResourceData(d *schema.ResourceData) {
	c.schemas = make(map[string]*_schema)
	if lsc, ok := d.GetOk(SCHEMA); ok {
		localSchemas := lsc.(*schema.Set).List()
		for _, schema := range localSchemas {
			sMap := schema.(map[string]interface{})
			s := &_schema{}
			s.readFromResourceData(sMap)
			c.schemas[sMap[NAME].(string)] = s
		}
	}
}

func (c *_config) readFromResponse(response connectors.ConnectorSchemaDetailsResponse) {
	c.schemas = make(map[string]*_schema)
	for k, v := range response.Data.Schemas {
		s := &_schema{}
		s.readFromResponse(k, v)
		c.schemas[k] = s
	}
}

func (c _config) toStateObject(sch string, local _config) map[string]interface{} {
	result := make(map[string]interface{})
	result[SCHEMA_CHANGE_HANDLING] = sch
	schemas := make([]interface{}, 0)

	for k, v := range c.schemas {
		var schemaState map[string]interface{}
		var include bool
		if ls, ok := local.schemas[k]; ok {
			schemaState, include = v.toStateObject(sch, ls)
		} else {
			schemaState, include = v.toStateObject(sch, nil)
		}
		if include {
			schemas = append(schemas, schemaState)
		}
	}
	result[SCHEMA] = schemas
	return result
}
