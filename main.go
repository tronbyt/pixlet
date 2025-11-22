package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd"
	"github.com/tronbyt/pixlet/cmd/community"
	"github.com/tronbyt/pixlet/cmd/private"
)

var (
	rootCmd = &cobra.Command{
		Use:          "pixlet",
		Short:        "pixel graphics rendering",
		Long:         "Pixlet renders graphics for pixel devices, like Tidbyt",
		SilenceUsage: true,
	}
)

func init() {
	rootCmd.AddCommand(cmd.ApiCmd)
	rootCmd.AddCommand(cmd.CheckCmd)
	rootCmd.AddCommand(cmd.CreateCmd)
	rootCmd.AddCommand(cmd.DeleteCmd)
	rootCmd.AddCommand(cmd.DevicesCmd)
	rootCmd.AddCommand(cmd.EncryptCmd)
	rootCmd.AddCommand(cmd.FormatCmd)
	rootCmd.AddCommand(cmd.LintCmd)
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.LoginCmd)
	rootCmd.AddCommand(cmd.ProfileCmd)
	rootCmd.AddCommand(cmd.PushCmd)
	rootCmd.AddCommand(cmd.RenderCmd)
	rootCmd.AddCommand(cmd.SchemaCmd)
	rootCmd.AddCommand(cmd.ServeCmd)
	rootCmd.AddCommand(cmd.SetAuthCmd)
	rootCmd.AddCommand(cmd.VersionCmd)
	rootCmd.AddCommand(community.CommunityCmd)
	rootCmd.AddCommand(private.PrivateCmd)
}

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.TimeOnly,
			NoColor:    !isatty.IsTerminal(os.Stderr.Fd()),
		}),
	))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
