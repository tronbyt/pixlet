package community

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/manifest"
)

func NewValidateManifestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate-manifest [path]",
		Short:   "Validates an app manifest is ready for publishing",
		Example: `  pixlet community validate-manifest manifest.yaml`,
		Long: `This command determines if your app manifest is configured properly by
validating the contents of each field.`,
		Args:              cobra.MaximumNArgs(1),
		RunE:              ValidateManifest,
		ValidArgsFunction: cobra.FixedCompletions([]string{"yaml"}, cobra.ShellCompDirectiveFilterFileExt),
	}
	return cmd
}

func ValidateManifest(_ *cobra.Command, args []string) error {
	path := manifest.ManifestFileName
	if len(args) != 0 {
		path = args[0]
	}

	if filepath.Base(path) != manifest.ManifestFileName {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if !info.IsDir() {
			return fmt.Errorf("supplied manifest must be named %s", manifest.ManifestFileName)
		}

		path = filepath.Join(path, manifest.ManifestFileName)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("couldn't open manifest: %w", err)
	}
	defer func() { _ = f.Close() }()

	m, err := manifest.LoadManifest(f)
	if err != nil {
		return fmt.Errorf("couldn't load manifest: %w", err)
	}

	err = m.Validate()
	if err != nil {
		return fmt.Errorf("couldn't validate manifest: %w", err)
	}

	return nil
}
