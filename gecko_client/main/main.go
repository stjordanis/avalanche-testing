package main

import (
	"fmt"
	"time"

	"github.com/ava-labs/avalanche-e2e-tests/gecko_client/apis"
	"github.com/ava-labs/gecko/api"
)

func main() {
	uri := "http://127.0.0.1:9650"
	timeout := 2 * time.Second

	generalClient := apis.NewClient(uri, timeout)
	keystore := generalClient.KeystoreAPI()
	avm := generalClient.XChainAPI()

	user := api.UserPass{Username: "wifbeir3iryb3r", Password: "winbuf3iyfb4ry84irybiwf"}
	if _, err := keystore.CreateUser(user); err != nil {
		fmt.Printf("Couldn't create user\n")
	}

	address, err := avm.CreateAddress(user)
	if err != nil {
		fmt.Printf("Failed: %s\n", err)
	}

	fmt.Printf("Created address: %s\n", address)
}
