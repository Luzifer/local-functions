package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/env"
	httpHelper "github.com/Luzifer/go_helpers/v2/http"
	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		Listen         string `flag:"listen" default:"127.0.0.1:3000" description:"Port/IP to listen on"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		ScriptDir      string `flag:"script-dir" default:"./scripts" description:"Directory to execute the script / binary from"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("local-functions %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{script}", handleScriptCall)

	var h http.Handler = r
	h = httpHelper.NewHTTPLogHandler(h)
	http.ListenAndServe(cfg.Listen, h)
}

func handleScriptCall(w http.ResponseWriter, r *http.Request) {
	var (
		vars   = mux.Vars(r)
		script = path.Join(cfg.ScriptDir, vars["script"])
	)

	if _, err := os.Stat(script); vars["script"] == "" || os.IsNotExist(err) {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var (
		stdout = new(bytes.Buffer)
		cmd    = exec.Command(script)
	)

	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = r.Body

	cmd.Env = env.MapToList(map[string]string{
		"ACCEPT":       r.Header.Get("Accept"),
		"CONTENT_TYPE": r.Header.Get("Content-Type"),
		"METHOD":       r.Method,
	})

	if err := cmd.Run(); err != nil {
		http.Error(w, "Script execution failed, see log", http.StatusInternalServerError)
		return
	}

	io.Copy(w, stdout)
}
