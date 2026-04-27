package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"rasgui/internal/catalog"
	"rasgui/internal/models"
	"rasgui/internal/ports"
	"rasgui/internal/rbac"
)

type Service struct {
	repo   ports.Repository
	exec   ports.CommandExecutor
	logger *slog.Logger
}

func New(repo ports.Repository, exec ports.CommandExecutor, logger *slog.Logger) *Service {
	return &Service{repo: repo, exec: exec, logger: logger}
}

func (s *Service) SeedAdmin(username, password string) error {
	return s.repo.SeedAdmin(username, password)
}

func (s *Service) SeedConnectionProfile(name, host string, port int, description string) error {
	return s.repo.SeedConnectionProfile(name, host, port, description)
}

func (s *Service) Authenticate(username, password string) (models.User, error) {
	user, err := s.repo.Authenticate(username, password)
	if err != nil {
		s.logger.Warn("authentication failed", "username", username, "error", err)
		return models.User{}, err
	}
	s.logger.Info("authentication succeeded", "username", username)
	return user, nil
}

func (s *Service) UserByID(id int64) (models.User, error) { return s.repo.UserByID(id) }
func (s *Service) ListUsers() ([]models.User, error)      { return s.repo.ListUsers() }
func (s *Service) ListRoles() ([]models.Role, error)      { return s.repo.ListRoles() }
func (s *Service) ListAudit(limit int) ([]models.AuditLog, error) {
	return s.repo.ListAudit(limit)
}
func (s *Service) ListConnectionProfiles() ([]models.ConnectionProfile, error) {
	return s.repo.ListConnectionProfiles()
}
func (s *Service) ListToolchainProfiles() ([]models.ToolchainProfile, error) {
	return s.repo.ListToolchainProfiles()
}
func (s *Service) ListFavoriteCommands(userID int64) ([]models.FavoriteCommand, error) {
	return s.repo.ListFavoriteCommands(userID)
}
func (s *Service) Operations() []models.Operation { return catalog.Catalog() }

func (s *Service) OperationsForUser(user models.User) []models.Operation {
	all := s.Operations()
	if rbac.Allowed(user, "*", "*", "*") {
		return all
	}
	allowed := make([]models.Operation, 0, len(all))
	for _, item := range all {
		if rbac.Allowed(user, item.ID, "*", "*") {
			allowed = append(allowed, item)
		}
	}
	return allowed
}

func (s *Service) FavoriteCommandsForUser(user models.User) ([]models.FavoriteCommand, error) {
	items, err := s.repo.ListFavoriteCommands(user.ID)
	if err != nil {
		return nil, err
	}
	if rbac.Allowed(user, "*", "*", "*") {
		return items, nil
	}
	allowed := make([]models.FavoriteCommand, 0, len(items))
	for _, item := range items {
		if rbac.Allowed(user, item.OperationID, "*", "*") {
			allowed = append(allowed, item)
		}
	}
	return allowed, nil
}

func (s *Service) CanManageUsers(user models.User) bool {
	return hasAdministrativeAccess(user)
}

func (s *Service) CanManageRoles(user models.User) bool {
	return hasAdministrativeAccess(user)
}

func (s *Service) CanManageConnections(user models.User) bool {
	return hasAdministrativeAccess(user)
}

func (s *Service) CanViewAudit(user models.User) bool {
	return hasAdministrativeAccess(user)
}

func (s *Service) CreateUser(username, password string, roleID int64) error {
	userID, err := s.repo.CreateUser(username, password)
	if err != nil {
		s.logger.Error("create user failed", "username", username, "error", err)
		return err
	}
	if roleID > 0 {
		if err := s.repo.AssignRole(userID, roleID); err != nil {
			return err
		}
	}
	s.logger.Info("user created", "username", username, "role_id", roleID)
	return nil
}

