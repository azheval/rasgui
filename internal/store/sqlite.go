package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"golang.org/x/crypto/bcrypt"

	"rasgui/internal/models"
)

type Store struct {
	db *sql.DB
}

func Open(dataDir string) (*Store, error) {
	dbPath := filepath.Join(dataDir, "rasgui.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.init(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) init() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL,
			system_role INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			user_id INTEGER NOT NULL,
			role_id INTEGER NOT NULL,
			PRIMARY KEY (user_id, role_id)
		);`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_id INTEGER NOT NULL,
			operation_id TEXT NOT NULL,
			cluster_scope TEXT NOT NULL,
			infobase_scope TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS toolchain_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			version TEXT NOT NULL,
			rac_path TEXT NOT NULL,
			ras_path TEXT NOT NULL,
			description TEXT NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS connection_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			host TEXT NOT NULL,
			port INTEGER NOT NULL,
			toolchain_id INTEGER NOT NULL DEFAULT 0,
			description TEXT NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS favorite_commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			operation_id TEXT NOT NULL,
			connection_id INTEGER NOT NULL DEFAULT 0,
			values_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			operation_id TEXT NOT NULL,
			cluster_scope TEXT NOT NULL,
			infobase_scope TEXT NOT NULL,
			command_line TEXT NOT NULL,
			allowed INTEGER NOT NULL,
			exit_code INTEGER NOT NULL,
			stdout TEXT NOT NULL,
			stderr TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL
		);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return s.applyMigrations()
}

type migration struct {
	version int
	name    string
	sql     string
}

func (s *Store) applyMigrations() error {
	migrations := []migration{
		{version: 1, name: "connection_profiles_toolchain_id", sql: `ALTER TABLE connection_profiles ADD COLUMN toolchain_id INTEGER NOT NULL DEFAULT 0`},
	}
	for _, item := range migrations {
		applied, err := s.migrationApplied(item.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if _, err := s.db.Exec(item.sql); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return err
		}
		if _, err := s.db.Exec(`INSERT OR IGNORE INTO schema_migrations(version, name, applied_at) VALUES(?, ?, ?)`, item.version, item.name, time.Now().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) migrationApplied(version int) (bool, error) {
	var found int
	err := s.db.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, version).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return found == 1, nil
}

func (s *Store) SeedAdmin(username, password string) error {
	count := 0
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	roleID, err := s.CreateRole("platform-admin", "Полный доступ ко всем операциям", true)
	if err != nil {
		return err
	}
	if _, err := s.AddPermission(roleID, "*", "*", "*"); err != nil {
		return err
	}
	userID, err := s.CreateUser(username, password)
	if err != nil {
		return err
	}
	return s.AssignRole(userID, roleID)
}

func (s *Store) CreateUser(username, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := s.db.Exec(`INSERT INTO users(username, password_hash, enabled, created_at) VALUES (?, ?, 1, ?)`, username, string(hash), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) Authenticate(username, password string) (models.User, error) {
	var (
		user models.User
		hash string
	)
	err := s.db.QueryRow(`SELECT id, username, password_hash, enabled, created_at FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username, &hash, &user.Enabled, &user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return models.User{}, errors.New("invalid credentials")
	}
	roles, err := s.UserRoles(user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Roles = roles
	return user, nil
}

func (s *Store) UserByID(id int64) (models.User, error) {
	var user models.User
	err := s.db.QueryRow(`SELECT id, username, enabled, created_at FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Username, &user.Enabled, &user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	roles, err := s.UserRoles(user.ID)
	if err != nil {
		return models.User{}, err
	}
	user.Roles = roles
	return user, nil
}

func (s *Store) ListUsers() ([]models.User, error) {
	rows, err := s.db.Query(`SELECT id, username, enabled, created_at FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Enabled, &user.CreatedAt); err != nil {
			return nil, err
		}
		roles, err := s.UserRoles(user.ID)
		if err != nil {
			return nil, err
		}
		user.Roles = roles
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Store) DeleteUser(userID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM user_roles WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM favorite_commands WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) CreateRole(name, description string, system bool) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO roles(name, description, system_role, created_at) VALUES (?, ?, ?, ?)`, name, description, boolToInt(system), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeleteRole(roleID int64) error {
	var systemRole int
	if err := s.db.QueryRow(`SELECT system_role FROM roles WHERE id = ?`, roleID).Scan(&systemRole); err != nil {
		return err
	}
	if systemRole == 1 {
		return errors.New("system role cannot be deleted")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM user_roles WHERE role_id = ?`, roleID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM permissions WHERE role_id = ?`, roleID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM roles WHERE id = ?`, roleID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SeedConnectionProfile(name, host string, port int, description string) error {
	count := 0
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM connection_profiles`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaultToolchainID, _ := s.defaultToolchainID()
	_, err := s.CreateConnectionProfile(name, host, port, defaultToolchainID, description, true)
	return err
}

func (s *Store) SeedToolchainProfile(name, version, racPath, rasPath, description string, isDefault bool) error {
	count := 0
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM toolchain_profiles WHERE name = ?`, name).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := s.CreateToolchainProfile(name, version, racPath, rasPath, description, isDefault)
	return err
}

