/*
Copyright 2025 Limrun, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deleteCmd

import (
	"fmt"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"

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
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				fmt.Println("You are logged in now")
				return nil
			}
			return fmt.Errorf("failed to delete iOS instance: %w", err)
		}
		fmt.Println("Deleted iOS instance:", id)
		return nil
	},
}
