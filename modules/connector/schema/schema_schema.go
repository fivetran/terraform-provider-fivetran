package schema

import (
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/connectors"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
)

type _schema struct {
	_element
	tables map[string]*_table
}

func (s _schema) prepareRequest() *connectors.ConnectorSchemaConfigSchema {
	result := fivetran.NewConnectorSchemaConfigSchema()
	if s.enabledPatched {
		result.Enabled(s.enabled)
	}
	for k, v := range s.tables {
		if v.updated {
			result.Table(k, v.prepareRequest())
		}
	}
	return result
}
func (s *_schema) override(local *_schema, sch string) error {
	if local != nil {
		if local.enabled != s.enabled {
			s.setEnabled(local.enabled)
		}
		for tName, t := range s.tables {
			if lTable, ok := local.tables[tName]; ok {
				err := t.override(lTable, sch)
				if err != nil {
					return fmt.Errorf("error while patching schema %s: \n\t%s", s.name, err.Error())
				}
				s.updated = s.updated || t.updated
			} else {
				err := t.override(nil, sch)
				if err != nil {
					return fmt.Errorf("error while patching schema %s: \n\t%s", s.name, err.Error())
				}
				s.updated = s.updated || t.updated
			}
		}
	} else {
		s.setEnabled(sch == ALLOW_ALL)
		for _, t := range s.tables {
			err := t.override(nil, sch)
			if err != nil {
				return fmt.Errorf("error while patching schema %s: \n\t%s", s.name, err.Error())
			}
			s.updated = s.updated || t.updated
		}
	}
	return nil
}

func (s *_schema) readFromResponse(name string, response *connectors.ConnectorSchemaConfigSchemaResponse) {
	s.name = name
	s.enabled = *response.Enabled

	// schema could be always set enabled/disabled
	s.patchAllowed = nil

	s.tables = make(map[string]*_table)
	for k, v := range response.Tables {
		t := &_table{}
		t.readFromResponse(k, v)
		s.tables[k] = t
	}
}

func getBoolValue(value interface{}) bool {
	if result, ok := value.(bool); ok {
		return result
	}
	if str, ok := value.(string); ok {
		return helpers.StrToBool(str)
	}
	return false
}

func (s *_schema) readFromResourceData(source map[string]interface{}, sch string) {
	s.name = source[NAME].(string)
	s.tables = make(map[string]*_table)
	tables := getTables(source)
	if len(tables) > 0 {
		s.readTables(tables, sch)
	}
	if enabled, ok := source[ENABLED]; ok {
		s.enabled = getBoolValue(enabled)
	} else {
		s.enabled = len(tables) > 0 || sch == ALLOW_ALL
	}
}

func getTables(source map[string]interface{}) []interface{} {
	tablesArray := []interface{}{}
	if tables, ok := source[TABLE].([]interface{}); ok {
		tablesArray = tables
	}
	return tablesArray
}

func (s *_schema) readTables(tables []interface{}, sch string) {
	for _, table := range tables {
		tMap := table.(map[string]interface{})
		t := &_table{}
		t.readFromResourceData(tMap, sch)
		s.tables[tMap[NAME].(string)] = t
	}
}

func (s _schema) toStateObject(sch string, local *_schema) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	result[ENABLED] = helpers.BoolToStr(s.enabled)
	result[NAME] = s.name
	tables := make([]interface{}, 0)

	for k, v := range s.tables {
		var tableState map[string]interface{}
		var include bool
		if local != nil {
			if lt, ok := local.tables[k]; ok {
				tableState, include = v.toStateObject(sch, lt)
			} else {
				tableState, include = v.toStateObject(sch, nil)
			}
		} else {
			tableState, include = v.toStateObject(sch, nil)
		}
		if include {
			tables = append(tables, tableState)
		}
	}

	result[TABLE] = tables

	// schema has been configured locally OR has tables to include OR schema inconsistent by policy
	include := local != nil || len(tables) > 0 || s.enabled != (sch == ALLOW_ALL)
	return result, include
}
