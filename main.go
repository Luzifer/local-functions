// Local-Functions Utility Server
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Luzifer/go_helpers/env"
	httphelper "github.com/Luzifer/go_helpers/http"
	"github.com/Luzifer/rconfig/v2"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	cfg = struct {
		CommandTimeout time.Duration `flag:"command-timeout" default:"30s" description:"How long a script may maximum take to execute"`
		Listen         string        `flag:"listen" default:"127.0.0.1:3000" description:"Port/IP to listen on"`
		LogLevel       string        `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		ScriptDir      string        `flag:"script-dir" default:"./scripts" description:"Directory to execute the script / binary from"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func initApp() (err error) {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		return fmt.Errorf("parsing CLI options: %w", err)
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parsing log-level: %w", err)
	}
	logrus.SetLevel(l)

	return nil
}

func main() {
	var err error
	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		fmt.Printf("local-functions %s\n", version) //nolint:forbidigo // fine to print version to stdout
		os.Exit(0)
	}

	r := mux.NewRouter()
	r.HandleFunc("/{script}", handleScriptCall)

	var hdl http.Handler = r
	hdl = httphelper.NewHTTPLogHandler(hdl)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           hdl,
		ReadHeaderTimeout: time.Second,
	}

	logrus.WithField("addr", cfg.Listen).Info("local-functions started")
	if err = srv.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("listening for HTTP traffic")
	}
}

func handleScriptCall(w http.ResponseWriter, r *http.Request) {
	var (
		vars   = mux.Vars(r)
		script = filepath.Join(cfg.ScriptDir, vars["script"])
	)

	if err := validateScriptContained(script); err != nil {
		logrus.WithField("script", script).Error("path traversal attempt")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(script); vars["script"] == "" || os.IsNotExist(err) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), cfg.CommandTimeout)
	defer cancel()

	var (
		stdout  = new(bytes.Buffer)
		cmd     = exec.CommandContext(ctx, script) //#nosec:G204 // Intended to run scripts provided in scripts dir
		envVars = env.ListToMap(os.Environ())
	)

	envVars["ACCEPT"] = r.Header.Get("Accept")
	envVars["CONTENT_TYPE"] = r.Header.Get("Content-Type")
	envVars["METHOD"] = r.Method

	cmd.Env = env.MapToList(envVars)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = r.Body

	if err := cmd.Run(); err != nil {
		http.Error(w, "Script execution failed, see log", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "*")

	_, _ = io.Copy(w, stdout)
}
