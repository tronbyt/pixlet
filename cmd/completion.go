package cmd

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/schema"
)

var formats = []string{"webp", "gif", "avif"}

func completeInstallations(cmd *cobra.Command, deviceID string) ([]string, cobra.ShellCompDirective) {
	creds, err := resolveCommandAPICredentials(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var installations []string
	for i, err := range getInstallations(deviceID, creds) {
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		installations = append(installations, i.Id+"\t"+i.AppId)
	}
	return installations, cobra.ShellCompDirectiveNoFileComp
}

func completeDevices(cmd *cobra.Command) ([]string, cobra.ShellCompDirective) {
	creds, err := resolveCommandAPICredentials(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var devices []string
	for d, err := range getDevices(creds) {
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		devices = append(devices, d.ID+"\t"+d.DisplayName)
	}
	return devices, cobra.ShellCompDirectiveNoFileComp
}

func completeWebPLevel(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	var s []string
	for i := range 10 {
		s = append(s, strconv.Itoa(i))
	}
	return s, cobra.ShellCompDirectiveNoFileComp
}

func completeRender(_ *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return []string{"star"}, cobra.ShellCompDirectiveFilterFileExt
	}

	applet, err := runtime.NewAppletFromPath(
		args[0],
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(flags.NewMeta().Metadata),
	)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer applet.Close()

	if applet.Schema == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if id, _, ok := strings.Cut(toComplete, "="); ok {
		// Complete field options
		if idx := slices.IndexFunc(applet.Schema.Fields, func(f schema.SchemaField) bool {
			return f.ID == id
		}); idx != -1 {
			field := applet.Schema.Fields[idx]
			prefix := field.ID + "="
			switch field.Type {
			case "color":
				return []string{prefix + "#" + strings.TrimPrefix(field.Default, "#")}, cobra.ShellCompDirectiveNoFileComp
			case "onoff":
				return []string{prefix + "true", prefix + "false"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
			case "dropdown":
				s := make([]string, 0, len(field.Options))
				for _, option := range field.Options {
					cmp := prefix + option.Value + "\t" + option.Text
					if field.Default == option.Value && !strings.Contains(strings.ToLower(cmp), "(default)") {
						cmp += " (Default)"
					}
					s = append(s, cmp)
				}
				return s, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
			}
			if field.Default != "" {
				return []string{prefix + field.Default}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	// Complete field IDs
	s := make([]string, 0, len(applet.Schema.Fields))
	for _, field := range applet.Schema.Fields {
		alreadySet := slices.ContainsFunc(args[1:], func(s string) bool {
			return strings.HasPrefix(s, field.ID+"=")
		})
		if !alreadySet {
			s = append(s, fmt.Sprintf("%s=\t%s - %s - %s", field.ID, field.Type, field.Name, field.Description))
		}
	}
	return s, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoSpace
}
