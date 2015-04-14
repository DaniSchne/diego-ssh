package authenticators

import "golang.org/x/crypto/ssh"

type PublicKeyAuthenticator interface {
	Authenticate(conn ssh.ConnMetadata, publicKey ssh.PublicKey) (*ssh.Permissions, error)
	PublicKey() ssh.PublicKey
	User() string
}

//go:generate counterfeiter -o fake_authenticators/fake_password_authenticator.go . PasswordAuthenticator
type PasswordAuthenticator interface {
	Authenticate(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error)
}