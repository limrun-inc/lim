package run

import (
	"archive/zip"
	"fmt"
	"github.com/limrun-inc/go-sdk/tunnel"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	limrun "github.com/limrun-inc/go-sdk"
	"github.com/limrun-inc/go-sdk/packages/param"

	"github.com/limrun-inc/lim/config"
	"github.com/limrun-inc/lim/errors"
)

var (
	portPair  string
	streamIos bool
)

func init() {
	IosCmd.PersistentFlags().StringVar(&portPair, "connect", "", "Connects given ports of the remote simulator. The format is <local port>:<simulator port>, e.g. 8100:8100.")
	IosCmd.PersistentFlags().BoolVar(&streamIos, "stream", true, "Opens the streaming page after creation. Default is true.")
	IosCmd.PersistentFlags().StringArrayVar(&assetNamesToInstall, "install-asset", []string{}, "List of asset names to install. It will return error if they are not already uploaded.")
	IosCmd.PersistentFlags().StringArrayVar(&localAppsToInstall, "install", []string{}, "List of local app folders to be installed. They will be zipped and uploaded to the asset storage if they are not uploaded yet. Subsequent calls will just use the uploaded ZIP file.")
}

var IosCmd = &cobra.Command{
	Use:   "ios",
	Short: "Creates a new iOS instance, sets the instance id for `lim simctl` commands and opens the streaming page.",
	RunE: func(cmd *cobra.Command, args []string) error {
		lim := cmd.Context().Value("lim").(limrun.Client)
		finalAssetNamesToInstall := assetNamesToInstall
		if len(localAppsToInstall) > 0 {
			for _, appPath := range localAppsToInstall {
				if appPath == "" {
					continue
				}
				appSt, err := os.Stat(appPath)
				if err != nil {
					return fmt.Errorf("failed to stat %s: %w", appPath, err)
				}
				size := appSt.Size()
				if appSt.IsDir() {
					fmt.Printf("Archiving %s\n", filepath.Base(appPath))
					tempZipFilePath := filepath.Join(os.TempDir(), filepath.Base(appPath)+".zip")
					defer func() {
						fmt.Printf("Deleting the temporary %s\n", tempZipFilePath)
						if err := os.RemoveAll(tempZipFilePath); err != nil {
							cmd.Printf("failed to remove temporary zip file at %s: %s", tempZipFilePath, err.Error())
						}
					}()
					size, err = zipFolderDeterministically(appPath, tempZipFilePath)
					if err != nil {
						return fmt.Errorf("failed to ZIP folder %s: %w", appPath, err)
					}
					appPath = tempZipFilePath
				}
				bar := progressbar.DefaultBytes(
					size,
					"",
				)
				fmt.Printf("Uploading %s\n", appPath)
				ass, err := lim.Assets.GetOrUpload(cmd.Context(), limrun.AssetGetOrUploadParams{
					Name:           param.NewOpt(appPath),
					Path:           appPath,
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
					return fmt.Errorf("failed to upload app at %s: %w", appPath, err)
				}
				if err := bar.Close(); err != nil {
					return err
				}
				finalAssetNamesToInstall = append(finalAssetNamesToInstall, ass.Name)
			}
			fmt.Printf("Successfully uploaded %d file(s)\n", len(localAppsToInstall))
		}
		params := limrun.IosInstanceNewParams{
			Wait: param.NewOpt(true),
			Spec: limrun.IosInstanceNewParamsSpec{},
		}
		if len(finalAssetNamesToInstall) > 0 {
			for _, assetName := range finalAssetNamesToInstall {
				params.Spec.InitialAssets = append(params.Spec.InitialAssets, limrun.IosInstanceNewParamsSpecInitialAsset{
					Kind:       "App",
					Source:     "AssetName",
					AssetName:  param.NewOpt(assetName),
					LaunchMode: "ForegroundIfRunning",
				})
			}
		}
		st := time.Now()
		i, err := lim.IosInstances.New(cmd.Context(), params)
		if err != nil {
			if errors.IsUnauthenticated(err) {
				if err := config.Login(cmd.Context()); err != nil {
					return err
				}
				cmd.Println("You are logged in now")
				return nil
			}
			return fmt.Errorf("failed to create a new iOS instance: %w", err)
		}
		viper.Set(config.ConfigKeySimctlInstanceID, i.Metadata.ID)
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		defaultConfigDir := filepath.Join(home, ".lim")
		if err := viper.WriteConfigAs(filepath.Join(defaultConfigDir, "config.yaml")); err != nil {
			return fmt.Errorf("failed to write initial config: %v", err)
		}
		u := fmt.Sprintf("https://console.limrun.com/stream/%s", i.Metadata.ID)
		cmd.Printf("Created a new instance in %s\n", time.Since(st))
		cmd.Printf("You can access it with the following URL: %s\n", u)
		if portPair == "" {
			return nil
		}
		ports := strings.Split(portPair, ":")
		if len(ports) != 2 {
			return fmt.Errorf("invalid port pair format: %s", args[1])
		}
		localPort := ports[0]
		remotePort, err := strconv.Atoi(ports[1])
		if err != nil {
			return fmt.Errorf("invalid remote port: %s %w", args[1], err)
		}
		pfUrl, err := url.Parse(i.Status.PortForwardWebSocketURL)
		if err != nil {
			return fmt.Errorf("failed to parse portForwardWebSocket URL %s: %w", i.Status.PortForwardWebSocketURL, err)
		}
		var opts []tunnel.MultiplexedOption
		if localPort != "" {
			localPortInt, err := strconv.Atoi(localPort)
			if err != nil {
				return fmt.Errorf("invalid local port: %s %w", localPort, err)
			}
			opts = append(opts, tunnel.MultiplexedWithLocalPort(localPortInt))
		}
		t, err := tunnel.NewMultiplexed(pfUrl, remotePort, i.Status.Token, opts...)
		if err != nil {
			return fmt.Errorf("failed to create tunnel: %w", err)
		}
		if err := t.Start(); err != nil {
			return fmt.Errorf("failed to start tunnel: %w", err)
		}
		defer t.Close()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		fmt.Printf("Listening for connections on %s\n", t.Addr())
		fmt.Println("Tunnel started. Press Ctrl+C to stop.")
		select {
		case sig := <-sigChan:
			fmt.Printf("Received signal %v, stopping tunnel...\n", sig)
		}
		return nil
	},
}

func openBrowser(url string) error {
	switch osName := runtime.GOOS; osName {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform to open browser: %s", osName)
	}
}

// zipFolderDeterministically creates a deterministic ZIP file from the given folder path and returns the path.
// The determinism covers ensuring the file order and resetting last-modified timestamps but not the original file
// permissions. If the file permissions change, it would result in a different ZIP file.
func zipFolderDeterministically(folderPath, outZipPath string) (int64, error) {
	zipFile, err := os.Create(outZipPath)
	if err != nil {
		return -1, fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer zipFile.Close()

	// Create a new ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Collect all files first to sort them for deterministic order
	var files []string
	err = filepath.Walk(folderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the root folder itself
		if filePath == folderPath {
			return nil
		}
		files = append(files, filePath)
		return nil
	})
	if err != nil {
		return -1, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort files for deterministic order
	sort.Strings(files)

	// Fixed timestamp for deterministic ZIP (Unix epoch)
	fixedTime := time.Unix(0, 0)

	// Add files to ZIP in sorted order
	for _, filePath := range files {
		info, err := os.Stat(filePath)
		if err != nil {
			return -1, fmt.Errorf("failed to stat file %s: %w", filePath, err)
		}

		// Get the relative path within the folder
		relPath, err := filepath.Rel(folderPath, filePath)
		if err != nil {
			return -1, fmt.Errorf("failed to get relative path: %w", err)
		}

		// Create a header for the file in the ZIP with deterministic values
		header := &zip.FileHeader{
			Name:     relPath,
			Modified: fixedTime,
			Method:   zip.Deflate,
		}

		// If it's a directory, just create the directory entry
		if info.IsDir() {
			header.Name += "/"
			header.Method = zip.Store
			_, err := zipWriter.CreateHeader(header)
			if err != nil {
				return -1, fmt.Errorf("failed to create directory header: %w", err)
			}
			continue
		}

		// Preserve original file permissions
		header.SetMode(info.Mode())

		// Create the file in the ZIP
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return -1, fmt.Errorf("failed to create file header: %w", err)
		}

		// Open the file and copy its contents to the ZIP
		file, err := os.Open(filePath)
		if err != nil {
			return -1, fmt.Errorf("failed to open file %s: %w", filePath, err)
		}

		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return -1, fmt.Errorf("failed to copy file %s: %w", filePath, err)
		}
	}
	st, err := zipFile.Stat()
	if err != nil {
		return -1, fmt.Errorf("failed to stat the zip file at %s: %w", zipFile.Name(), err)
	}
	return st.Size(), nil
}
