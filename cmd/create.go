package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/community"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/tools/generator"
	"github.com/tronbyt/pixlet/tools/repo"
)

// NewCreateCmd prompts the user for info and generates a new app.
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		GroupID: groups.Applet,
		Short:   "Creates a new app",
		Long:    `This command will prompt for all of the information we need to generate a new Tronbyt app.`,
		RunE:    createRun,

		ValidArgsFunction: cobra.NoFileCompletions,
	}

	return cmd
}

func createRun(_ *cobra.Command, _ []string) error {
	// Get the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("app creation failed, something went wrong with your local filesystem: %w", err)
	}

	// Determine what type of app this is and what the root should be.
	root := cwd
	inAppsRepo := repo.IsInRepo(cwd, "apps", "community", "tidbyt")
	if inAppsRepo {
		if root, err = repo.Root(cwd); err != nil {
			return fmt.Errorf("app creation failed, something went wrong with your git repo: %w", err)
		}
	}

	// Prompt the user for input.
	app, err := community.ManifestPrompt()
	if err != nil {
		return fmt.Errorf("app creation, couldn't get user input: %w", err)
	}

	// Generate app.
	g, err := generator.NewGenerator(root, inAppsRepo)
	if err != nil {
		return fmt.Errorf("app creation failed %w", err)
	}
	absolutePath, err := g.GenerateApp(app)
	if err != nil {
		return fmt.Errorf("app creation failed: %w", err)
	}

	// Get the relative path from where the user started. Note, we're not
	// using the root here, given the root can be git repo specific.
	relativePath, err := filepath.Rel(cwd, absolutePath)
	if err != nil {
		return fmt.Errorf("app was created, but we don't know where: %w", err)
	}

	// Let the user know where the app is and how to use it.
	fmt.Println("")
	fmt.Println("App created at:")
	fmt.Printf("\t%s\n", absolutePath)
	fmt.Println("")
	fmt.Println("To start the app, run:")
	fmt.Printf("\tpixlet serve %s\n", relativePath)
	fmt.Println("")
	fmt.Println("For docs, head to:")
	fmt.Printf("\thttps://tidbyt.dev\n")
	return nil
}
