package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

type listOptions struct {
	apiToken string
	baseURL  string
}

func NewListCmd() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list [device ID]",
		Short: "Lists all apps installed on a Tronbyt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRun(cmd, args, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeDevices(cmd)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&opts.apiToken, "api-token", "t", opts.apiToken, "Tronbyt API token")
	_ = cmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&opts.baseURL, "url", "u", opts.baseURL, "base URL of Tronbyt API")
	_ = cmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)

	return cmd
}

func listRun(cmd *cobra.Command, args []string, opts *listOptions) error {
	deviceID := args[0]

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	defer w.Flush()

	creds, err := resolveAPICredentials(opts.baseURL, opts.apiToken)
	if err != nil {
		return err
	}

	for inst, err := range getInstallations(deviceID, creds) {
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%s\n", inst.Id, inst.AppId)
	}

	return nil
}

func getInstallations(deviceID string, creds apiCredentials) iter.Seq2[*tronbytapi.Installation, error] {
	return func(yield func(*tronbytapi.Installation, error) bool) {
		client := &http.Client{}
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("%s/v0/devices/%s/installations", creds.baseURL, deviceID), nil)
		if err != nil {
			yield(nil, fmt.Errorf("creating request: %w", err))
			return
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", creds.token))

		resp, err := client.Do(req)
		if err != nil {
			yield(nil, fmt.Errorf("listing installation: %w", err))
			return
		}

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			yield(nil, fmt.Errorf("tronbyt api error %s: %s", resp.Status, body))
			return
		}

		var installations tronbytapi.Installations
		err = json.Unmarshal(body, &installations)
		if err != nil {
			yield(nil, fmt.Errorf("decoding json: %s", body))
		}

		for _, installation := range installations.Installations {
			if !yield(&installation, nil) {
				return
			}
		}
	}
}
