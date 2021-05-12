package gofivetran

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type UsersListService struct {
	c      *Client
	limit  int
	cursor string
}

type UsersList struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []struct {
			ID         string    `json:"id"`
			Email      string    `json:"email"`
			GivenName  string    `json:"given_name"`
			FamilyName string    `json:"family_name"`
			Verified   bool      `json:"verified"`
			Invited    bool      `json:"invited"`
			Picture    string    `json:"picture"`
			Phone      string    `json:"phone"`
			LoggedInAt time.Time `json:"logged_in_at"`
			CreatedAt  time.Time `json:"created_at"`
		} `json:"items"`
		NextCursor string `json:"next_cursor"`
	} `json:"data"`
}

func (c *Client) NewUsersListService() *UsersListService {
	return &UsersListService{c: c}
}

func (s *UsersListService) Limit(limit int) *UsersListService {
	s.limit = limit
	return s
}

func (s *UsersListService) Cursor(cursor string) *UsersListService {
	s.cursor = cursor
	return s
}

func (s *UsersListService) Do(ctx context.Context) (UsersList, error) {
	url := fmt.Sprintf("%v/users", BaseURL)
	expectedStatus := 200
	headers := make(map[string]string)
	queries := make(map[string]string)

	headers["Authorization"] = s.c.Authorization

	if s.cursor != "" {
		queries["cursor"] = s.cursor
	}

	if s.limit != 0 {
		queries["limit"] = fmt.Sprint(s.limit)
	}

	r := request{
		method:  "GET",
		url:     url,
		body:    nil,
		queries: queries,
		headers: headers,
	}

	respBody, respStatus, err := httpRequest(r, ctx)
	if err != nil {
		return UsersList{}, err
	}

	var usersList UsersList
	if err := json.Unmarshal(respBody, &usersList); err != nil {
		return UsersList{}, err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected %v", respStatus, expectedStatus)
		return usersList, err
	}

	return usersList, nil
}
