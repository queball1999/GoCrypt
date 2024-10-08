package ui

import (
	"flag"
	"fmt"
	"golang.org/x/term"
	"syscall"
)

// SetupFlags initializes the command-line flags and returns pointers to the flag variables.
func SetupFlags() (*string, *bool, *int) {
	outputDir := flag.String("output", "", "Specify the output directory")
	flag.StringVar(outputDir, "o", "", "Specify the output directory (alias: -o)")

	noUI := flag.Bool("no-ui", false, "Disable the GUI")
	flag.BoolVar(noUI, "n", false, "Disable the GUI (alias: -n)")

	layers := flag.Int("layers", 5, "Layers of encryption")
	flag.IntVar(layers, "l", 5, "Layers of encryption (alias: -l)")

	flag.Parse()

	return outputDir, noUI, layers
}

// promptPasswordCLI handles secure password input for CLI mode.
func PromptPasswordCLI() (string, error) {
	// Read password from the terminal securely
	fmt.Printf("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Newline after password input
	if err != nil {
		return "", err
	}

	// Ask for password confirmation
	fmt.Printf("Confirm password: ")
	confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}

	password := string(passwordBytes)
	confirmPassword := string(confirmPasswordBytes)

	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}