package gofivetran

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type UserDetailsService struct {
	c      *Client
	userId string
}

type UserDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
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
	} `json:"data"`
}

func (c *Client) NewUserDetailsService() *UserDetailsService {
	return &UserDetailsService{c: c}
}

func (s *UserDetailsService) UserId(userId string) *UserDetailsService {
	s.userId = userId
	return s
}

func (s *UserDetailsService) Do(ctx context.Context) (UserDetails, error) {
	if s.userId == "" { // we don't validate business rules (unless it is strictly necessary) // in this case the result would be an empty UserDetails{} with a 200 status code
		err := fmt.Errorf("missing required UserId")
		return UserDetails{}, err
	}

	url := fmt.Sprintf("%v/users/%v", BaseURL, s.userId)
	expectedStatus := 200
	headers := make(map[string]string)

	headers["Authorization"] = s.c.Authorization

	r := request{
		method:  "GET",
		url:     url,
		body:    nil,
		queries: nil,
		headers: headers,
	}

	respBody, respStatus, err := httpRequest(r, ctx)
	if err != nil {
		return UserDetails{}, err
	}

	var userDetails UserDetails
	if err := json.Unmarshal(respBody, &userDetails); err != nil {
		return UserDetails{}, err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected %v", respStatus, expectedStatus)
		return userDetails, err
	}

	return userDetails, nil
}