func (s *Service) DeleteUser(currentUserID, targetUserID int64) error {
	if currentUserID == targetUserID {
		return errors.New("current user cannot be deleted")
	}
	if err := s.repo.DeleteUser(targetUserID); err != nil {
		s.logger.Error("delete user failed", "target_user_id", targetUserID, "error", err)
		return err
	}
	s.logger.Info("user deleted", "target_user_id", targetUserID)
	return nil
}

func (s *Service) CreateRole(name, description string) error {
	_, err := s.repo.CreateRole(name, description, false)
	if err != nil {
		s.logger.Error("create role failed", "role", name, "error", err)
		return err
	}
	s.logger.Info("role created", "role", name)
	return nil
}

func (s *Service) DeleteRole(roleID int64) error {
	if err := s.repo.DeleteRole(roleID); err != nil {
		s.logger.Error("delete role failed", "role_id", roleID, "error", err)
		return err
	}
	s.logger.Info("role deleted", "role_id", roleID)
	return nil
}

func (s *Service) AddPermission(roleID int64, operationID, clusterScope, infobaseScope string) error {
	_, err := s.repo.AddPermission(roleID, operationID, clusterScope, infobaseScope)
	if err != nil {
		s.logger.Error("add permission failed", "role_id", roleID, "operation", operationID, "error", err)
		return err
	}
	s.logger.Info("permission added", "role_id", roleID, "operation", operationID, "cluster_scope", clusterScope, "infobase_scope", infobaseScope)
	return nil
}

func (s *Service) ReplaceRolePermissions(roleID int64, operationIDs []string, clusterScopes, infobaseScopes []string) error {
	if err := s.repo.ReplaceRolePermissions(roleID, operationIDs, clusterScopes, infobaseScopes); err != nil {
		s.logger.Error("replace role permissions failed", "role_id", roleID, "operation_count", len(operationIDs), "cluster_scope_count", len(clusterScopes), "infobase_scope_count", len(infobaseScopes), "error", err)
		return err
	}
	s.logger.Info("role permissions replaced", "role_id", roleID, "operation_count", len(operationIDs), "cluster_scope_count", len(clusterScopes), "infobase_scope_count", len(infobaseScopes))
	return nil
}

func (s *Service) ReplaceRoleScopeSet(roleID int64, oldClusterScope, oldInfobaseScope, newClusterScope, newInfobaseScope string, operationIDs []string) error {
	if err := s.repo.ReplaceRoleScopeSet(roleID, oldClusterScope, oldInfobaseScope, newClusterScope, newInfobaseScope, operationIDs); err != nil {
		s.logger.Error("replace role scope set failed", "role_id", roleID, "old_cluster_scope", oldClusterScope, "old_infobase_scope", oldInfobaseScope, "new_cluster_scope", newClusterScope, "new_infobase_scope", newInfobaseScope, "operation_count", len(operationIDs), "error", err)
		return err
	}
	s.logger.Info("role scope set replaced", "role_id", roleID, "old_cluster_scope", oldClusterScope, "old_infobase_scope", oldInfobaseScope, "new_cluster_scope", newClusterScope, "new_infobase_scope", newInfobaseScope, "operation_count", len(operationIDs))
	return nil
}

func (s *Service) DeleteRoleScopeSet(roleID int64, clusterScope, infobaseScope string) error {
	if err := s.repo.DeleteRoleScopeSet(roleID, clusterScope, infobaseScope); err != nil {
		s.logger.Error("delete role scope set failed", "role_id", roleID, "cluster_scope", clusterScope, "infobase_scope", infobaseScope, "error", err)
		return err
	}
	s.logger.Info("role scope set deleted", "role_id", roleID, "cluster_scope", clusterScope, "infobase_scope", infobaseScope)
	return nil
}

func (s *Service) CreateConnectionProfile(name, host string, port int, description string, isDefault bool) error {
	return s.CreateConnectionProfileWithToolchain(name, host, port, 0, description, isDefault)
}

