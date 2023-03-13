package fivetran

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
)

// DbtTransformationModifyService implements the dbt transformation management,
// modify a dbt atransformation api
// Ref
type DbtTransformationmodifyService struct {
	c                   *fivetran.Client
	dbtTransformationId *string
	schedule            *DbtTransformationSchedule
	runTests            *bool
}

type dbtTransformationModifyRequestBase struct {
	RunTests *bool `json:"run_tests"`
}

type dbtTransformationModifyRequest struct {
	Schedule *dbtTransformationScheduleRequest `json:"schedule,omitempty"`
	dbtTransformationModifyRequestBase
}

type DbtTransformationModifyResponseDataBase struct {
	ID              string    `json:"id"`
	DbtModelId      string    `json:"dbt_model_id`
	OutputModelName string    `json:"output_model_name`
	DbtProjectId    string    `json:"dbt_project_id"`
	LastRun         time.Time `json:"last_run"`
	NextRun         time.Time `json:"next_run"`
	Status          string    `json:"status"`
	RunTests        *bool     `json:"run_tests"`
	ConnectorIds    []string  `json:"connector_ids"`
	ModelIds        []string  `json:"model_ids"`
}

type DbtTransformationModifyResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DbtTransformationModifyResponseDataBase
		Schedule DbtTransformationScheduleResponse `json:"schedule`
	}
}

func (c *Client) NewDbtTransformationModifyService() *DbtTransformationModifyService {
	return &DbtTransformationModifyService{c: c}
}

func (s *DbtTransformationModifyService) requestBase() dbtTransformationModifyRequestBase {
	return dbtTransformationModifyRequestBase{
		RunTests: s.runTests,
	}
}

func (s *DbtTransformationModifyService) request() *dbtTransformationModifyRequest {
	var schedule *dbtTransformationScheduleRequest
	if s.schedule != nil {
		schedule = s.schedule.request()
	}

	return *dbtTransformationModifyRequest{
		Schedule:                           schedule,
		DbtTransformationModifyRequestBase: s.requestBase(),
	}
}

func (s *DbtTransformationModifyService) RunTests(value bool) *DbtTransformationModifyService {
	s.runTests = &value
	return s
}

func (s *DbtTransformationModifyService) Schedule(value *DbtTransformationSchedule) *DbtTransformationModifyService {
	s.schedule = value
	return s
}

func (s *DbtTransformationModifyService) do(ctx context.Context, req, response any) error {
	if s.dbtTransformationID == nil {
		return fmt.Errorf("missing required DbtTransformationID")
	}

	url := fmt.Sprintf("%v/dbt_transformations/%v", s.c.baseURL, *s.dbtTransformationID)
	expectedStatus := 200

	headers := s.c.commonHeaders()
	headers["Content-Type"] = "application/json"
	headers["Accept"] = restAPIv2

	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r := request{
		method:  "PATCH",
		url:     url,
		body:    reqBody,
		queries: nil,
		headers: headers,
		client:  s.c.httpClient,
	}

	respBody, respStatus, err := r.httpRequest(ctx)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected: %v", respStatus, expectedStatus)
		return err
	}

	return nil
}

func (s *DbtTransformationModifyService) Do(ctx context.Context) (DbtTransformationModifyResponse, error) {
	var response DbtTransformationModifyResponse

	err := s.do(ctx, s.request(), &response)

	return response, err
}
