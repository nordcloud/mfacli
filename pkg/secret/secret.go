package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

type SecretValue struct {
	value string
	isSet bool
}

func (s *SecretValue) String() string {
	return s.value
}

func (s *SecretValue) Set(arg string) error {
	var err error

	if strings.HasPrefix(arg, "pass:") && len(arg) > 5 {
		s.value = arg[5:]
	} else if strings.HasPrefix(arg, "file:") && len(arg) > 5 {
		err = s.setFromFile(arg[5:])
	} else if strings.HasPrefix(arg, "env:") && len(arg) > 4 {
		err = s.setFromEnv(arg[4:])
	} else {
		err = fmt.Errorf("Invalid secret format")
	}

	if err == nil {
		s.isSet = true
	}

	return err
}

func (s *SecretValue) Type() string {
	return "secret"
}

func (s *SecretValue) IsSet() bool {
	return s.isSet
}

func (s *SecretValue) setFromFile(filename string) error {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	s.value = string(body)
	return nil
}

func (s *SecretValue) setFromEnv(env string) error {
	val, ok := os.LookupEnv(env)
	if !ok {
		return fmt.Errorf("Secret env %s is not set", env)
	}
	s.value = val
	return nil
}

func ReadSecret(prompt, confirmPrompt string) (string, error) {
	value, err := readSecret(prompt)
	if confirmPrompt == "" || err != nil {
		return value, err
	}

	var confirmation string
	for value != confirmation {
		confirmation, err = readSecret(confirmPrompt)
		if err != nil {
			return "", err
		}
	}
	return value, nil
}

func readSecret(prompt string) (string, error) {
	for {
		fmt.Fprint(os.Stderr, prompt)
		data, err := terminal.ReadPassword(0)
		fmt.Fprint(os.Stderr, "\n")
		if err != nil {
			return "", err
		}

		if len(data) > 0 {
			return string(data), nil
		}

		fmt.Fprintln(os.Stderr, "Empty input is not allowed")
	}
}
