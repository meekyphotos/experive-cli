package utils

import (
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

type PasswordProvider interface {
	ReadPassword() (string, error)
}

type TerminalPasswordReader struct{}

func (pr TerminalPasswordReader) ReadPassword() (string, error) {
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	return string(password), err
}
