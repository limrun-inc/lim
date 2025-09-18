/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
	"path/filepath"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/packages/param"
	"github.com/spf13/cobra"
	"go.jetify.com/typeid/v2"
)

var (
	downloadAssetName string
	outDir            string
)

func init() {
	PullCmd.PersistentFlags().StringVarP(&downloadAssetName, "name", "n", "", "Name of the asset.")
	PullCmd.PersistentFlags().StringVarP(&outDir, "output", "o", ".", "Output directory. Defaults to current directory.")
	RootCmd.AddCommand(PullCmd)
}

// PullCmd represents the push command
var PullCmd = &cobra.Command{
	Use:  "pull [ID or Name]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lim := cmd.Context().Value("lim").(limrun.Client)
		var id string
		_, err := typeid.Parse(args[0])
		if err != nil {
			downloadAssetName = args[0]
		} else {
			id = args[0]
		}
		var ass limrun.Asset
		switch {
		case id != "":
			fetched, err := lim.Assets.Get(cmd.Context(), id, limrun.AssetGetParams{
				IncludeDownloadURL: param.NewOpt(true),
			})
			if err != nil {
				if errors.IsUnauthenticated(err) {
					if err := config.Login(cmd.Context()); err != nil {
						return err
					}
					fmt.Println("You are logged in now")
					return nil
				}
				return fmt.Errorf("failed to get asset: %w", err)
			}
			ass = *fetched
		case downloadAssetName != "":
			fetched, err := lim.Assets.List(cmd.Context(), limrun.AssetListParams{
				NameFilter:         param.NewOpt(downloadAssetName),
				IncludeDownloadURL: param.NewOpt(true),
			})
			if err != nil {
				return fmt.Errorf("failed to get asset with name %s: %w", downloadAssetName, err)
			}
			if len(*fetched) == 0 {
				return fmt.Errorf("asset with name %s not found", downloadAssetName)
			}
			ass = (*fetched)[0]
		default:
			return fmt.Errorf("no asset id or name specified")
		}
		fullPath, err := filepath.Abs(filepath.Join(outDir, ass.Name))
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		resp, err := http.Get(ass.SignedDownloadURL)
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to download file: %s", string(b))
		}
		fmt.Printf("Pulling to %s\n", fullPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		file, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
		}
		defer file.Close()
		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			"",
		)
		_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}
		fmt.Printf("Done!\n")
		return nil
	},
}
