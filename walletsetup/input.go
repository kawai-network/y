package walletsetup

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/term"
)

// InputReader handles secure CLI input.
type InputReader struct {
	scanner *bufio.Scanner
}

// NewInputReader creates a new InputReader.
func NewInputReader() *InputReader {
	return &InputReader{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// ReadLine reads a line of input with a prompt.
func (r *InputReader) ReadLine(prompt string) (string, error) {
	fmt.Print(prompt)
	if r.scanner.Scan() {
		return r.scanner.Text(), nil
	}
	return "", r.scanner.Err()
}

// ReadInt reads an integer with validation and retry.
func (r *InputReader) ReadInt(prompt string, minVal, maxVal int) (int, error) {
	fmt.Print(prompt)
	line, err := r.ReadLine("")
	if err != nil {
		return 0, err
	}
	var val int
	_, err = fmt.Sscanf(line, "%d", &val)
	if err != nil || val < minVal || val > maxVal {
		return 0, fmt.Errorf("invalid input: must be between %d and %d", minVal, maxVal)
	}
	return val, nil
}

// ReadPassword reads a password with hidden input and confirmation.
func (r *InputReader) ReadPassword(prompt string) (string, error) {
	for {
		fmt.Print(prompt)
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println()

		passStr := string(password)
		if len(passStr) < 8 {
			fmt.Println("Password must be at least 8 characters. Please try again.")
			continue
		}

		fmt.Print("Confirm password: ")
		confirm, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("failed to read password confirmation: %w", err)
		}
		fmt.Println()

		if passStr != string(confirm) {
			fmt.Println("Passwords do not match. Please try again.")
			continue
		}

		return passStr, nil
	}
}

// ReadSecureLine reads sensitive input with hidden display.
func (r *InputReader) ReadSecureLine(prompt string) (string, error) {
	fmt.Print(prompt)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(data), nil
}
