/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"

	"github.com/spf13/cobra"

	limrun "github.com/limrun-inc/go-sdk"
)

var (
	uploadAssetName string
)

func init() {
	PushCmd.PersistentFlags().StringVarP(&uploadAssetName, "name", "n", "", "Name of the asset. Defaults to file name.")
	RootCmd.AddCommand(PushCmd)
}

// PushCmd represents the upload asset command
var PushCmd = &cobra.Command{
	Use:  "push [file path]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lim := cmd.Context().Value("lim").(limrun.Client)
		params := limrun.AssetGetOrUploadParams{
			Path: args[0],
		}
		if uploadAssetName != "" {
			params.Name = &uploadAssetName
		}
		ass, err := lim.Assets.GetOrUpload(cmd.Context(), params)
		if err != nil {
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				fmt.Println("You are logged in now")
				return nil
			}
			return err
		}
		fmt.Printf("Asset %s with ID of %s is ready", ass.Name, ass.ID)
		return nil
	},
}
