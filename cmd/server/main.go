package main

import (
	"fmt"
	"os"

	"github.com/Dan4ik7/ssh"
)

func main() {
	var (
		err error
	)
	authorizedKeyBytes, err := os.ReadFile("mykey.pub")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	privateKey, err := os.ReadFile("server.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	if err = ssh.StartServer(privateKey, authorizedKeyBytes); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
