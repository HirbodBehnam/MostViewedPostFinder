package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

// noSignUp can be embedded to prevent signing up.
type noSignUp struct{}

func (c noSignUp) SignUp(context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("not implemented")
}

func (c noSignUp) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

// SimpleAuth implements authentication via terminal.
type SimpleAuth struct {
	noSignUp
	PhoneNumber string
}

func (a SimpleAuth) Phone(context.Context) (string, error) {
	return a.PhoneNumber, nil
}

func (a SimpleAuth) Password(context.Context) (string, error) {
	return Prompt("Enter 2FA password: ", true)
}

func (a SimpleAuth) Code(context.Context, *tg.AuthSentCode) (string, error) {
	return Prompt("Enter code: ", false)
}

// Prompt asks user for an input
// If password is true it means that the characters won't be echoed
func Prompt(message string, password bool) (result string, err error) {
	fmt.Print(message)
	if password {
		var readData []byte
		readData, err = term.ReadPassword(int(syscall.Stdin))
		result = string(readData)
	} else {
		result, err = bufio.NewReader(os.Stdin).ReadString('\n')
	}
	return strings.TrimSpace(result), err
}
