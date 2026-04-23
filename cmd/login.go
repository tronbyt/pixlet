package cmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/config"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

func NewLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "login",
		GroupID: groups.Tronbyt,
		Short:   "Log in to your Tronbyt instance",
		Long:    "This command will prompt for your Tronbyt URL and API key, store them locally, then verify by listing your devices.",
		RunE:    loginRun,

		ValidArgsFunction: cobra.NoFileCompletions,
	}

	return cmd
}

func loginRun(cmd *cobra.Command, _ []string) error {
	config.InitConfig()

	// Prompt for Tronbyt URL.
	existingURL := config.Config.GetString(config.URLKey)
	urlPrompt := promptui.Prompt{
		Label:   "Tronbyt URL",
		Default: existingURL,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("URL is required")
			}
			return nil
		},
	}
	tronbytURL, err := urlPrompt.Run()
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	tronbytURL = strings.TrimRight(strings.TrimSpace(tronbytURL), "/")

	// Prompt for API key.
	fmt.Printf("You can find your API key at %s/auth/edit\n", tronbytURL)
	tokenPrompt := promptui.Prompt{
		Label: "API Key",
		Mask:  '*',
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("API key is required")
			}
			return nil
		},
	}
	apiToken, err := tokenPrompt.Run()
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	apiToken = strings.TrimSpace(apiToken)

	// Verify credentials before saving.
	fmt.Println("Verifying credentials...")
	fmt.Println()

	client, err := tronbytapi.NewClient(tronbytURL, apiToken)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	var count int
	for d, err := range client.GetDevices(cmd.Context()) {
		if err != nil {
			return fmt.Errorf("fetching devices: %w", err)
		}
		count++
		fmt.Printf("  %s (%s)\n", d.ID, d.DisplayName)
	}

	// Save to config.
	config.Config.Set(config.URLKey, tronbytURL)
	config.Config.Set(config.TokenKey, apiToken)
	if err := config.Config.WriteConfig(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println()
	if count == 0 {
		fmt.Println("Login successful! No devices found.")
	} else {
		fmt.Printf("Login successful! Found %d device(s).\n", count)
	}

	return nil
}
