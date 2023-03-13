package fivetran

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
)

// DbtCreateService implements the DBT management, create a dbt API.
// Ref. TODO: Add link to docs

type DbtProjectCreateService struct {
	c             *fivetran.Client
	groupID       *string
	dbtVersion    *string
	gitRemoteUrl  *string
	gitBranch     *string
	defaultSchema *string
	folderPath    *string
	targetName    *string
	threads       *int
}

type dbtProjectCreateRequestBase struct {
	GroupID       *string `json:"groupId,omitempty"`
	DbtVersion    *string `json:"dbtVersion,omitempty"`
	GitRemoteUrl  *string `json:"gitRemoteUrl,omitempty"`
	GitBranch     *string `json:"gitBranch,omitempty"`
	DefaultSchema *string `json:"defaultSchema,omitempty"`
	FolderPath    *string `json:"folderPath,omitempty"`
	TargetName    *string `json:"targetName,omitempty"`
	Threads       *int    `json:"threads,omitempty"`
}

type DbtProjectCreateResponseDataBase struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedById   string    `json:"created_by_id"`
	PublicKey     string    `json:"public_key"`
	GitRemoteUrl  string    `json:"git_remote_url"`
	GitBranch     string    `json:"git_branch"`
	DefaultSchema string    `json:"default_schema"`
	FolderPath    string    `json:"folder_path"`
	TargetName    string    `json:"target_name"`
}

type DbtProjectCreateResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DbtProjectCreateResponseDataBase
	} `json:"data"`
}

func (c *fivetran.Client) NewDbtProjectCreate() *DbtProjectCreateService {
	return &DbtProjectCreateService{c: c}
}

func (s *DbtProjectCreateService) requestBase() dbtProjectCreateRequestBase {
	return dbtProjectCreateRequestBase{
		Service:       s.service,
		GroupID:       s.groupID,
		DbtVersion:    s.dbtVersion,
		GitRemoteUrl:  s.gitRemoteUrl,
		GitBranch:     s.gitBranch,
		DefaultSchema: s.defaultSchema,
		FolderPath:    s.folderPath,
		TargetName:    s.targetName,
		Threads:       s.threads,
	}
}

func (s *DbtProjectCreateService) request() *dbtProjectCreateRequest {
	var auth *connectorAuthRequest

	if s.auth != nil {
		auth = s.auth.request()
	}

	r := &dbtProjectCreateRequest{
		dbtCreateRequestBase: s.requestBase(),
		Auth:                 auth,
	}

	return r
}

func (s *DbtProjectCreateService) Service(value string) *DbtProjectCreateService {
	s.service = &value
	return s
}

func (s *DbtProjectCreateService) GroupID(value string) *DbtProjectCreateService {
	s.groupID = &value
	return s
}

func (s *DbtProjectCreateService) DbtVersion(value string) *DbtProjectCreateService {
	s.dbtVersion = &value
	return s
}

func (s *DbtProjectCreateService) GitRemoteUrl(value string) *DbtProjectCreateService {
	s.gitRemoteUrl = &value
	return s
}

func (s *DbtProjectCreateService) GitBranch(value string) *DbtProjectCreateService {
	s.gitBranch = &value
	return s
}

func (s *DbtProjectCreateService) DefaultSchema(value string) *DbtProjectCreateService {
	s.defaultSchema = &value
	return s
}

func (s *DbtProjectCreateService) FolderPath(value string) *DbtProjectCreateService {
	s.folderPath = &value
	return s
}

func (s *DbtProjectCreateService) TargetName(value string) *DbtProjectCreateService {
	s.targetName = &value
	return s
}

func (s *DbtProjectCreateService) Threads(value int) *DbtProjectCreateService {
	s.threads = &value
	return s
}

func (s *DbtProjectCreateService) do(ctx context.Context, req, response any) error {
	url := fmt.Sprintf("%v/dbt", s.c.BaseUrl)
	expectedStatus := 201

	headers := s.c.commonHeaders()
	headers["Content-Type"] = "appplication/json"
	headers["Accept"] = restAPIv2

	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r := request{
		method:  "POST",
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

func (s *DbtProjectCreateService) Do(ctx context.Context) (DbtProjectCreateResponse, error) {
	var response DbtProjectCreateResponse

	err := s.do(ctx, s.request(), &response)

	return response, err
}
