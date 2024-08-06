package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	var (
		err error
	)
	privateKey, err := os.ReadFile("mykey.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	publicKey, err := os.ReadFile("server.pub")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	privateKeyParsed, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	publicKeyParsed, _, _, _, err := ssh.ParseAuthorizedKey(publicKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	config := &ssh.ClientConfig{
		User: "username",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKeyParsed),
		},
		HostKeyCallback: ssh.FixedHostKey(publicKeyParsed),
	}
	client, err := ssh.Dial("tcp", "localhost:2022", config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	defer session.Close()

	out, err := session.Output("whoami")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Output is: %s\n", out)
}
