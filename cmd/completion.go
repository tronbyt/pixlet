package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
)

var formats = []string{"webp", "gif", "avif"}

func completeInstallations(deviceID string) ([]string, cobra.ShellCompDirective) {
	var installations []string
	for i, err := range getInstallations(deviceID) {
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		installations = append(installations, i.Id+"\t"+i.AppId)
	}
	return installations, cobra.ShellCompDirectiveNoFileComp
}

func completeDevices() ([]string, cobra.ShellCompDirective) {
	var devices []string
	for d, err := range getDevices() {
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
