package flags

import (
	"github.com/spf13/cobra"
)

const (
	FlagURL      = "url"
	FlagAPIToken = "api-token"
)

func NewAPICredentials() *APICredentials {
	return &APICredentials{}
}

type APICredentials struct {
	URL      string
	APIToken string
}

func (a *APICredentials) Register(cmd *cobra.Command) {
	fs := cmd.Flags()

	fs.StringVarP(&a.APIToken, FlagAPIToken, "t", a.APIToken, "Tronbyt API token")
	_ = cmd.RegisterFlagCompletionFunc(FlagAPIToken, cobra.NoFileCompletions)

	fs.StringVarP(&a.URL, FlagURL, "u", a.URL, "base URL of Tronbyt API")
	_ = cmd.RegisterFlagCompletionFunc(FlagURL, cobra.NoFileCompletions)
}
