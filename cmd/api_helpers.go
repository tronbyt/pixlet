package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/config"
)

type apiCredentials struct {
	baseURL string
	token   string
}

func resolveAPICredentials(baseURL, token string) (apiCredentials, error) {
	var err error
	if baseURL == "" {
		if baseURL, err = config.GetURL(); err != nil {
			return apiCredentials{}, err
		}
	}
	if token == "" {
		if token, err = config.GetToken(); err != nil {
			return apiCredentials{}, err
		}
	}

	return apiCredentials{
		baseURL: baseURL,
		token:   token,
	}, nil
}

func resolveCommandAPICredentials(cmd *cobra.Command) (apiCredentials, error) {
	baseURL, err := cmd.Flags().GetString("url")
	if err != nil {
		return apiCredentials{}, err
	}

	token, err := cmd.Flags().GetString("api-token")
	if err != nil {
		return apiCredentials{}, err
	}

	return resolveAPICredentials(baseURL, token)
}
