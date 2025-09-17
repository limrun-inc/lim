package get

import (
	"fmt"

	"github.com/limrun-inc/go-sdk"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// GetAndroidCmd represents the get command for Android
var GetAndroidCmd = &cobra.Command{
	Use:     "android [ID]",
	Aliases: []string{"a", "androids"},
	Short:   "Get all Android instances, or specific instance if an ID is provided.",
	Long: `Examples:

Get all Android instances:
$ lim get android

Get a specific Android instance:
$ lim get android <ID>
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
		var instances []limrun.AndroidInstance
		if id == "" {
			fetched, err := lim.AndroidInstances.List(cmd.Context(), limrun.AndroidInstanceListParams{
				State: limrun.AndroidInstanceListParamsStateReady,
			})
			if err != nil {
				return fmt.Errorf("failed to list android instances: %w", err)
			}
			instances = *fetched
		} else {
			fetched, err := lim.AndroidInstances.Get(cmd.Context(), id)
			if err != nil {
				return fmt.Errorf("failed to get android instance: %w", err)
			}
			instances = []limrun.AndroidInstance{*fetched}
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
