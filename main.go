package main

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/gofivetran"
)

var apiKey string    // temp; values set in hidden temp_credentials.go
var apiSecret string // temp; values set in hidden temp_credentials.go

func main() {
	tempCredentials()
	client := gofivetran.NewClient(apiKey, apiSecret)

	test := "usersList"

	switch test {
	case "usersList":
		usersList(client)
	case "userDetails":
		userDetails(client)
	case "userInvite":
		userInvite(client)
	case "userModify":
		userModify(client)
	case "userDelete":
		userDelete(client)
	default:
	}
}

func usersList(client *gofivetran.Client) {
	users := client.NewUsersListService()
	usersList, err := users.Limit(1).Cursor("eyJza2lwIjoxfQ").Do(context.Background())
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", usersList)
		return
	}
	fmt.Printf("%+v\n", usersList)
}

func userDetails(client *gofivetran.Client) {
	udetails := client.NewUserDetailsService()
	user, err := udetails.UserId("vivo_lumpiness").Do(context.Background())
	// user, err := udetails.Do(context.Background())

	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", user)
		return
	}
	fmt.Printf("%+v\n", user)
}

func userInvite(client *gofivetran.Client) {
	newUserInvite := client.NewUserInviteService()
	// user, err := udetails.UserId("113690429484351494364").Do(context.Background())
	user, err := newUserInvite.
		Email("felipe.neuwald+apitest12@fivetran.com").
		GivenName("Felipe").
		FamilyName("Neuwald API TEST 12").
		Role("ReadOnly").
		Phone("+353 83 111 1111").
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", user)
		return
	}
	fmt.Printf("%+v\n", user)
}

func userModify(client *gofivetran.Client) {
	newUserModify := client.NewUserModifyService()
	// user, err := udetails.UserId("113690429484351494364").Do(context.Background())
	user, err := newUserModify.
		UserId("vivo_lumpiness").
		GivenName("Felipe").
		// FamilyName("Neuwald API MODIFY TEST 11").
		// Role("ReadOnly").
		// Phone("+353 83 111 1111").
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", user)
		return
	}
	fmt.Printf("%+v\n", user)
}

func userDelete(client *gofivetran.Client) {
	newUserDelete := client.NewUserDeleteService()
	user, err := newUserDelete.UserId("monk_replace").Do(context.Background())
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", user)
		return
	}
	fmt.Printf("%+v\n", user)
}
