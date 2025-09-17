package cmd

import (
	"fmt"

	"github.com/limrun-inc/go-sdk"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [android|ios|asset] [ID]",
	Args:  cobra.MinimumNArgs(1),
	Short: "Get all instances for the given kind, or specific instance if an ID is provided.",
	Long: `Examples:

Get all Android instances:
$ lim get android

Get a specific Android instance:
$ lim get android <ID>

Get all iOS instances:
$ lim get ios

Get a specific iOS instance:
$ lim get ios <ID>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		kind := args[0]
		var id string
		if len(args) > 1 {
			id = args[1]
		}
		var data [][]string
		lim := cmd.Context().Value("lim").(limrun.Client)
		switch kind {
		case "android":
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
					return fmt.Errorf("failed to list android instances: %w", err)
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
		case "ios":
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
					return fmt.Errorf("failed to list ios instances: %w", err)
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
		default:
			return fmt.Errorf("invalid kind: %s", kind)
		}
		table := tablewriter.NewWriter(cmd.OutOrStdout())
		table.Header([]string{"ID", "Name", "Region", "State"})
		if err := table.Bulk(data); err != nil {
			return err
		}
		return table.Render()
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
