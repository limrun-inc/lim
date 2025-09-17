/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/packages/param"
	"github.com/spf13/cobra"
)

var (
	downloadAssetName string
)

func init() {
	DownloadAssetCmd.PersistentFlags().StringVarP(&downloadAssetName, "name", "n", "", "Name of the asset")
	DownloadCmd.AddCommand(DownloadAssetCmd)
	RootCmd.AddCommand(DownloadCmd)
}

// DownloadAssetCmd represents the download asset command
var DownloadAssetCmd = &cobra.Command{
	Use: "asset [ID]",
	RunE: func(cmd *cobra.Command, args []string) error {
		lim := cmd.Context().Value("lim").(limrun.Client)
		var id string
		if len(args) > 0 {
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
		fileName := filepath.Base(ass.Name)
		resp, err := http.Get(ass.SignedDownloadURL)
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to download file: %s", string(b))
		}
		fmt.Printf("Downloading file %s\n", fileName)
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fileName, err)
		}
		defer file.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", fileName, err)
		}
		fmt.Printf("Successfully downloaded %s\n", fileName)
		return nil
	},
}

var DownloadCmd = &cobra.Command{
	Use: "download",
	Run: func(cmd *cobra.Command, args []string) {},
}