func (s *Store) CreateToolchainProfile(name, version, racPath, rasPath, description string, isDefault bool) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if isDefault {
		if _, err := tx.Exec(`UPDATE toolchain_profiles SET is_default = 0`); err != nil {
			return 0, err
		}
	}

	res, err := tx.Exec(`INSERT INTO toolchain_profiles(name, version, rac_path, ras_path, description, is_default, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		name, version, racPath, rasPath, description, boolToInt(isDefault), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateToolchainProfile(id int64, name, version, racPath, rasPath, description string, isDefault bool) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if isDefault {
		if _, err := tx.Exec(`UPDATE toolchain_profiles SET is_default = 0`); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE toolchain_profiles SET name = ?, version = ?, rac_path = ?, ras_path = ?, description = ?, is_default = ? WHERE id = ?`,
		name, version, racPath, rasPath, description, boolToInt(isDefault), id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) DeleteToolchainProfile(id int64) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM connection_profiles WHERE toolchain_id = ?`, id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return errors.New("toolchain is used by connection profiles")
	}
	_, err := s.db.Exec(`DELETE FROM toolchain_profiles WHERE id = ?`, id)
	return err
}

func (s *Store) ListToolchainProfiles() ([]models.ToolchainProfile, error) {
	rows, err := s.db.Query(`SELECT id, name, version, rac_path, ras_path, description, is_default, created_at FROM toolchain_profiles ORDER BY is_default DESC, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ToolchainProfile
	for rows.Next() {
		var (
			item      models.ToolchainProfile
			isDefault int
		)
		if err := rows.Scan(&item.ID, &item.Name, &item.Version, &item.RACPath, &item.RASPath, &item.Description, &isDefault, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.IsDefault = isDefault == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ToolchainProfileByID(id int64) (models.ToolchainProfile, error) {
	var (
		item      models.ToolchainProfile
		isDefault int
	)
	err := s.db.QueryRow(`SELECT id, name, version, rac_path, ras_path, description, is_default, created_at FROM toolchain_profiles WHERE id = ?`, id).
		Scan(&item.ID, &item.Name, &item.Version, &item.RACPath, &item.RASPath, &item.Description, &isDefault, &item.CreatedAt)
	if err != nil {
		return models.ToolchainProfile{}, err
	}
	item.IsDefault = isDefault == 1
	return item, nil
}

func (s *Store) CreateConnectionProfile(name, host string, port int, toolchainID int64, description string, isDefault bool) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if isDefault {
		if _, err := tx.Exec(`UPDATE connection_profiles SET is_default = 0`); err != nil {
			return 0, err
		}
	}
	if toolchainID == 0 {
		toolchainID, _ = s.defaultToolchainID()
	}

	res, err := tx.Exec(`INSERT INTO connection_profiles(name, host, port, toolchain_id, description, is_default, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		name, host, port, toolchainID, description, boolToInt(isDefault), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateConnectionProfile(id int64, name, host string, port int, toolchainID int64, description string, isDefault bool) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if isDefault {
		if _, err := tx.Exec(`UPDATE connection_profiles SET is_default = 0`); err != nil {
			return err
		}
	}
	if toolchainID == 0 {
		toolchainID, _ = s.defaultToolchainID()
	}
	if _, err := tx.Exec(`UPDATE connection_profiles SET name = ?, host = ?, port = ?, toolchain_id = ?, description = ?, is_default = ? WHERE id = ?`,
		name, host, port, toolchainID, description, boolToInt(isDefault), id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) DeleteConnectionProfile(id int64) error {
	_, err := s.db.Exec(`DELETE FROM connection_profiles WHERE id = ?`, id)
	return err
}

func (s *Store) ListConnectionProfiles() ([]models.ConnectionProfile, error) {
	rows, err := s.db.Query(`
		SELECT cp.id, cp.name, cp.host, cp.port, cp.toolchain_id, COALESCE(tp.name, ''), cp.description, cp.is_default, cp.created_at
		FROM connection_profiles cp
		LEFT JOIN toolchain_profiles tp ON tp.id = cp.toolchain_id
		ORDER BY cp.is_default DESC, cp.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ConnectionProfile
	for rows.Next() {
		var (
			item      models.ConnectionProfile
			isDefault int
		)
		if err := rows.Scan(&item.ID, &item.Name, &item.Host, &item.Port, &item.ToolchainID, &item.ToolchainName, &item.Description, &isDefault, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.IsDefault = isDefault == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ConnectionProfileByID(id int64) (models.ConnectionProfile, error) {
	var (
		item      models.ConnectionProfile
		isDefault int
	)
	err := s.db.QueryRow(`
		SELECT cp.id, cp.name, cp.host, cp.port, cp.toolchain_id, COALESCE(tp.name, ''), cp.description, cp.is_default, cp.created_at
		FROM connection_profiles cp
		LEFT JOIN toolchain_profiles tp ON tp.id = cp.toolchain_id
		WHERE cp.id = ?`, id).
		Scan(&item.ID, &item.Name, &item.Host, &item.Port, &item.ToolchainID, &item.ToolchainName, &item.Description, &isDefault, &item.CreatedAt)
	if err != nil {
		return models.ConnectionProfile{}, err
	}
	item.IsDefault = isDefault == 1
	return item, nil
}

func (s *Store) CreateFavoriteCommand(userID int64, name, operationID string, connectionID int64, values map[string]string) (int64, error) {
	payload, err := json.Marshal(values)
	if err != nil {
		return 0, err
	}
	res, err := s.db.Exec(`INSERT INTO favorite_commands(user_id, name, operation_id, connection_id, values_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, name, operationID, connectionID, string(payload), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListFavoriteCommands(userID int64) ([]models.FavoriteCommand, error) {
	rows, err := s.db.Query(`SELECT id, user_id, name, operation_id, connection_id, values_json, created_at FROM favorite_commands WHERE user_id = ? ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.FavoriteCommand
	for rows.Next() {
		var (
			item     models.FavoriteCommand
			rawValue string
		)
		if err := rows.Scan(&item.ID, &item.UserID, &item.Name, &item.OperationID, &item.ConnectionID, &rawValue, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Values = map[string]string{}
		if rawValue != "" {
			if err := json.Unmarshal([]byte(rawValue), &item.Values); err != nil {
				return nil, err
			}
			for key := range item.Values {
				if isSensitiveFavoriteKey(key) || key == "host" || key == "admin_port" {
					delete(item.Values, key)
				}
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) DeleteFavoriteCommand(userID, favoriteID int64) error {
	_, err := s.db.Exec(`DELETE FROM favorite_commands WHERE id = ? AND user_id = ?`, favoriteID, userID)
	return err
}

func (s *Store) ListRoles() ([]models.Role, error) {
	rows, err := s.db.Query(`SELECT id, name, description, system_role, created_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.System, &role.CreatedAt); err != nil {
			return nil, err
		}
		perms, err := s.RolePermissions(role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = perms
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *Store) AssignRole(userID, roleID int64) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO user_roles(user_id, role_id) VALUES (?, ?)`, userID, roleID)
	return err
}

func (s *Store) UserRoles(userID int64) ([]models.Role, error) {
	rows, err := s.db.Query(`
		SELECT r.id, r.name, r.description, r.system_role, r.created_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = ?
		ORDER BY r.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.System, &role.CreatedAt); err != nil {
			return nil, err
		}
		perms, err := s.RolePermissions(role.ID)
		if err != nil {
			return nil, err
		}
		role.Permissions = perms
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *Store) AddPermission(roleID int64, operationID, clusterScope, infobaseScope string) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO permissions(role_id, operation_id, cluster_scope, infobase_scope, created_at) VALUES (?, ?, ?, ?, ?)`,
		roleID, operationID, defaultScope(clusterScope), defaultScope(infobaseScope), time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("add permission: %w", err)
	}
	return res.LastInsertId()
}

func (s *Store) ReplaceRolePermissions(roleID int64, operationIDs []string, clusterScopes, infobaseScopes []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM permissions WHERE role_id = ?`, roleID); err != nil {
		return err
	}
	if len(clusterScopes) == 0 {
		clusterScopes = []string{"*"}
	}
	if len(infobaseScopes) == 0 {
		infobaseScopes = []string{"*"}
	}
	for _, operationID := range operationIDs {
		if strings.TrimSpace(operationID) == "" {
			continue
		}
		for _, clusterScope := range clusterScopes {
			for _, infobaseScope := range infobaseScopes {
				if _, err := tx.Exec(`INSERT INTO permissions(role_id, operation_id, cluster_scope, infobase_scope, created_at) VALUES (?, ?, ?, ?, ?)`,
					roleID, operationID, defaultScope(clusterScope), defaultScope(infobaseScope), time.Now().Format(time.RFC3339)); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}

func (s *Store) ReplaceRoleScopeSet(roleID int64, oldClusterScope, oldInfobaseScope, newClusterScope, newInfobaseScope string, operationIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	oldClusterScope = defaultScope(strings.TrimSpace(oldClusterScope))
	oldInfobaseScope = defaultScope(strings.TrimSpace(oldInfobaseScope))
	newClusterScope = defaultScope(strings.TrimSpace(newClusterScope))
	newInfobaseScope = defaultScope(strings.TrimSpace(newInfobaseScope))

	if oldClusterScope != "" && oldInfobaseScope != "" {
		if _, err := tx.Exec(`DELETE FROM permissions WHERE role_id = ? AND cluster_scope = ? AND infobase_scope = ?`,
			roleID, oldClusterScope, oldInfobaseScope); err != nil {
			return err
		}
	}
	for _, operationID := range operationIDs {
		if strings.TrimSpace(operationID) == "" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO permissions(role_id, operation_id, cluster_scope, infobase_scope, created_at) VALUES (?, ?, ?, ?, ?)`,
			roleID, operationID, newClusterScope, newInfobaseScope, time.Now().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) DeleteRoleScopeSet(roleID int64, clusterScope, infobaseScope string) error {
	_, err := s.db.Exec(`DELETE FROM permissions WHERE role_id = ? AND cluster_scope = ? AND infobase_scope = ?`,
		roleID, defaultScope(strings.TrimSpace(clusterScope)), defaultScope(strings.TrimSpace(infobaseScope)))
	return err
}

func (s *Store) RolePermissions(roleID int64) ([]models.Permission, error) {
	rows, err := s.db.Query(`SELECT id, role_id, operation_id, cluster_scope, infobase_scope, created_at FROM permissions WHERE role_id = ? ORDER BY operation_id`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var permission models.Permission
		if err := rows.Scan(&permission.ID, &permission.RoleID, &permission.OperationID, &permission.ClusterScope, &permission.InfobaseScope, &permission.CreatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (s *Store) WriteAudit(item models.AuditLog) error {
	_, err := s.db.Exec(`INSERT INTO audit_log(username, operation_id, cluster_scope, infobase_scope, command_line, allowed, exit_code, stdout, stderr, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.Username, item.OperationID, item.ClusterScope, item.InfobaseScope, item.CommandLine, boolToInt(item.Allowed), item.ExitCode, item.Stdout, item.Stderr, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) ListAudit(limit int) ([]models.AuditLog, error) {
	rows, err := s.db.Query(`SELECT id, username, operation_id, cluster_scope, infobase_scope, command_line, allowed, exit_code, stdout, stderr, created_at
	FROM audit_log ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AuditLog
	for rows.Next() {
		var (
			item    models.AuditLog
			allowed int
		)
		if err := rows.Scan(&item.ID, &item.Username, &item.OperationID, &item.ClusterScope, &item.InfobaseScope, &item.CommandLine, &allowed, &item.ExitCode, &item.Stdout, &item.Stderr, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Allowed = allowed == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func defaultScope(value string) string {
	if value == "" {
		return "*"
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func isSensitiveFavoriteKey(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if strings.Contains(lower, "pwd") || strings.Contains(lower, "password") {
		return true
	}
	switch lower {
	case "cluster-user", "infobase-user", "db-user", "agent-user", "os-user":
		return true
	default:
		return false
	}
}

func (s *Store) defaultToolchainID() (int64, error) {
	var id int64
	err := s.db.QueryRow(`SELECT id FROM toolchain_profiles ORDER BY is_default DESC, id LIMIT 1`).Scan(&id)
	return id, err
}
