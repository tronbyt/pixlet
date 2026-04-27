package community

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/manifest"
)

func NewCreateManifestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "create-manifest [PATH]",
		Short:             "Creates an app manifest from a prompt",
		Example:           `  pixlet community create-manifest manifest.yaml`,
		Long:              `This command creates an app manifest by asking a series of prompts.`,
		Args:              cobra.MaximumNArgs(1),
		RunE:              CreateManifest,
		ValidArgsFunction: cobra.FixedCompletions([]string{"yaml"}, cobra.ShellCompDirectiveFilterFileExt),
	}
	return cmd
}

func CreateManifest(_ *cobra.Command, args []string) error {
	path := manifest.ManifestFileName
	if len(args) != 0 {
		path = args[0]
	}

	if filepath.Base(path) != manifest.ManifestFileName {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}

		if !info.IsDir() {
			return fmt.Errorf("supplied manifest must be named %s", manifest.ManifestFileName)
		}

		path = filepath.Join(path, manifest.ManifestFileName)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("couldn't open manifest: %w", err)
	}
	defer func() { _ = f.Close() }()

	m, err := ManifestPrompt()
	if err != nil {
		return fmt.Errorf("prompting: %w", err)
	}

	err = m.WriteManifest(f)
	if err != nil {
		return fmt.Errorf("couldn't write manifest: %w", err)
	}

	return nil
}

func ManifestPrompt() (*manifest.Manifest, error) {
	// Get the name of the app.
	namePrompt := promptui.Prompt{
		Label:    "Name (what do you want to call your app?)",
		Validate: manifest.ValidateName,
	}
	name, err := namePrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("app creation failed %w", err)
	}

	// Get the summary of the app.
	summaryPrompt := promptui.Prompt{
		Label:    "Summary (what's the short and sweet of what this app does?)",
		Validate: manifest.ValidateSummary,
	}
	summary, err := summaryPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("app creation failed %w", err)
	}

	// Get the description of the app.
	descPrompt := promptui.Prompt{
		Label:    "Description (what's the long form of what this app does?)",
		Validate: manifest.ValidateDesc,
	}
	desc, err := descPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("app creation failed %w", err)
	}

	// Get the author of the app.
	authorPrompt := promptui.Prompt{
		Label:    "Author (your name or your Github handle)",
		Validate: manifest.ValidateAuthor,
	}
	author, err := authorPrompt.Run()
	if err != nil {
		return nil, fmt.Errorf("app creation failed %w", err)
	}

	return &manifest.Manifest{
		ID:      manifest.GenerateID(name),
		Name:    name,
		Summary: summary,
		Desc:    desc,
		Author:  author,
	}, nil
}
