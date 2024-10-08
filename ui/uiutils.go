package ui

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
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

// PromptPasswordCLI handles secure password input for CLI mode with validation.
func PromptPasswordCLI() (string, error) {
	for {
		// Read password from the terminal securely
		fmt.Printf("Enter password:")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Newline after password input
		if err != nil {
			return "", err
		}

		password := string(passwordBytes)

		if len(password) == 0 {
			fmt.Println("Password cannot be blank!")
			continue // Re-prompt the user for a new password
		}

		// Validate the password
		if err := validatePassword(password); err != nil {
			// If the password is invalid, ask if the user wants to force a weak password
			for {
				fmt.Print("This is a weak password. Do you want to use it anyway? (y/n): ")
				choice, errormsg := term.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if errormsg != nil {
					return "", errormsg
				}

				choiceStr := strings.ToLower(strings.TrimSpace(string(choice)))

				// Check if the user entered 'y' or 'n'
				if choiceStr == "y" {
					fmt.Println("Warning: You are using a weak password.")
					break // Exit the loop and proceed with the weak password
				} else if choiceStr == "n" {
					fmt.Printf("Reason: %v\n", err)
					fmt.Println("Please try again. (ctrl+c to exit)")
					continue // Exit the inner loop and re-prompt for a new password
				} else {
					// Invalid input, re-prompt for a valid choice
					fmt.Println("Invalid input. Please enter 'y' or 'n'.")
				} 
			}
		}

		// Ask for password confirmation
		fmt.Printf("Confirm password: ")
		confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return "", err
		}

		confirmPassword := string(confirmPasswordBytes)

		// Check if the passwords match
		if password != confirmPassword {
			fmt.Println("Passwords do not match, please try again.")
			return "", fmt.Errorf("passwords do not match")
		}

		// Return the valid and confirmed password
		return password, nil
	}
}

// validatePassword checks if the password is at least 6 characters long and contains letters and numbers or special characters.
func validatePassword(password string) error {
	// Password must be at least 6 characters
	if len(password) < 6 {
		return fmt.Errorf("password should be at least 6 characters long")
	}

	// Check if the password contains at least one letter
	hasLetter, _ := regexp.MatchString("[a-zA-Z]", password)
	if !hasLetter {
		return fmt.Errorf("password should contain at least one letter")
	}

	// Check if the password contains at least one number or special character
	hasNumberOrSpecial, _ := regexp.MatchString("[0-9\\W]", password)
	if !hasNumberOrSpecial {
		return fmt.Errorf("password should contain at least one number or special character")
	}

	// If the password passes all checks, return nil
	return nil
}