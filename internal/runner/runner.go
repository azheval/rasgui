package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"rasgui/internal/catalog"
	"rasgui/internal/config"
	"rasgui/internal/models"
)

type Runner struct {
	cfg config.Config
}

func New(cfg config.Config) Runner {
	return Runner{cfg: cfg}
}

func (r Runner) Execute(ctx context.Context, request models.CommandRequest) (models.CommandResult, error) {
	operation, ok := catalog.Find(request.OperationID)
	if !ok {
		return models.CommandResult{}, fmt.Errorf("unknown operation %s", request.OperationID)
	}

	binary := r.cfg.RACPath
	if strings.TrimSpace(request.RACPath) != "" {
		binary = request.RACPath
	}
	if operation.Utility == "ras" {
		binary = r.cfg.RASPath
		if strings.TrimSpace(request.RASPath) != "" {
			binary = request.RASPath
		}
	}

	args, err := buildArgs(operation, request.Values)
	if err != nil {
		return models.CommandResult{}, err
	}

	started := time.Now()
	cmd := exec.CommandContext(ctx, binary, args...)
	out, err := cmd.CombinedOutput()
	finished := time.Now()

	result := models.CommandResult{
		Command:    append([]string{binary}, args...),
		Stdout:     decodeOutput(out),
		StartedAt:  started,
		FinishedAt: finished,
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		result.Stderr = err.Error()
	}

	return result, nil
}

func decodeOutput(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	if utf8.Valid(raw) {
		return string(raw)
	}
	if runtime.GOOS == "windows" {
		if text, err := charmap.CodePage866.NewDecoder().Bytes(raw); err == nil && utf8.Valid(text) {
			return string(text)
		}
		if text, err := charmap.Windows1251.NewDecoder().Bytes(raw); err == nil && utf8.Valid(text) {
			return string(text)
		}
	}
	return string(raw)
}

func buildArgs(operation models.Operation, values map[string]string) ([]string, error) {
	args := []string{}
	if operation.Mode != "" {
		args = append(args, operation.Mode)
	}
	if len(operation.Subcommands) > 0 {
		args = append(args, operation.Subcommands...)
	}

	if operation.Utility == "rac" {
		for _, param := range operation.Params {
			switch param.Name {
			case "host", "admin_port", "extra_args":
				continue
			}
			if err := appendNamedArg(&args, param, values[param.Name], values); err != nil {
				return nil, err
			}
		}
		host := values["host"]
		if host == "" {
			host = "localhost"
		}
		if port := values["admin_port"]; port != "" {
			host = host + ":" + port
		}
		args = append(args, host)
	} else {
		for _, param := range operation.Params {
			if param.Name == "extra_args" {
				continue
			}
			if param.Name == "target_host" {
				continue
			}
			if err := appendNamedArg(&args, param, values[param.Name], values); err != nil {
				return nil, err
			}
		}
		if target := values["target_host"]; target != "" {
			args = append(args, target)
		}
	}

	if extra := strings.TrimSpace(values["extra_args"]); extra != "" {
		extraArgs, err := splitCommandLine(extra)
		if err != nil {
			return nil, err
		}
		args = append(args, extraArgs...)
	}
	if wizardJSON := strings.TrimSpace(values["wizard_args_json"]); wizardJSON != "" {
		var wizardArgs []string
		if err := json.Unmarshal([]byte(wizardJSON), &wizardArgs); err != nil {
			return nil, fmt.Errorf("invalid wizard args: %w", err)
		}
		args = append(args, wizardArgs...)
	}

	return args, nil
}

func appendNamedArg(args *[]string, spec models.ParamSpec, value string, values map[string]string) error {
	if spec.Type == models.ParamBool {
		if values[spec.Name] == "on" || strings.EqualFold(values[spec.Name], "true") || values[spec.Name] == "1" {
			*args = append(*args, "--"+spec.Name)
		}
		return nil
	}
	if spec.Required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("missing required parameter %s", spec.Name)
	}
	if strings.TrimSpace(value) == "" {
		return nil
	}
	if spec.Name == "mode" {
		if value == "partial" {
			*args = append(*args, "--partial")
		} else {
			*args = append(*args, "--full")
		}
		return nil
	}
	if spec.Positional {
		*args = append(*args, value)
		return nil
	}
	argName := spec.Name
	if strings.TrimSpace(spec.ArgName) != "" {
		argName = spec.ArgName
	}
	*args = append(*args, fmt.Sprintf("--%s=%s", argName, value))
	return nil
}

func splitCommandLine(value string) ([]string, error) {
	var result []string
	var current strings.Builder
	var quote rune
	escaped := false
	for _, r := range value {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			current.WriteRune(r)
			continue
		}
		switch {
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quoted argument")
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result, nil
}
