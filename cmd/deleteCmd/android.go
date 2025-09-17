package deleteCmd

import (
	"fmt"

	"github.com/limrun-inc/go-sdk"
	"github.com/spf13/cobra"
)

// AndroidCmd represents the delete command for Android
var AndroidCmd = &cobra.Command{
	Use:     "android [ID]",
	Aliases: []string{"a", "androids"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete given Android instance.",
	Long: `Examples:

$ lim delete android <ID>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		lim := cmd.Context().Value("lim").(limrun.Client)
		if err := lim.AndroidInstances.Delete(cmd.Context(), id); err != nil {
			return fmt.Errorf("failed to delete Android instance: %w", err)
		}
		fmt.Println("Deleted Android instance:", id)
		return nil
	},
}
