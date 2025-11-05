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

// AssetCmd represents the delete command for Asset
var AssetCmd = &cobra.Command{
	Use:     "asset [ID]",
	Aliases: []string{"ass", "assets"},
	Args:    cobra.ExactArgs(1),
	Short:   "Delete given Asset.",
	Long: `Examples:

$ lim delete asset asset_someid
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		lim := cmd.Context().Value("lim").(limrun.Client)
		if err := lim.Assets.Delete(cmd.Context(), id); err != nil {
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				fmt.Println("You are logged in now")
				return nil
			}
			return fmt.Errorf("failed to delete asset: %w", err)
		}
		fmt.Println("Deleted asset:", id)
		return nil
	},
}
