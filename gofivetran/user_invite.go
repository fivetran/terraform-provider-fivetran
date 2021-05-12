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
type UserInviteService struct {
	c           *Client
	Femail      string `json:"email,omitempty"`
	FgivenName  string `json:"given_name,omitempty"`
	FfamilyName string `json:"family_name,omitempty"`
	Fphone      string `json:"phone,omitempty"`
	Fpicture    string `json:"picture,omitempty"`
	Frole       string `json:"role,omitempty"`
}

type UserInvite struct {
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

func (c *Client) NewUserInviteService() *UserInviteService {
	return &UserInviteService{c: c}
}

func (s *UserInviteService) Email(email string) *UserInviteService {
	s.Femail = email
	return s
}

func (s *UserInviteService) GivenName(givenName string) *UserInviteService {
	s.FgivenName = givenName
	return s
}

func (s *UserInviteService) FamilyName(familyName string) *UserInviteService {
	s.FfamilyName = familyName
	return s
}

func (s *UserInviteService) Phone(phone string) *UserInviteService {
	s.Fphone = phone
	return s
}

func (s *UserInviteService) Picture(picture string) *UserInviteService {
	s.Fpicture = picture
	return s
}

func (s *UserInviteService) Role(role string) *UserInviteService {
	s.Frole = role
	return s
}

func (s *UserInviteService) Do(ctx context.Context) (UserInvite, error) {
	url := fmt.Sprintf("%v/users", BaseURL)
	expectedStatus := 201
	headers := make(map[string]string)

	headers["Authorization"] = s.c.Authorization
	headers["Content-Type"] = "application/json"

	reqBody, err := json.Marshal(s)
	if err != nil {
		return UserInvite{}, err
	}

	r := request{
		method:  "POST",
		url:     url,
		body:    bytes.NewReader(reqBody),
		queries: nil,
		headers: headers,
	}

	respBody, respStatus, err := httpRequest(r, ctx)
	if err != nil {
		return UserInvite{}, err
	}

	var userInvite UserInvite
	if err := json.Unmarshal(respBody, &userInvite); err != nil {
		return UserInvite{}, err
	}

	if respStatus != expectedStatus {
		err := fmt.Errorf("status code: %v; expected %v", respStatus, expectedStatus)
		return userInvite, err
	}

	return userInvite, nil
}
