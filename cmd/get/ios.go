package get

import (
	"fmt"
	"github.com/limrun-inc/go-sdk"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// GetIOSCmd represents the get command
var GetIOSCmd = &cobra.Command{
	Use:     "ios [ID]",
	Aliases: []string{"i"},
	Short:   "Get all iOS instances, or specific instance if an ID is provided.",
	Long: `Examples:

Get all iOS instances:
$ lim get ios

Get a specific iOS instance:
$ lim get ios <ID>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var id string
		if len(args) > 1 {
			id = args[1]
		}
		var data [][]string
		table := tablewriter.NewWriter(cmd.OutOrStdout())
		lim := cmd.Context().Value("lim").(limrun.Client)
		table.Header([]string{"ID", "Name", "Region", "State"})
		var instances []limrun.IosInstance
		if id == "" {
			fetched, err := lim.IosInstances.List(cmd.Context(), limrun.IosInstanceListParams{
				State: limrun.IosInstanceListParamsStateReady,
			})
			if err != nil {
				return fmt.Errorf("failed to list ios instances: %w", err)
			}
			instances = *fetched
		} else {
			fetched, err := lim.IosInstances.Get(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("failed to get ios instance: %w", err)
			}
			instances = []limrun.IosInstance{*fetched}
		}
		data = make([][]string, len(instances))
		for i, instance := range instances {
			data[i] = []string{
				instance.Metadata.ID,
				instance.Metadata.DisplayName,
				instance.Spec.Region,
				instance.Status.State,
			}
		}
		if err := table.Bulk(data); err != nil {
			return err
		}
		return table.Render()
	},
}
