package deleteCmd

import (
	"fmt"

	"github.com/limrun-inc/go-sdk"
	"github.com/spf13/cobra"
)

// IOSCmd represents the delete command for iOS
var IOSCmd = &cobra.Command{
	Use:     "ios [ID]",
	Aliases: []string{"i", "ios"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete given iOS instance.",
	Long: `Examples:

$ lim delete ios <ID>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		lim := cmd.Context().Value("lim").(limrun.Client)
		if err := lim.IosInstances.Delete(cmd.Context(), id); err != nil {
			return fmt.Errorf("failed to delete iOS instance: %w", err)
		}
		fmt.Println("Deleted iOS instance:", id)
		return nil
	},
}
