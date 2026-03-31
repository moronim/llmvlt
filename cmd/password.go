package cmd

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("could not read password: %w", err)
	}
	return string(pw), nil
}

func promptPasswordConfirm() (string, error) {
	pw, err := promptPassword("Enter new vault password: ")
	if err != nil {
		return "", err
	}
	confirm, err := promptPassword("Confirm password: ")
	if err != nil {
		return "", err
	}
	if pw != confirm {
		return "", fmt.Errorf("passwords do not match")
	}
	return pw, nil
}
