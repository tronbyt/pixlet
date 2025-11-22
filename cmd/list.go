package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/config"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

var (
	listURL string
)

func init() {
	ListCmd.Flags().StringVarP(&apiToken, "api-token", "t", "", "Tronbyt API token")
	_ = ListCmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	ListCmd.Flags().StringVarP(&listURL, "url", "u", "", "base URL of Tronbyt API")
	_ = ListCmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)
}

var ListCmd = &cobra.Command{
	Use:   "list [device ID]",
	Short: "Lists all apps installed on a Tronbyt",
	Args:  cobra.MinimumNArgs(1),
	RunE:  listInstallationsRun,
	ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeDevices()
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func listInstallationsRun(cmd *cobra.Command, args []string) error {
	deviceID := args[0]

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 22, 8, 0, '\t', 0)
	defer w.Flush()

	for inst, err := range getInstallations(deviceID) {
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\t%s\n", inst.AppId, inst.Id)
	}

	return nil
}

func getInstallations(deviceID string) iter.Seq2[*tronbytapi.Installation, error] {
	return func(yield func(*tronbytapi.Installation, error) bool) {
		if listURL == "" {
			var err error
			if listURL, err = config.GetURL(); err != nil {
				yield(nil, err)
				return
			}
		}

		if apiToken == "" {
			var err error
			if apiToken, err = config.GetToken(); err != nil {
				yield(nil, err)
				return
			}
		}

		client := &http.Client{}
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf("%s/v0/devices/%s/installations", listURL, deviceID), nil)
		if err != nil {
			yield(nil, fmt.Errorf("creating request: %w", err))
			return
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

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
