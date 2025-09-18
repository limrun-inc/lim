/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/packages/param"

	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
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
		f, err := os.Stat(args[0])
		if err != nil {
			return err
		}
		name := filepath.Base(args[0])
		if uploadAssetName != "" {
			name = uploadAssetName
		}
		fmt.Printf("Name: %s\n", name)
		bar := progressbar.DefaultBytes(
			f.Size(),
			"",
		)
		params := limrun.AssetGetOrUploadParams{
			Path:           args[0],
			ProgressWriter: bar,
			Name:           param.NewOpt(name),
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
		if err := bar.Close(); err != nil {
			return err
		}
		fmt.Printf("ID: %s\n", ass.ID)
		fmt.Printf("\nDone!\n")
		return nil
	},
}
