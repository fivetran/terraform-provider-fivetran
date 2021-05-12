package gofivetran

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// F stands for Field
// needs to be exported because of json.Marshal()
type UserModifyService struct {
	c           *Client
	userId      string
	FgivenName  string `json:"given_name,omitempty"`
	FfamilyName string `json:"family_name,omitempty"`
	Fphone      string `json:"phone,omitempty"`
	Fpicture    string `json:"picture,omitempty"`
	Frole       string `json:"role,omitempty"`
}

type UserModify struct {
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

func (c *Client) NewUserModifyService() *UserModifyService {
	return &UserModifyService{c: c}
}

func (s *UserModifyService) UserId(userId string) *UserModifyService {
	s.userId = userId
	return s
}

func (s *UserModifyService) GivenName(givenName string) *UserModifyService {
	s.FgivenName = givenName
	return s
}

func (s *UserModifyService) FamilyName(familyName string) *UserModifyService {
	s.FfamilyName = familyName
	return s
}

func (s *UserModifyService) Phone(phone string) *UserModifyService {
	s.Fphone = phone
	return s
}

func (s *UserModifyService) Picture(picture string) *UserModifyService {
	s.Fpicture = picture
	return s
}

func (s *UserModifyService) Role(role string) *UserModifyService {
	s.Frole = role
	return s
}

func (s *UserModifyService) Do(ctx context.Context) (UserModify, error) {
	if s.userId == "" { // we don't validate business rules (unless it is strictly necessary) // in this case the result would be an empty UserDetails{} with a 200 status code
		err := fmt.Errorf("missing required UserId")
		return UserModify{}, err
	}

	url := fmt.Sprintf("%v/users/%v", BaseURL, s.userId)
	expectedStatus := 200
	headers := make(map[string]string)

	headers["Authorization"] = s.c.Authorization
	headers["Content-Type"] = "application/json"

	reqBody, err := json.Marshal(s)
	if err != nil {
		return UserModify{}, err
	}

	r := request{
		method:  "PATCH",
		url:     url,
		body:    bytes.NewReader(reqBody),
		queries: nil,
		headers: headers,
	}

	respBody, respStatus, err := httpRequest(r, ctx)
	if err != nil {
		return UserModify{}, err
	}

	var userModify UserModify
	if err := json.Unmarshal(respBody, &userModify); err != nil {
		return UserModify{}, err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected %v", respStatus, expectedStatus)
		return userModify, err
	}

	return userModify, nil
}
