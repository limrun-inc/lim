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

package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/limrun-inc/lim/version"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
)

const (
	ConfigKeyAPIKey          = "api-key"
	ConfigKeyAPIEndpoint     = "api-endpoint"
	ConfigKeyConsoleEndpoint = "console-endpoint"
)

func Login(ctx context.Context) error {
	loggedIn := make(chan bool)
	mux := http.NewServeMux()
	mux.HandleFunc("/authn/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		apiKey := r.URL.Query().Get(ConfigKeyAPIKey)
		if apiKey == "" {
			http.Error(w, "missing apiKey", http.StatusBadRequest)
			loggedIn <- false
			return
		}
		viper.Set(ConfigKeyAPIKey, apiKey)
		if err := viper.WriteConfig(); err != nil {
			http.Error(w, fmt.Sprintf("failed to write config: %v", err), http.StatusInternalServerError)
			loggedIn <- false
		}
		w.WriteHeader(http.StatusOK)
		loggedIn <- true
	})
	srv := &http.Server{
		Addr:    ":32412",
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("HTTP server error:", err)
		}
	}()
	consoleUrl, err := url.Parse(viper.GetString(ConfigKeyConsoleEndpoint))
	if err != nil {
		return fmt.Errorf("failed to parse %s as console endpoint: %w", viper.GetString(ConfigKeyConsoleEndpoint), err)
	}
	u := consoleUrl.JoinPath("authn", "login")
	vals := u.Query()
	vals.Set("user-agent", "lim/"+version.Version)
	u.RawQuery = vals.Encode()
	if err := openBrowser(u.String()); err != nil {
		return err
	}
	<-loggedIn
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %v", err)
	}
	return nil
}

func openBrowser(url string) error {
	switch os := runtime.GOOS; os {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform to open browser: %s", os)
	}
}
