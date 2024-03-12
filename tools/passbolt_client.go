package tools

import (
	"context"
	"fmt"
	"github.com/passbolt/go-passbolt/api"
)

type PassboltClient struct {
	Client     *api.Client
	Url        string
	PrivateKey string
	Password   string
	Context    context.Context
}

func Login(client *PassboltClient) {
	err := client.Client.Login(client.Context)
	if err != nil {
		return
	}
	fmt.Println("Logged in!")
}
