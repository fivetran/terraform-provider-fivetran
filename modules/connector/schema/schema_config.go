package schema

import (
	"github.com/fivetran/go-fivetran/connectors"
)

type SchemaConfig struct {
	schemas map[string]*_schema
}

func (c SchemaConfig) HasUpdates() bool {
	result := false
	for _, v := range c.schemas {
		result = result || v.updated
	}
	return result
}

func (c SchemaConfig) PrepareRequest(svc *fivetran.ConnectorSchemaConfigUpdateService) *fivetran.ConnectorSchemaConfigUpdateService {
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

func (c SchemaConfig) GetSchemas(sch string, local SchemaConfig) []interface{} {
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

	return schemas
}

func (c SchemaConfig) ToStateObject(sch string, local SchemaConfig) map[string]interface{} {
	result := make(map[string]interface{})
	result[SCHEMA_CHANGE_HANDLING] = sch
	result[SCHEMA] = c.GetSchemas(sch, local)
	return result
}
