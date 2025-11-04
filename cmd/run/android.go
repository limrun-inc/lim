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

package run

import (
	"fmt"
	"github.com/limrun-inc/go-sdk/packages/param"
	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
	"github.com/schollz/progressbar/v3"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/tunnel"
	"github.com/spf13/cobra"
)

var (
	adbPath      string
	connect      bool
	stream       bool
	deleteOnExit bool

	assetNamesToInstall []string
	localAppsToInstall  []string
)

func init() {
	AndroidCmd.PersistentFlags().StringVar(&adbPath, "adb-path", "adb", "Optional path to the adb binary, defaults to `adb`")
	AndroidCmd.PersistentFlags().BoolVar(&connect, "connect", true, "Connect to the Android instance, e.g. start ADB tunnel. Default is true.")
	AndroidCmd.PersistentFlags().BoolVar(&stream, "stream", true, "Stream the Android instance for control. Default is true. Connect flag must be true.")
	AndroidCmd.PersistentFlags().BoolVar(&deleteOnExit, "rm", false, "Delete the instance on exit. Default is false.")
	AndroidCmd.PersistentFlags().StringArrayVar(&assetNamesToInstall, "install-asset", []string{}, "List of asset names to install. It will return error if they are not already uploaded. Asset names that will be installed together should be separated by comma.")
	AndroidCmd.PersistentFlags().StringArrayVar(&localAppsToInstall, "install", []string{}, "List of local app files to install. If not uploaded already, they will be uploaded to the asset storage first. Files that will be installed together should be separated by comma.")
}

// AndroidCmd represents the connect command for Android
var AndroidCmd = &cobra.Command{
	Use:   "android",
	Short: "Creates a new Android instance, connects and starts streaming.",
	RunE: func(cmd *cobra.Command, args []string) error {
		lim := cmd.Context().Value("lim").(limrun.Client)
		var finalAssetNamesToInstall [][]string
		if len(assetNamesToInstall) > 0 {
			for _, assetName := range assetNamesToInstall {
				var arr []string
				for _, n := range strings.Split(assetName, ",") {
					if n == "" {
						continue
					}
					arr = append(arr, n)
				}
				finalAssetNamesToInstall = append(finalAssetNamesToInstall, arr)
			}
		}
		if len(localAppsToInstall) > 0 {
			for _, appPaths := range localAppsToInstall {
				var assetNamesForSingleApp []string
				for _, singleApkPath := range strings.Split(appPaths, ",") {
					if singleApkPath == "" {
						continue
					}
					f, err := os.Stat(singleApkPath)
					if err != nil {
						return err
					}
					fmt.Printf("%s\n", filepath.Base(singleApkPath))
					bar := progressbar.DefaultBytes(
						f.Size(),
						"",
					)
					ass, err := lim.Assets.GetOrUpload(cmd.Context(), limrun.AssetGetOrUploadParams{
						Name:           param.NewOpt(filepath.Base(singleApkPath)),
						Path:           singleApkPath,
						ProgressWriter: bar,
					})
					if err != nil {
						if errors.IsUnauthenticated(err) {
							if err := config.Login(cmd.Context()); err != nil {
								return err
							}
							fmt.Println("You are logged in now")
							return nil
						}
						return fmt.Errorf("failed to upload app at %s: %w", singleApkPath, err)
					}
					if err := bar.Close(); err != nil {
						return err
					}
					assetNamesForSingleApp = append(assetNamesForSingleApp, ass.Name)
				}
				finalAssetNamesToInstall = append(finalAssetNamesToInstall, assetNamesForSingleApp)
			}
			fmt.Printf("Successfully uploaded %d file(s)\n", len(localAppsToInstall))
		}
		st := time.Now()
		params := limrun.AndroidInstanceNewParams{
			Wait: param.NewOpt(true),
			Spec: limrun.AndroidInstanceNewParamsSpec{},
		}
		if len(finalAssetNamesToInstall) > 0 {
			for _, assetNames := range finalAssetNamesToInstall {
				params.Spec.InitialAssets = append(params.Spec.InitialAssets, limrun.AndroidInstanceNewParamsSpecInitialAsset{
					Kind:       "App",
					Source:     "AssetNames",
					AssetNames: assetNames,
				})
			}
		}
		i, err := lim.AndroidInstances.New(cmd.Context(), params)
		if err != nil {
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				fmt.Println("You are logged in now")
				return nil
			}
			return fmt.Errorf("failed to create a new Android instance: %w", err)
		}
		if deleteOnExit {
			defer func() {
				if err := lim.AndroidInstances.Delete(cmd.Context(), i.Metadata.ID); err != nil {
					fmt.Printf("Failed to delete instance: %s", err)
					return
				}
				fmt.Printf("%s is deleted\n", i.Metadata.ID)
			}()
		}
		fmt.Printf("Created a new instance in %s\n", time.Since(st))
		if connect {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			t, err := tunnel.NewADB(i.Status.AdbWebSocketURL, i.Status.Token, tunnel.WithADBPath(adbPath))
			if err != nil {
				return fmt.Errorf("failed to create tunnel: %w", err)
			}
			if err := t.Start(); err != nil {
				return fmt.Errorf("failed to start tunnel: %w", err)
			}
			defer t.Close()
			if stream {
				go func() {
					if out, err := exec.CommandContext(cmd.Context(), "scrcpy", "-s", t.Addr()).CombinedOutput(); err != nil {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "failed to start scrcpy: %s %s", err.Error(), string(out))
					}
					sigChan <- syscall.SIGTERM
				}()
			}
			fmt.Println("Tunnel started. Press Ctrl+C to stop.")
			select {
			case sig := <-sigChan:
				fmt.Printf("Received signal %v, stopping tunnel...\n", sig)
			}
		} else {
			cmd.Printf("Created instance %s\n", i.Metadata.ID)
		}
		return nil
	},
}
