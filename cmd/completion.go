package cmd

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/schema"
)

func completeInstallations(cmd *cobra.Command, creds *flags.APICredentials, deviceID string) ([]string, cobra.ShellCompDirective) {
	client, err := tronbytapi.NewClient(creds.URL, creds.APIToken)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var installations []string
	for i, err := range client.GetInstallations(cmd.Context(), deviceID) {
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		installations = append(installations, i.ID+"\t"+i.AppID)
	}
	return installations, cobra.ShellCompDirectiveNoFileComp
}

func completeDevices(cmd *cobra.Command, creds *flags.APICredentials) ([]string, cobra.ShellCompDirective) {
	client, err := tronbytapi.NewClient(creds.URL, creds.APIToken)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var devices []string
	for d, err := range client.GetDevices(cmd.Context()) {
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		devices = append(devices, d.ID+"\t"+d.DisplayName)
	}
	return devices, cobra.ShellCompDirectiveNoFileComp
}

func completeWebPLevel(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	s := make([]string, 0, 10)
	for i := range 10 {
		s = append(s, strconv.Itoa(i))
	}
	return s, cobra.ShellCompDirectiveNoFileComp
}

func completeRender(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) == 0 && !strings.Contains(toComplete, "=") {
		return []string{"star"}, cobra.ShellCompDirectiveFilterFileExt
	}

	configFile, _ := cmd.Flags().GetString("config")
	path, config, args, err := loadConfig(configFile, args)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	applet, err := runtime.NewAppletFromPath(
		cmd.Context(),
		path,
		runtime.WithPrintDisabled(),
		runtime.WithCanvasMeta(flags.NewMeta().Metadata),
	)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer func() { _ = applet.Close() }()

	if applet.Schema == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if id, val, ok := strings.Cut(toComplete, "="); ok {
		// Complete field options
		if idx := slices.IndexFunc(applet.Schema.Fields, func(f schema.SchemaField) bool {
			return f.ID == id
		}); idx != -1 {
			field := applet.Schema.Fields[idx]
			prefix := field.ID + "="
			switch field.Type {
			case "typeahead", "locationbased":
				if field.Handler == "" || val == "" {
					return nil, cobra.ShellCompDirectiveNoFileComp
				}

				result, err := applet.CallSchemaHandler(cmd.Context(), field.Handler, val, config)
				if err != nil {
					return nil, cobra.ShellCompDirectiveError
				}

				var options []schema.SchemaOption
				if err := json.Unmarshal([]byte(result), &options); err != nil {
					return nil, cobra.ShellCompDirectiveError
				}

				s := make([]string, 0, len(options))
				for _, opt := range options {
					s = append(s, prefix+opt.Value+"\t"+opt.Text)
				}
				return s, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder
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

	isAlreadySet := func(field schema.SchemaField) bool {
		return slices.ContainsFunc(args, func(s string) bool {
			return strings.HasPrefix(s, field.ID+"=")
		})
	}

	// Complete field IDs
	s := make([]string, 0, len(applet.Schema.Fields))
	addField := func(field schema.SchemaField) {
		s = append(s, fmt.Sprintf("%s=\t%s - %s - %s", field.ID, field.Type, field.Name, field.Description))
	}

	for _, field := range applet.Schema.Fields {
		if isAlreadySet(field) {
			continue
		}

		switch field.Type {
		case "generated":
			if field.Handler == "" {
				continue
			}

			param, ok := config[field.Source].(string)
			if !ok || param == "" {
				continue
			}

			result, err := applet.CallSchemaHandler(cmd.Context(), field.Handler, param, config)
			if err != nil {
				continue
			}

			var genSchema schema.Schema
			if err := json.Unmarshal([]byte(result), &genSchema); err != nil {
				continue
			}

			s = slices.Grow(s, len(genSchema.Fields))

			for _, genField := range genSchema.Fields {
				if !isAlreadySet(genField) {
					addField(genField)
				}
			}
		default:
			addField(field)
		}
	}
	return s, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveKeepOrder | cobra.ShellCompDirectiveNoSpace
}
