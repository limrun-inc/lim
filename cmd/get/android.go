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

package get

import (
	"fmt"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"

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
				if errors.IsUnauthenticated(err) {
					if err := config.Login(cmd.Context()); err != nil {
						return err
					}
					fmt.Println("You are logged in now")
					return nil
				}
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
