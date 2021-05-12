package gofivetran

import (
	"context"
	"encoding/json"
	"fmt"
)

type UserDeleteService struct {
	c      *Client
	userId string
}

type UserDelete struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) NewUserDeleteService() *UserDeleteService {
	return &UserDeleteService{c: c}
}

func (s *UserDeleteService) UserId(userId string) *UserDeleteService {
	s.userId = userId
	return s
}

func (s *UserDeleteService) Do(ctx context.Context) (UserDelete, error) {
	if s.userId == "" { // we don't validate business rules (unless it is strictly necessary) // in this case the result would be an empty UserDetails{} with a 200 status code
		err := fmt.Errorf("missing required UserId")
		return UserDelete{}, err
	}

	url := fmt.Sprintf("%v/users/%v", BaseURL, s.userId)
	expectedStatus := 200
	headers := make(map[string]string)

	headers["Authorization"] = s.c.Authorization

	r := request{
		method:  "DELETE",
		url:     url,
		body:    nil,
		queries: nil,
		headers: headers,
	}

	respBody, respStatus, err := httpRequest(r, ctx)
	if err != nil {
		return UserDelete{}, err
	}

	var userDelete UserDelete
	if err := json.Unmarshal(respBody, &userDelete); err != nil {
		return UserDelete{}, err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected %v", respStatus, expectedStatus)
		return userDelete, err
	}

	return userDelete, nil
}
