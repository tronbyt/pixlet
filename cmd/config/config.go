package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration for Pixlet.",
	}

	cmd.AddCommand(
		NewSetCmd(),
		NewGetCmd(),
	)

	return cmd
}

const (
	URLKey      = "url"
	TokenKey    = "token"
	APITokenEnv = "PIXLET_TOKEN"
)

var Config = viper.New()

func init() {
	if ucd, err := os.UserConfigDir(); err == nil {
		configPath := filepath.Join(ucd, "tronbyt")

		if err := os.MkdirAll(configPath, os.ModePerm); err == nil {
			Config.AddConfigPath(configPath)
		}
	}

	Config.SetConfigName("config")
	Config.SetConfigType("yaml")
	Config.SetConfigPermissions(0600)

	Config.SafeWriteConfig()
	Config.ReadInConfig()
}

var ErrNoURL = fmt.Errorf("Tronbyt URL not set. Use `tronbyt config set url <url>` to set it.")

func GetURL() (string, error) {
	if url := Config.GetString(URLKey); url != "" {
		return url, nil
	}
	return "", ErrNoURL
}

var ErrNoToken = fmt.Errorf("Tronbyt API token not set. Use `tronbyt config set token <token>` to set it.")

func GetToken() (string, error) {
	if token := os.Getenv(APITokenEnv); token != "" {
		return token, nil
	}
	if token := Config.GetString(TokenKey); token != "" {
		return token, nil
	}
	return "", ErrNoToken
}
