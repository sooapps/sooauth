package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func loadEnvFiles() {
	if explicit := os.Getenv("ENV_FILE"); explicit != "" {
		if err := godotenv.Load(explicit); err != nil {
			slog.Debug("env file not loaded", "path", explicit, "err", err)
		}
		return
	}

	candidates := envFileCandidates()
	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := godotenv.Load(path); err != nil {
			slog.Warn("env file found but not loaded", "path", path, "err", err)
			return
		}
		slog.Info("loaded env file", "path", path)
		return
	}
}

func envFileCandidates() []string {
	var out []string
	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for range 4 {
			out = append(out, filepath.Join(dir, ".env"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	out = append(out, ".env", "../.env")
	return out
}
