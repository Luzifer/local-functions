package main

import (
	"fmt"
	"path/filepath"
)

func validateScriptContained(script string) (err error) {
	scriptDirAbs, err := filepath.Abs(cfg.ScriptDir)
	if err != nil {
		return fmt.Errorf("resolving script dir: %w", err)
	}

	scriptAbs, err := filepath.Abs(script)
	if err != nil {
		return fmt.Errorf("resolving script path: %w", err)
	}

	relScript, err := filepath.Rel(scriptDirAbs, scriptAbs)
	if err != nil {
		return fmt.Errorf("determining relative script path: %w", err)
	}

	if !filepath.IsLocal(relScript) {
		return fmt.Errorf("path traversal attempt")
	}

	return nil
}
