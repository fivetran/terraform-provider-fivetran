package fivetran

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
)

// DbtProjectModifyService implements the dbt management, modify a dbt api
// Ref. https:// TODO find a link

type DbtProjectModifyService struct {
	c             *fivetran.Client
	dbtID         *string
	groupID       *string
	dbtVersion    *string
	gitRemoteUrl  *string
	gitBranch     *string
	defaultSchema *string
	folderPath    *string
	targetName    *string
	threads       *int
}

type dbtProjectModifyRequestBase struct {
	GroupID       *string `json:"group_id"`
	DbtVersion    *string `json:"dbt_version"`
	GitRemoteUrl  *string `json:"git_remote_url"`
	GitBranch     *string `json:"git_branch"`
	DefaultSchema *string `json:"default_schema"`
	FolderPath    *string `json:"folder_path"`
	TargetName    *string `json:"target_name"`
	Threads       *int    `json:"threads"`
}

type dbtProjectModifyRequest struct {
	dbtModifyRequestBase
}

type DbtProjectModifyResponseDataBase struct {
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

type DbtProjectModifyResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DbtProjectCreateResponseDataBase
	} `json:"data"`
}

func (c *Client) NewDbtModify() *DbtProjectModifyService {
	return &DbtProjectModifyService{c: c}
}

func (s *DbtProjectModifyService) requestBase() dbtProjectModifyRequestBase {
	return dbtProjectModifyRequestBase{
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

func (s *DbtProjectModifyService) request() *dbtProjectModifyRequest {
	return &dbtProjectModifyRequest{
		dbtModifyRequestBase: s.requestBase(),
	}
}

func (s *DbtProjectModifyService) DbtID(value string) *DbtProjectModifyService {
	s.dbtID = &value
	return s
}

func (s *DbtProjectModifyService) GroupID(value string) *DbtProjectModifyService {
	s.groupID = &value
	return s
}

func (s *DbtProjectModifyService) DbtVersion(value string) *DbtProjectModifyService {
	s.dbtVersion = &value
	return s
}

func (s *DbtProjectModifyService) GitRemoteUrl(value string) *DbtProjectModifyService {
	s.gitRemoteUrl = &value
	return s
}

func (s *DbtProjectModifyService) GitBranch(value string) *DbtProjectModifyService {
	s.gitBranch = &value
	return s
}

func (s *DbtProjectModifyService) DefaultSchema(value string) *DbtProjectModifyService {
	s.defaultSchema = &value
	return s
}

func (s *DbtProjectModifyService) FolderPath(value string) *DbtProjectModifyService {
	s.folderPath = &value
	return s
}

func (s *DbtProjectModifyService) TargetName(value string) *DbtProjectModifyService {
	s.targetName = &value
	return s
}

func (s *DbtProjectModifyService) Threads(value int) *DbtProjectModifyService {
	s.threads = &value
	return s
}

func (s *DbtProjectModifyService) do(ctx context.Context, req, response any) error {

	if s.dbtID == nil {
		return fmt.Errorf("missing required DbtID")
	}

	url := fmt.Sprintf("%v/dbt/%v", s.c.baseUrl, *s.connectorID)
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
		return nil, err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected: %v", respStatus, expectedStatus)
		return err
	}

	return nil
}

func (s *DbtProjectModifyService) Do(ctx context.Context) (DbtProjectModifyResponse, error) {
	var response DbtProjectModifyResponse

	err := s.do(ctx, s.request(), &response)

	return response, err
}