func (s *Service) CreateConnectionProfileWithToolchain(name, host string, port int, toolchainID int64, description string, isDefault bool) error {
	_, err := s.repo.CreateConnectionProfile(name, host, port, toolchainID, description, isDefault)
	if err != nil {
		s.logger.Error("create connection profile failed", "name", name, "host", host, "port", port, "toolchain_id", toolchainID, "error", err)
		return err
	}
	s.logger.Info("connection profile created", "name", name, "host", host, "port", port, "toolchain_id", toolchainID, "is_default", isDefault)
	return nil
}

func (s *Service) UpdateConnectionProfile(id int64, name, host string, port int, toolchainID int64, description string, isDefault bool) error {
	if err := s.repo.UpdateConnectionProfile(id, name, host, port, toolchainID, description, isDefault); err != nil {
		s.logger.Error("update connection profile failed", "id", id, "name", name, "error", err)
		return err
	}
	s.logger.Info("connection profile updated", "id", id, "name", name, "toolchain_id", toolchainID, "is_default", isDefault)
	return nil
}

func (s *Service) DeleteConnectionProfile(id int64) error {
	if err := s.repo.DeleteConnectionProfile(id); err != nil {
		s.logger.Error("delete connection profile failed", "id", id, "error", err)
		return err
	}
	s.logger.Info("connection profile deleted", "id", id)
	return nil
}

func (s *Service) CheckConnectionProfile(id int64) (string, error) {
	profile, err := s.repo.ConnectionProfileByID(id)
	if err != nil {
		return "", err
	}
	parts := []string{}
	if profile.ToolchainID > 0 {
		if toolchainMessage, err := s.CheckToolchainProfile(profile.ToolchainID); err == nil {
			parts = append(parts, toolchainMessage)
		}
	}
	address := net.JoinHostPort(strings.TrimSpace(profile.Host), strconv.Itoa(profile.Port))
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		parts = append(parts, "RAS endpoint is not reachable")
	} else {
		_ = conn.Close()
		parts = append(parts, "RAS endpoint is reachable")
	}
	return strings.Join(parts, "; "), nil
}

func (s *Service) SeedToolchainProfile(name, version, racPath, rasPath, description string, isDefault bool) error {
	return s.repo.SeedToolchainProfile(name, version, racPath, rasPath, description, isDefault)
}

func (s *Service) CreateToolchainProfile(name, version, racPath, rasPath, description string, isDefault bool) error {
	_, err := s.repo.CreateToolchainProfile(name, version, racPath, rasPath, description, isDefault)
	if err != nil {
		s.logger.Error("create toolchain profile failed", "name", name, "version", version, "rac_path", racPath, "ras_path", rasPath, "error", err)
		return err
	}
	s.logger.Info("toolchain profile created", "name", name, "version", version, "is_default", isDefault)
	return nil
}

func (s *Service) UpdateToolchainProfile(id int64, name, version, racPath, rasPath, description string, isDefault bool) error {
	if err := s.repo.UpdateToolchainProfile(id, name, version, racPath, rasPath, description, isDefault); err != nil {
		s.logger.Error("update toolchain profile failed", "id", id, "name", name, "error", err)
		return err
	}
	s.logger.Info("toolchain profile updated", "id", id, "name", name, "is_default", isDefault)
	return nil
}

func (s *Service) DeleteToolchainProfile(id int64) error {
	if err := s.repo.DeleteToolchainProfile(id); err != nil {
		s.logger.Error("delete toolchain profile failed", "id", id, "error", err)
		return err
	}
	s.logger.Info("toolchain profile deleted", "id", id)
	return nil
}

