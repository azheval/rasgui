package models

import "time"

type ParamType string

const (
	ParamString   ParamType = "string"
	ParamPassword ParamType = "password"
	ParamBool     ParamType = "bool"
	ParamInt      ParamType = "int"
	ParamSelect   ParamType = "select"
)

type ParamSpec struct {
	Name        string    `json:"name"`
	ArgName     string    `json:"arg_name,omitempty"`
	Label       string    `json:"label"`
	Type        ParamType `json:"type"`
	Required    bool      `json:"required"`
	Positional  bool      `json:"positional"`
	Description string    `json:"description"`
	Options     []string  `json:"options,omitempty"`
}

type Operation struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Utility     string      `json:"utility"`
	Mode        string      `json:"mode"`
	Subcommands []string    `json:"subcommands"`
	Description string      `json:"description"`
	RiskLevel   string      `json:"risk_level"`
	Params      []ParamSpec `json:"params"`
}

type User struct {
	ID        int64
	Username  string
	Password  string
	Enabled   bool
	Roles     []Role
	CreatedAt string
}

type Role struct {
	ID          int64
	Name        string
	Description string
	System      bool
	Permissions []Permission
	CreatedAt   string
}

type ConnectionProfile struct {
	ID            int64
	Name          string
	Host          string
	Port          int
	ToolchainID   int64
	ToolchainName string
	Description   string
	IsDefault     bool
	CreatedAt     string
}

type ToolchainProfile struct {
	ID          int64
	Name        string
	Version     string
	RACPath     string
	RASPath     string
	IsDefault   bool
	Description string
	CreatedAt   string
}

type FavoriteCommand struct {
	ID           int64
	UserID       int64
	Name         string
	OperationID  string
	ConnectionID int64
	Values       map[string]string
	CreatedAt    string
}

type Permission struct {
	ID            int64
	RoleID        int64
	OperationID   string
	ClusterScope  string
	InfobaseScope string
	CreatedAt     string
}

type CommandRequest struct {
	OperationID string
	Values      map[string]string
	Actor       string
	RACPath     string
	RASPath     string
}

type CommandResult struct {
	Command    []string
	Stdout     string
	Stderr     string
	ExitCode   int
	StartedAt  time.Time
	FinishedAt time.Time
}

type AuditLog struct {
	ID            int64
	Username      string
	OperationID   string
	ClusterScope  string
	InfobaseScope string
	CommandLine   string
	Allowed       bool
	ExitCode      int
	Stdout        string
	Stderr        string
	CreatedAt     string
}
