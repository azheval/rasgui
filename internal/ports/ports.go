package ports

import (
	"context"

	"rasgui/internal/models"
)

type Repository interface {
	SeedAdmin(username, password string) error
	Authenticate(username, password string) (models.User, error)
	UserByID(id int64) (models.User, error)
	ListUsers() ([]models.User, error)
	CreateUser(username, password string) (int64, error)
	DeleteUser(userID int64) error
	ListRoles() ([]models.Role, error)
	CreateRole(name, description string, system bool) (int64, error)
	DeleteRole(roleID int64) error
	AssignRole(userID, roleID int64) error
	AddPermission(roleID int64, operationID, clusterScope, infobaseScope string) (int64, error)
	ReplaceRolePermissions(roleID int64, operationIDs []string, clusterScopes, infobaseScopes []string) error
	ReplaceRoleScopeSet(roleID int64, oldClusterScope, oldInfobaseScope, newClusterScope, newInfobaseScope string, operationIDs []string) error
	DeleteRoleScopeSet(roleID int64, clusterScope, infobaseScope string) error
	SeedConnectionProfile(name, host string, port int, description string) error
	SeedToolchainProfile(name, version, racPath, rasPath, description string, isDefault bool) error
	ListConnectionProfiles() ([]models.ConnectionProfile, error)
	ConnectionProfileByID(id int64) (models.ConnectionProfile, error)
	CreateConnectionProfile(name, host string, port int, toolchainID int64, description string, isDefault bool) (int64, error)
	UpdateConnectionProfile(id int64, name, host string, port int, toolchainID int64, description string, isDefault bool) error
	DeleteConnectionProfile(id int64) error
	ListToolchainProfiles() ([]models.ToolchainProfile, error)
	ToolchainProfileByID(id int64) (models.ToolchainProfile, error)
	CreateToolchainProfile(name, version, racPath, rasPath, description string, isDefault bool) (int64, error)
	UpdateToolchainProfile(id int64, name, version, racPath, rasPath, description string, isDefault bool) error
	DeleteToolchainProfile(id int64) error
	ListFavoriteCommands(userID int64) ([]models.FavoriteCommand, error)
	CreateFavoriteCommand(userID int64, name, operationID string, connectionID int64, values map[string]string) (int64, error)
	DeleteFavoriteCommand(userID, favoriteID int64) error
	ListAudit(limit int) ([]models.AuditLog, error)
	WriteAudit(item models.AuditLog) error
}

type CommandExecutor interface {
	Execute(ctx context.Context, request models.CommandRequest) (models.CommandResult, error)
}
