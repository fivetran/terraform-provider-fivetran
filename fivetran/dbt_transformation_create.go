package fivetran

import (
	"time"

	"github.com/fivetran/go-fivetran"
)

// DbtTransformationCreateService implements the dbt transformation management,
// create a dbt transformation API
// Ref. TODO find proper linl

type DbtTransformationCreateService struct {
	c          *fivetran.Client
	dbtModelID *string
	schedule   *DbtTransformationSchedule
	runTests   *bool
}

type dbtTransformationCreateRequestBase struct {
	DbtModelID *string `json:"dbt_model_id,omitempty"`
	RunTests   *bool   `json:"run_tests,omitempty`
}

type dbtTransformationCreateRequest struct {
	dbtTransformationCreateRequestBase
	Schedule *dbtTransformationScheduleRequest `json:"schedule,omitempty"`
}

type DbtTransformationCreateResponseBase struct {
	ID              string    `json:"id"`
	DbtModelID      string    `json:"dbt_model_id"`
	OutputModelName string    `json:"output_model_name"`
	DbtProjectID    string    `json:"dbt_project_id"`
	LastRun         time.Time `json:"last_run"`
	NextRun         time.Time `json:"next_run"`
	Status          string    `json:"status"`
	RunTests        bool      `json:"run_tests"`
	ConnectorIDs    []string  `json:"connector_ids"`
	ModelIDs        []string  `json:"model_ids"`
}

type DbtTransformationCreateResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DbtTransformationCreateResponseBase
		Schedule DbtTransformationScheduleResponse
	} `json:"data"`
}

func (c *fivetran.Client) NewDbtTransformationCreate() *DbtTransformationCreateService {
	return &DbtTransformationCreateService{c: c}
}

func (s *DbtTransformationCreateService) requestBase() dbtTransformationCreateRequestBase {
	return dbtTransformationCreateRequestBase{
		DbtModelID: s.dbtModelID,
		RunTests:   s.runTests,
	}
}

func (s *DbtTransformationCreateService) request() *dbtTransformationCreateRequest {
	var schedule *dbtTransformationScheduleRequest
	if s.schedule != nil {
		schedule = s.schedule.request()
	}

	r := &dbtTransformationCreateRequest{
		dbtTransformationCreateRequestBase: s.requestBase(),
		Schedule:                           schedule,
	}

	return r
}

func (s *DbtTransformationCreateService) DbtModelID(value string) *DbtTransformationCreateService {
	s.dbtModelID = &value
	return s
}

func (s *DbtTransformationCreateService) RunTests(value bool) *DbtTransformationCreateService {
	s.runTests = &value
	return s
}
