package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	WorkDir              string `json:"-"`
	HTTPPort             int    `json:"http_port"`
	DataDir              string `json:"data_dir"`
	LogDir               string `json:"log_dir"`
	LogLevel             string `json:"log_level"`
	DefaultRASHost       string `json:"default_ras_host"`
	DefaultRASPort       int    `json:"default_ras_port"`
	RACPath              string `json:"rac_path"`
	RASPath              string `json:"ras_path"`
	DefaultAdminUser     string `json:"default_admin_user"`
	DefaultAdminPassword string `json:"default_admin_password"`
}

func Load(workdir string) (Config, error) {
	cfg := defaults(workdir)
	cfg.WorkDir = workdir

	configPath := filepath.Join(workdir, "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	if value := os.Getenv("RASGUI_HTTP_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, err
		}
		cfg.HTTPPort = port
	}
	if value := os.Getenv("RASGUI_DATA_DIR"); value != "" {
		cfg.DataDir = value
	}
	if value := os.Getenv("RASGUI_RAC_PATH"); value != "" {
		cfg.RACPath = value
	}
	if value := os.Getenv("RASGUI_RAS_PATH"); value != "" {
		cfg.RASPath = value
	}
	if value := os.Getenv("RASGUI_LOG_DIR"); value != "" {
		cfg.LogDir = value
	}
	if value := os.Getenv("RASGUI_LOG_LEVEL"); value != "" {
		cfg.LogLevel = value
	}
	if value := os.Getenv("RASGUI_DEFAULT_RAS_HOST"); value != "" {
		cfg.DefaultRASHost = value
	}
	if value := os.Getenv("RASGUI_DEFAULT_RAS_PORT"); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, err
		}
		cfg.DefaultRASPort = port
	}

	if !filepath.IsAbs(cfg.DataDir) {
		cfg.DataDir = filepath.Join(workdir, cfg.DataDir)
	}
	if !filepath.IsAbs(cfg.LogDir) {
		cfg.LogDir = filepath.Join(workdir, cfg.LogDir)
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return Config{}, err
	}
	if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func defaults(workdir string) Config {
	cfg := Config{
		HTTPPort:             8099,
		DataDir:              filepath.Join(workdir, "data"),
		LogDir:               filepath.Join(workdir, "logs"),
		LogLevel:             "info",
		DefaultRASHost:       "localhost",
		DefaultRASPort:       1545,
		RACPath:              "rac",
		RASPath:              "ras",
		DefaultAdminUser:     "admin",
		DefaultAdminPassword: "admin123",
	}

	if runtime.GOOS == "windows" {
		cfg.RACPath = `C:\Program Files\1cv8\8.3.27.2074\bin\rac.exe`
		cfg.RASPath = `C:\Program Files\1cv8\8.3.27.2074\bin\ras.exe`
	}

	return cfg
}

type ToolchainSeed struct {
	Name        string
	Version     string
	RACPath     string
	RASPath     string
	Description string
	IsDefault   bool
}

func DiscoverLocalToolchains(cfg Config) []ToolchainSeed {
	items := []ToolchainSeed{{
		Name:        "default",
		Version:     detectVersionFromPath(cfg.RACPath),
		RACPath:     cfg.RACPath,
		RASPath:     cfg.RASPath,
		Description: "Default toolchain from config",
		IsDefault:   true,
	}}
	if runtime.GOOS != "windows" {
		return items
	}
	candidates := map[string]ToolchainSeed{
		cfg.RACPath + "\x00" + cfg.RASPath: items[0],
	}
	patterns := []string{
		`C:\Program Files\1cv8\*\bin\rac.exe`,
		`C:\Program Files (x86)\1cv8\*\bin\rac.exe`,
	}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, racPath := range matches {
			base := filepath.Dir(racPath)
			rasPath := filepath.Join(base, "ras.exe")
			if _, err := os.Stat(rasPath); err != nil {
				continue
			}
			version := detectVersionFromPath(racPath)
			item := ToolchainSeed{
				Name:        "1C " + version,
				Version:     version,
				RACPath:     racPath,
				RASPath:     rasPath,
				Description: "Auto-discovered local 1C toolchain",
			}
			candidates[racPath+"\x00"+rasPath] = item
		}
	}
	keys := make([]string, 0, len(candidates))
	for key := range candidates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]ToolchainSeed, 0, len(keys))
	for _, key := range keys {
		result = append(result, candidates[key])
	}
	return result
}

func detectVersionFromPath(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if strings.Count(part, ".") >= 2 {
			return part
		}
	}
	return "custom"
}