func (s *Service) CheckToolchainProfile(id int64) (string, error) {
	item, err := s.repo.ToolchainProfileByID(id)
	if err != nil {
		return "", err
	}
	problems := []string{}
	if _, err := os.Stat(item.RACPath); err != nil {
		problems = append(problems, "RAC path is not accessible")
	}
	if _, err := os.Stat(item.RASPath); err != nil {
		problems = append(problems, "RAS path is not accessible")
	}
	if len(problems) > 0 {
		return strings.Join(problems, "; "), nil
	}
	return "Toolchain check passed", nil
}

func (s *Service) CreateFavoriteCommand(userID int64, name, operationID string, connectionID int64, values map[string]string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("favorite name is required")
	}
	if strings.TrimSpace(operationID) == "" {
		return errors.New("operation is required")
	}
	_, err := s.repo.CreateFavoriteCommand(userID, name, operationID, connectionID, values)
	if err != nil {
		s.logger.Error("create favorite failed", "user_id", userID, "name", name, "operation", operationID, "error", err)
		return err
	}
	s.logger.Info("favorite created", "user_id", userID, "name", name, "operation", operationID)
	return nil
}

func (s *Service) DeleteFavoriteCommand(userID, favoriteID int64) error {
	if err := s.repo.DeleteFavoriteCommand(userID, favoriteID); err != nil {
		s.logger.Error("delete favorite failed", "user_id", userID, "favorite_id", favoriteID, "error", err)
		return err
	}
	s.logger.Info("favorite deleted", "user_id", userID, "favorite_id", favoriteID)
	return nil
}

func (s *Service) ExecuteOperation(ctx context.Context, user models.User, values map[string]string) (*models.CommandResult, string, error) {
	opID := values["operation_id"]
	clusterScope := firstNonEmpty(values["cluster"], "*")
	infobaseScope := firstNonEmpty(values["infobase"], values["name"])
	allowed := rbac.Allowed(user, opID, clusterScope, infobaseScope)

	audit := models.AuditLog{
		Username:      user.Username,
		OperationID:   opID,
		ClusterScope:  clusterScope,
		InfobaseScope: infobaseScope,
		Allowed:       allowed,
		CommandLine:   opID,
		ExitCode:      -1,
	}

	if !allowed {
		audit.Stderr = "access denied"
		_ = s.repo.WriteAudit(audit)
		s.logger.Warn("operation denied", "user", user.Username, "operation", opID, "cluster_scope", clusterScope, "infobase_scope", infobaseScope)
		return nil, "execute.denied", nil
	}

	connectionProfileID, _ := parseInt64(values["connection_profile_id"])
	toolchain := models.ToolchainProfile{}
	if connectionProfileID > 0 {
		if profile, err := s.repo.ConnectionProfileByID(connectionProfileID); err == nil && profile.ToolchainID > 0 {
			if item, err := s.repo.ToolchainProfileByID(profile.ToolchainID); err == nil {
				toolchain = item
			}
		}
	}
	result, err := s.exec.Execute(ctx, models.CommandRequest{
		OperationID: opID,
		Values:      values,
		Actor:       user.Username,
		RACPath:     toolchain.RACPath,
		RASPath:     toolchain.RASPath,
	})
	if err != nil {
		s.logger.Error("operation execution failed", "user", user.Username, "operation", opID, "error", err)
	}

	audit.CommandLine = strings.Join(result.Command, " ")
	audit.ExitCode = result.ExitCode
	audit.Stdout = result.Stdout
	audit.Stderr = result.Stderr
	_ = s.repo.WriteAudit(audit)
	s.logger.Info("operation executed", "user", user.Username, "operation", opID, "exit_code", result.ExitCode)

	return &result, "", err
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "*"
}

func parseInt64(value string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(value), 10, 64)
}

func hasAdministrativeAccess(user models.User) bool {
	for _, role := range user.Roles {
		if role.System {
			return true
		}
		for _, permission := range role.Permissions {
			if permission.OperationID == "*" && permission.ClusterScope == "*" && permission.InfobaseScope == "*" {
				return true
			}
		}
	}
	return false
}
