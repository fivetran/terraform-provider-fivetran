package fivetran

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fivetran/go-fivetran"
)

// DbtProjectDeleteService implements the dbt management, delete a dbt api.
// Ref. TODO add url
type DbtProjectDeleteService struct {
	c     *fivetran.Client
	dbtID *string
}

type DbtProjectDeleteResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *fivetran.Client) NewDbtDelete() *DbtProjectDeleteService {
	return &DbtProjectDeleteService{c: c}
}

func (s *DbtProjectDeleteService) DbtID(value string) *DbtProjectDeleteService {
	s.dbtID = &value
	return s
}

func (s *DbtProjectDeleteService) Do(ctx context.Context) (DbtProjectDeleteResponse, error) {
	var response DbtProjectDeleteResponse

	if s.dbtID == nil {
		return response, fmt.Errorf("missing required DbtID")
	}

	url := fmt.Sprintf("%v/dbt/%v", s.c.baseURL, *s.dbtID)
	expectedStatus := 200

	headers := s.c.commonHeaders()

	r := request{
		method:  "DELETE",
		url:     url,
		body:    nil,
		queries: nil,
		headers: headers,
		client:  s.c.httpClient,
	}

	respBody, respStatus, err := r.httpRequest(ctx)
	if err != nil {
		return response, err
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return response, err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected: %v", respStatus, expectedStatus)
		return response, err
	}

	return response, nil
}
