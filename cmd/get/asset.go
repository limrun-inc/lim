package get

import (
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/packages/param"
)

var (
	assetName          string
	includeDownloadUrl bool
	includeUploadUrl   bool
)

func init() {
	GetAssetCmd.PersistentFlags().StringVar(&assetName, "name", "", "The name of the asset")
	GetAssetCmd.PersistentFlags().BoolVar(&includeDownloadUrl, "download-url", false, "Include a download URL in the response")
	GetAssetCmd.PersistentFlags().BoolVar(&includeUploadUrl, "upload-url", false, "Include an upload URL in the response")
}

// GetAssetCmd represents the get command for Assets
var GetAssetCmd = &cobra.Command{
	Use:     "asset [ID]",
	Aliases: []string{"ass", "assets"},
	Short:   "Get all assets, or specific asset if an ID is provided.",
	Long: `Examples:

Get all asset:
$ lim get assets

Get a specific asset:
$ lim get asset <ID>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var id string
		if len(args) > 1 {
			id = args[1]
		}
		var data [][]string
		table := tablewriter.NewWriter(cmd.OutOrStdout())
		lim := cmd.Context().Value("lim").(limrun.Client)
		var instances []limrun.Asset
		if id == "" {
			params := limrun.AssetListParams{
				IncludeDownloadURL: param.NewOpt(includeDownloadUrl),
				IncludeUploadURL:   param.NewOpt(includeUploadUrl),
			}
			if assetName != "" {
				params.NameFilter = param.NewOpt(assetName)
			}
			fetched, err := lim.Assets.List(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("failed to list assets: %w", err)
			}
			instances = *fetched
		} else {
			fetched, err := lim.Assets.Get(cmd.Context(), id, limrun.AssetGetParams{})
			if err != nil {
				return fmt.Errorf("failed to get asset: %w", err)
			}
			instances = []limrun.Asset{*fetched}
		}
		data = make([][]string, len(instances))
		for i, instance := range instances {
			data[i] = []string{
				instance.ID,
				instance.Name,
				instance.Md5,
			}
			if includeDownloadUrl {
				data[i] = append(data[i], instance.SignedDownloadURL)
			}
			if includeUploadUrl {
				data[i] = append(data[i], instance.SignedUploadURL)
			}
		}
		cols := []string{"ID", "Name", "MD5"}
		if includeDownloadUrl {
			cols = append(cols, "Download URL")
		}
		if includeUploadUrl {
			cols = append(cols, "Upload URL")
		}
		table.Header(cols)
		if err := table.Bulk(data); err != nil {
			return err
		}
		return table.Render()
	},
}
