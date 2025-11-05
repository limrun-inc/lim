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

package cmd

import (
	"fmt"
	"strings"

	"github.com/limrun-inc/lim/cmd/deleteCmd"
	"github.com/spf13/cobra"
)

// DeleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use:  "delete",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		switch strings.Split(id, "_")[0] {
		case "android":
			return deleteCmd.AndroidCmd.RunE(cmd, []string{id})
		case "ios":
			return deleteCmd.IOSCmd.RunE(cmd, []string{id})
		default:
			return fmt.Errorf("invalid id: %s", id)
		}
	},
}

func init() {
	DeleteCmd.AddCommand(deleteCmd.AndroidCmd)
	DeleteCmd.AddCommand(deleteCmd.IOSCmd)
	DeleteCmd.AddCommand(deleteCmd.AssetCmd)
	RootCmd.AddCommand(DeleteCmd)
}
