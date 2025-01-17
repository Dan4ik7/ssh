package ssh

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func StartServer(privateKey []byte, authorizedKeys []byte) error {
	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeys) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeys)
		if err != nil {
			return fmt.Errorf("parse authorized keys error: %s", err)
		}
		authorizedKeysMap[string(pubKey.Marshal())] = true
		// if you have only one key then it will be passed t
		//authorizedKeys and the rest will be empty
		//thererfor we assign empty to authorized keys and it will finish the loop.
		authorizedKeys = rest
	}

	config := &ssh.ServerConfig{
		//publicKeyCallback will validate whether a public key is correct
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			//so this checks if the key exists in this map
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}
	private, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("parse private key error: %s", err)
	}
	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		return fmt.Errorf("failed to listen for connection: %s", err)
	}

	for {
		nConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("failed to accept incoming connection: %s\n", err)
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			fmt.Printf("failed to handshake: %s\n", err)
		}
		if conn != nil && conn.Permissions != nil {
			log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])
		}
		go ssh.DiscardRequests(reqs)

		go handleConnection(conn, chans)
	}
}

func handleConnection(conn *ssh.ServerConn, chans <-chan ssh.NewChannel) {

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			fmt.Printf("Could not accept channel: %v", err)
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				fmt.Printf("Request Type made by client: %s\n", req.Type)
				switch req.Type {
				case "exec":
					payload := bytes.TrimPrefix(req.Payload, []byte{0, 0, 0, 6})
					channel.Write([]byte(execSomething(conn, payload)))
					channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
					req.Reply(true, nil)
					channel.Close()
				case "shell":
					req.Reply(true, nil)
				case "pty-req":
					createTerminal(conn, channel)
				default:
					req.Reply(false, nil)
				}
				req.Reply(req.Type == "shell", nil)
			}
		}(requests)
	}
}

func createTerminal(conn *ssh.ServerConn, channel ssh.Channel) {
	termInstance := term.NewTerminal(channel, "> ")
	go func() {
		defer channel.Close()
		for {
			line, err := termInstance.ReadLine()
			if err != nil {
				fmt.Printf("ReadLine error: %s", err)
				break
			}
			switch line {
			case "whoami":
				termInstance.Write([]byte(execSomething(conn, []byte("whoami"))))
			case "":
			case "quit":
				termInstance.Write([]byte("Goodbye!\n"))
				channel.Close()
			default:
				termInstance.Write([]byte("Command not found\n"))
			}
		}
	}()
}

func execSomething(conn *ssh.ServerConn, payload []byte) string {
	switch string(payload) {
	case "whoami":
		return fmt.Sprintf("You are: %s", conn.Conn.User())
	default:
		return fmt.Sprintf("Command not Found: %s \n", string(payload))

	}
}
