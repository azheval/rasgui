package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"rasgui/internal/app"
	"rasgui/internal/config"
	"rasgui/internal/i18n"
	"rasgui/internal/models"
)

type Server struct {
	cfg      config.Config
	app      *app.Service
	logger   *slog.Logger
	sessions map[string]sessionState
	mu       sync.RWMutex
}

type sessionState struct {
	UserID    int64
	CSRFToken string
	ExpiresAt time.Time
}

type pageData struct {
	TitleKey           string
	CurrentPath        string
	Lang               string
	CSRFToken          string
	CurrentUser        models.User
	CanManageUsers     bool
	CanManageRoles     bool
	CanManageConnections bool
	CanViewAudit       bool
	Users              []models.User
	Roles              []models.Role
	ConnectionProfiles []models.ConnectionProfile
	ToolchainProfiles  []models.ToolchainProfile
	DiscoverableToolchains []config.ToolchainSeed
	Favorites          []models.FavoriteCommand
	Audit              []models.AuditLog
	Operations         []models.Operation
	OperationsJS       template.JS
	ProfilesJS         template.JS
	FavoritesJS        template.JS
	FormValuesJS       template.JS
	LastResult         *models.CommandResult
	ParsedResultRows   []parsedRow
	Message            string
	DefaultPort        int
	DefaultRAC         string
	DefaultRAS         string
	DefaultRASHost     string
	DefaultRASPort     int
	RoleMatrix         []roleMatrixSection
}

type parsedRow struct {
	Key   string
	Value string
}

type roleMatrixSection struct {
	Mode         string
	Operations   []roleMatrixOperation
	BundleLabels []string
	ClusterScope string
	InfobaseScope string
}

type roleMatrixOperation struct {
	Operation models.Operation
	Bundle    string
}

type roleScopeSetView struct {
	ClusterScope  string
	InfobaseScope string
	OperationIDs  map[string]bool
}

func New(cfg config.Config, appSvc *app.Service, logger *slog.Logger) (*Server, error) {
	return &Server{
		cfg:      cfg,
		app:      appSvc,
		logger:   logger,
		sessions: map[string]sessionState{},
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/set-lang", s.handleSetLang)
	mux.HandleFunc("/login", s.handleLogin)
	mux.HandleFunc("/logout", s.handleLogout)
	mux.HandleFunc("/", s.auth(s.handleDashboard))
	mux.HandleFunc("/catalog", s.auth(s.handleCatalog))
	mux.HandleFunc("/users", s.auth(s.handleUsers))
	mux.HandleFunc("/roles", s.auth(s.handleRoles))
	mux.HandleFunc("/connections", s.auth(s.handleConnections))
	mux.HandleFunc("/execute", s.auth(s.handleExecute))
	mux.HandleFunc("/audit", s.auth(s.handleAudit))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(s.cfg.WorkDir, "web", "static")))))
	return mux
}

func (s *Server) handleSetLang(w http.ResponseWriter, r *http.Request) {
	lang := i18n.Normalize(r.URL.Query().Get("lang"))
	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}
	http.SetCookie(w, &http.Cookie{Name: "lang", Value: lang, Path: "/", SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	lang := s.resolveLang(r)
	if r.Method == http.MethodGet {
		s.render(w, "login.html", pageData{TitleKey: "page.login", Lang: lang, CurrentPath: "/login", CSRFToken: s.ensureAnonCSRF(w, r)})
		return
	}
	if !s.verifyAnonCSRF(r) {
		http.Error(w, "csrf validation failed", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := s.app.Authenticate(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		s.render(w, "login.html", pageData{
			TitleKey:    "page.login",
			Lang:        lang,
			CurrentPath: "/login",
			CSRFToken:   s.ensureAnonCSRF(w, r),
			Message:     i18n.T(lang, "login.invalid"),
		})
		return
	}

	sid := randomToken()
	csrfToken := randomToken()
	s.mu.Lock()
	s.sessions[sid] = sessionState{UserID: user.ID, CSRFToken: csrfToken, ExpiresAt: time.Now().Add(12 * time.Hour)}
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: sid, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	http.SetCookie(w, &http.Cookie{Name: "anon_csrf", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && !s.verifyCSRF(r) {
		http.Error(w, "csrf validation failed", http.StatusForbidden)
		return
	}
	if cookie, err := r.Cookie("session_id"); err == nil {
		s.mu.Lock()
		delete(s.sessions, cookie.Value)
		s.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request, user models.User) {
	lang := s.resolveLang(r)
	users, _ := s.app.ListUsers()
	roles, _ := s.app.ListRoles()
	profiles, _ := s.app.ListConnectionProfiles()
	toolchains, _ := s.app.ListToolchainProfiles()
	discoverable := undiscoveredToolchains(config.DiscoverLocalToolchains(s.cfg), toolchains)
	audit, _ := s.app.ListAudit(10)
	favorites, _ := s.app.FavoriteCommandsForUser(user)
	csrfToken, _ := s.sessionToken(r)
	s.render(w, "dashboard.html", pageData{
		TitleKey:           "page.dashboard",
		CurrentPath:        "/",
		Lang:               lang,
		CSRFToken:          csrfToken,
		CurrentUser:        user,
		CanManageUsers:     s.app.CanManageUsers(user),
		CanManageRoles:     s.app.CanManageRoles(user),
		CanManageConnections: s.app.CanManageConnections(user),
		CanViewAudit:       s.app.CanViewAudit(user),
		Users:              users,
		Roles:              roles,
		ConnectionProfiles: profiles,
		ToolchainProfiles: toolchains,
		DiscoverableToolchains: discoverable,
		Favorites:          favorites,
		Audit:              audit,
		Operations:         s.app.OperationsForUser(user),
		DefaultPort:        s.cfg.HTTPPort,
		DefaultRAC:         s.cfg.RACPath,
		DefaultRAS:         s.cfg.RASPath,
		DefaultRASHost:     s.cfg.DefaultRASHost,
		DefaultRASPort:     s.cfg.DefaultRASPort,
	})
}

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request, user models.User) {
	lang := s.resolveLang(r)
	csrfToken, _ := s.sessionToken(r)
	operations := s.app.OperationsForUser(user)
	opsJSON, _ := json.Marshal(operations)
	s.render(w, "catalog.html", pageData{
		TitleKey:     "page.catalog",
		CurrentPath:  "/catalog",
		Lang:         lang,
		CSRFToken:    csrfToken,
		CurrentUser:  user,
		CanManageUsers: s.app.CanManageUsers(user),
		CanManageRoles: s.app.CanManageRoles(user),
		CanManageConnections: s.app.CanManageConnections(user),
		CanViewAudit: s.app.CanViewAudit(user),
		Operations:   operations,
		OperationsJS: template.JS(opsJSON),
	})
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request, user models.User) {
	if !s.app.CanManageUsers(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lang := s.resolveLang(r)
	if r.Method == http.MethodPost {
		if !s.verifyCSRF(r) {
			http.Error(w, "csrf validation failed", http.StatusForbidden)
			return
		}
		if err := r.ParseForm(); err == nil {
			switch r.FormValue("action") {
			case "delete_user":
				targetUserID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
				_ = s.app.DeleteUser(user.ID, targetUserID)
			default:
				roleID, _ := strconv.ParseInt(r.FormValue("role_id"), 10, 64)
				_ = s.app.CreateUser(r.FormValue("username"), r.FormValue("password"), roleID)
			}
		}
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}
	users, _ := s.app.ListUsers()
	roles, _ := s.app.ListRoles()
	csrfToken, _ := s.sessionToken(r)
	s.render(w, "users.html", pageData{
		TitleKey:    "page.users",
		CurrentPath: "/users",
		Lang:        lang,
		CSRFToken:   csrfToken,
		CurrentUser: user,
		CanManageUsers: true,
		CanManageRoles: s.app.CanManageRoles(user),
		CanManageConnections: s.app.CanManageConnections(user),
		CanViewAudit: s.app.CanViewAudit(user),
		Users:       users,
		Roles:       roles,
	})
}

func (s *Server) handleRoles(w http.ResponseWriter, r *http.Request, user models.User) {
	if !s.app.CanManageRoles(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lang := s.resolveLang(r)
	allOperations := s.app.Operations()
	if r.Method == http.MethodPost {
		if !s.verifyCSRF(r) {
			http.Error(w, "csrf validation failed", http.StatusForbidden)
			return
		}
		_ = r.ParseForm()
		switch r.FormValue("action") {
		case "create_role":
			_ = s.app.CreateRole(r.FormValue("name"), r.FormValue("description"))
		case "add_permission":
			roleID, _ := strconv.ParseInt(r.FormValue("role_id"), 10, 64)
			_ = s.app.AddPermission(roleID, r.FormValue("operation_id"), r.FormValue("cluster_scope"), r.FormValue("infobase_scope"))
		case "save_scope_set":
			roleID, _ := strconv.ParseInt(r.FormValue("role_id"), 10, 64)
			_ = s.app.ReplaceRoleScopeSet(
				roleID,
				r.FormValue("old_cluster_scope"),
				r.FormValue("old_infobase_scope"),
				r.FormValue("cluster_scope"),
				r.FormValue("infobase_scope"),
				r.PostForm["operation_ids"],
			)
		case "delete_scope_set":
			roleID, _ := strconv.ParseInt(r.FormValue("role_id"), 10, 64)
			_ = s.app.DeleteRoleScopeSet(roleID, r.FormValue("cluster_scope"), r.FormValue("infobase_scope"))
		case "delete_role":
			roleID, _ := strconv.ParseInt(r.FormValue("role_id"), 10, 64)
			_ = s.app.DeleteRole(roleID)
		}
		http.Redirect(w, r, "/roles", http.StatusSeeOther)
		return
	}
	roles, _ := s.app.ListRoles()
	opsJSON, _ := json.Marshal(allOperations)
	csrfToken, _ := s.sessionToken(r)
	s.render(w, "roles.html", pageData{
		TitleKey:     "page.roles",
		CurrentPath:  "/roles",
		Lang:         lang,
		CSRFToken:    csrfToken,
		CurrentUser:  user,
		CanManageUsers: s.app.CanManageUsers(user),
		CanManageRoles: true,
		CanManageConnections: s.app.CanManageConnections(user),
		CanViewAudit: s.app.CanViewAudit(user),
		Roles:        roles,
		Operations:   allOperations,
		RoleMatrix:   buildRoleMatrix(allOperations),
		OperationsJS: template.JS(opsJSON),
	})
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request, user models.User) {
	if !s.app.CanManageConnections(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lang := s.resolveLang(r)
	message := ""
	if r.Method == http.MethodPost {
		if !s.verifyCSRF(r) {
			http.Error(w, "csrf validation failed", http.StatusForbidden)
			return
		}
		if err := r.ParseForm(); err == nil {
			switch r.FormValue("action") {
			case "import_discovered_toolchain":
				if err := s.app.CreateToolchainProfile(
					r.FormValue("toolchain_name"),
					r.FormValue("toolchain_version"),
					r.FormValue("rac_path"),
					r.FormValue("ras_path"),
					r.FormValue("toolchain_description"),
					r.FormValue("toolchain_is_default") == "on",
				); err != nil {
					message = err.Error()
				}
			case "create_toolchain":
				if err := s.app.CreateToolchainProfile(
					r.FormValue("toolchain_name"),
					r.FormValue("toolchain_version"),
					r.FormValue("rac_path"),
					r.FormValue("ras_path"),
					r.FormValue("toolchain_description"),
					r.FormValue("toolchain_is_default") == "on",
				); err != nil {
					message = err.Error()
				}
			case "update_toolchain":
				id, _ := strconv.ParseInt(r.FormValue("toolchain_id"), 10, 64)
				if err := s.app.UpdateToolchainProfile(id, r.FormValue("toolchain_name"), r.FormValue("toolchain_version"), r.FormValue("rac_path"), r.FormValue("ras_path"), r.FormValue("toolchain_description"), r.FormValue("toolchain_is_default") == "on"); err != nil {
					message = err.Error()
				}
			case "delete_toolchain":
				id, _ := strconv.ParseInt(r.FormValue("toolchain_id"), 10, 64)
				if err := s.app.DeleteToolchainProfile(id); err != nil {
					message = err.Error()
				}
			case "check_toolchain":
				id, _ := strconv.ParseInt(r.FormValue("toolchain_id"), 10, 64)
				if msg, err := s.app.CheckToolchainProfile(id); err != nil {
					message = err.Error()
				} else {
					message = msg
				}
			case "update_connection":
				id, _ := strconv.ParseInt(r.FormValue("connection_id"), 10, 64)
				port, _ := strconv.Atoi(r.FormValue("port"))
				toolchainID, _ := strconv.ParseInt(r.FormValue("toolchain_id"), 10, 64)
				if err := s.app.UpdateConnectionProfile(id, r.FormValue("name"), r.FormValue("host"), port, toolchainID, r.FormValue("description"), r.FormValue("is_default") == "on"); err != nil {
					message = err.Error()
				}
			case "delete_connection":
				id, _ := strconv.ParseInt(r.FormValue("connection_id"), 10, 64)
				if err := s.app.DeleteConnectionProfile(id); err != nil {
					message = err.Error()
				}
			case "check_connection":
				id, _ := strconv.ParseInt(r.FormValue("connection_id"), 10, 64)
				if msg, err := s.app.CheckConnectionProfile(id); err != nil {
					message = err.Error()
				} else {
					message = msg
				}
			default:
				port, _ := strconv.Atoi(r.FormValue("port"))
				toolchainID, _ := strconv.ParseInt(r.FormValue("toolchain_id"), 10, 64)
				isDefault := r.FormValue("is_default") == "on"
				if err := s.app.CreateConnectionProfileWithToolchain(r.FormValue("name"), r.FormValue("host"), port, toolchainID, r.FormValue("description"), isDefault); err != nil {
					message = err.Error()
				}
			}
		}
		if message == "" && r.FormValue("action") != "check_toolchain" && r.FormValue("action") != "check_connection" {
			http.Redirect(w, r, "/connections", http.StatusSeeOther)
			return
		}
	}
	profiles, _ := s.app.ListConnectionProfiles()
	toolchains, _ := s.app.ListToolchainProfiles()
	discoverable := undiscoveredToolchains(config.DiscoverLocalToolchains(s.cfg), toolchains)
	csrfToken, _ := s.sessionToken(r)
	s.render(w, "connections.html", pageData{
		TitleKey:           "page.connections",
		CurrentPath:        "/connections",
		Lang:               lang,
		CSRFToken:          csrfToken,
		CurrentUser:        user,
		CanManageUsers:     s.app.CanManageUsers(user),
		CanManageRoles:     s.app.CanManageRoles(user),
		CanManageConnections: true,
		CanViewAudit:       s.app.CanViewAudit(user),
		ConnectionProfiles: profiles,
		ToolchainProfiles: toolchains,
		DiscoverableToolchains: discoverable,
		DefaultRASPort:     s.cfg.DefaultRASPort,
		DefaultRAC:         s.cfg.RACPath,
		DefaultRAS:         s.cfg.RASPath,
		Message:            message,
	})
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request, user models.User) {
	lang := s.resolveLang(r)
	operations := s.app.OperationsForUser(user)
	var lastResult *models.CommandResult
	var parsedRows []parsedRow
	formValues := map[string]string{}
	message := ""
	if r.Method == http.MethodPost {
		if !s.verifyCSRF(r) {
			http.Error(w, "csrf validation failed", http.StatusForbidden)
			return
		}
		_ = r.ParseForm()
		formValues = buildFormState(r.PostForm)
		switch r.FormValue("action") {
		case "save_favorite":
			connectionID, _ := strconv.ParseInt(r.FormValue("favorite_connection_id"), 10, 64)
			values := buildFavoriteValues(r.PostForm)
			if err := s.app.CreateFavoriteCommand(user.ID, r.FormValue("favorite_name"), r.FormValue("operation_id"), connectionID, values); err != nil {
				message = err.Error()
			} else {
				message = i18n.T(lang, "execute.saved")
				formValues["favorite_name"] = ""
			}
		case "delete_favorite":
			favoriteID, _ := strconv.ParseInt(r.FormValue("favorite_id"), 10, 64)
			if err := s.app.DeleteFavoriteCommand(user.ID, favoriteID); err != nil {
				message = err.Error()
			}
		default:
			values := map[string]string{}
			for key, item := range r.PostForm {
				if len(item) > 0 {
					values[key] = item[0]
				}
			}
			result, businessMessage, err := s.app.ExecuteOperation(context.Background(), user, values)
			if businessMessage != "" {
				message = i18n.T(lang, businessMessage)
			}
			if err != nil {
				message = err.Error()
			}
			if result != nil {
				lastResult = result
				parsedRows = parseKeyValueOutput(result.Stdout)
			}
		}
	}
	opsJSON, _ := json.Marshal(operations)
	profiles, _ := s.app.ListConnectionProfiles()
	toolchains, _ := s.app.ListToolchainProfiles()
	profilesJSON, _ := json.Marshal(profiles)
	favorites, _ := s.app.FavoriteCommandsForUser(user)
	favoritesJSON, _ := json.Marshal(favorites)
	formValuesJSON, _ := json.Marshal(formValues)
	csrfToken, _ := s.sessionToken(r)
	s.render(w, "execute.html", pageData{
		TitleKey:           "page.execute",
		CurrentPath:        "/execute",
		Lang:               lang,
		CSRFToken:          csrfToken,
		CurrentUser:        user,
		CanManageUsers:     s.app.CanManageUsers(user),
		CanManageRoles:     s.app.CanManageRoles(user),
		CanManageConnections: s.app.CanManageConnections(user),
		CanViewAudit:       s.app.CanViewAudit(user),
		ConnectionProfiles: profiles,
		ToolchainProfiles: toolchains,
		Favorites:          favorites,
		Operations:         operations,
		OperationsJS:       template.JS(opsJSON),
		ProfilesJS:         template.JS(profilesJSON),
		FavoritesJS:        template.JS(favoritesJSON),
		FormValuesJS:       template.JS(formValuesJSON),
		LastResult:         lastResult,
		ParsedResultRows:   parsedRows,
		Message:            message,
		DefaultRASHost:     s.cfg.DefaultRASHost,
		DefaultRASPort:     s.cfg.DefaultRASPort,
	})
}

func (s *Server) handleAudit(w http.ResponseWriter, r *http.Request, user models.User) {
	if !s.app.CanViewAudit(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	lang := s.resolveLang(r)
	audit, _ := s.app.ListAudit(100)
	csrfToken, _ := s.sessionToken(r)
	s.render(w, "audit.html", pageData{
		TitleKey:    "page.audit",
		CurrentPath: "/audit",
		Lang:        lang,
		CSRFToken:   csrfToken,
		CurrentUser: user,
		CanManageUsers: s.app.CanManageUsers(user),
		CanManageRoles: s.app.CanManageRoles(user),
		CanManageConnections: s.app.CanManageConnections(user),
		CanViewAudit: true,
		Audit:       audit,
	})
}

func (s *Server) auth(next func(http.ResponseWriter, *http.Request, models.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		s.mu.RLock()
		session, ok := s.sessions[cookie.Value]
		s.mu.RUnlock()
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if time.Now().After(session.ExpiresAt) {
			s.mu.Lock()
			delete(s.sessions, cookie.Value)
			s.mu.Unlock()
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		user, err := s.app.UserByID(session.UserID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		session.ExpiresAt = time.Now().Add(12 * time.Hour)
		s.mu.Lock()
		s.sessions[cookie.Value] = session
		s.mu.Unlock()
		next(w, r, user)
	}
}

func (s *Server) render(w http.ResponseWriter, name string, data pageData) {
	tmpl, err := template.New("base.html").Funcs(template.FuncMap{
		"t": func(key string) string { return i18n.T(data.Lang, key) },
		"hasPermission": roleHasPermission,
		"roleScopeSummary": roleScopeSummary,
		"roleScopeSets": roleScopeSets,
		"scopeSetHas": scopeSetHas,
	}).ParseFiles(
		filepath.Join(s.cfg.WorkDir, "web", "templates", "base.html"),
		filepath.Join(s.cfg.WorkDir, "web", "templates", name),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) resolveLang(r *http.Request) string {
	if value := r.URL.Query().Get("lang"); value != "" {
		return i18n.Normalize(value)
	}
	if cookie, err := r.Cookie("lang"); err == nil {
		return i18n.Normalize(cookie.Value)
	}
	return "ru"
}

func (s *Server) sessionToken(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[cookie.Value]
	if !ok {
		return "", false
	}
	return session.CSRFToken, true
}

func (s *Server) verifyCSRF(r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		return false
	}
	token, ok := s.sessionToken(r)
	if !ok {
		return false
	}
	return token != "" && token == r.FormValue("csrf_token")
}

func (s *Server) ensureAnonCSRF(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie("anon_csrf"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	token := randomToken()
	http.SetCookie(w, &http.Cookie{Name: "anon_csrf", Value: token, Path: "/", SameSite: http.SameSiteLaxMode})
	return token
}

func (s *Server) verifyAnonCSRF(r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		return false
	}
	cookie, err := r.Cookie("anon_csrf")
	if err != nil || cookie.Value == "" {
		return false
	}
	return cookie.Value == r.FormValue("csrf_token")
}

func randomToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func parseKeyValueOutput(input string) []parsedRow {
	lines := strings.Split(input, "\n")
	rows := make([]parsedRow, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, ":") {
			return nil
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		rows = append(rows, parsedRow{Key: key, Value: value})
	}
	if len(rows) < 2 {
		return nil
	}
	return rows
}

func buildRoleMatrix(operations []models.Operation) []roleMatrixSection {
	grouped := map[string][]roleMatrixOperation{}
	for _, item := range operations {
		grouped[item.Mode] = append(grouped[item.Mode], roleMatrixOperation{
			Operation: item,
			Bundle:    bundleForOperation(item),
		})
	}
	order := make([]string, 0, len(grouped))
	for mode := range grouped {
		order = append(order, mode)
	}
	sort.Strings(order)
	result := make([]roleMatrixSection, 0, len(order))
	for _, mode := range order {
		bundleSet := map[string]struct{}{}
		for _, item := range grouped[mode] {
			bundleSet[item.Bundle] = struct{}{}
		}
		bundles := make([]string, 0, len(bundleSet))
		for bundle := range bundleSet {
			bundles = append(bundles, bundle)
		}
		sort.Strings(bundles)
		result = append(result, roleMatrixSection{
			Mode:         mode,
			Operations:   grouped[mode],
			BundleLabels: bundles,
		})
	}
	return result
}

func bundleForOperation(item models.Operation) string {
	id := item.ID
	switch {
	case strings.Contains(id, ".admin.") || strings.Contains(id, ".profile.") || strings.Contains(id, ".rule."):
		return "access"
	case strings.HasSuffix(id, ".list") || strings.HasSuffix(id, ".info") || strings.Contains(id, ".summary.") || strings.HasSuffix(id, ".values") || strings.HasSuffix(id, ".version") || strings.Contains(id, "get-dirs"):
		return "view"
	case strings.HasSuffix(id, ".create") || strings.HasSuffix(id, ".insert") || strings.HasSuffix(id, ".register") || strings.HasSuffix(id, ".drop") || strings.HasSuffix(id, ".remove"):
		return "lifecycle"
	case strings.HasSuffix(id, ".disconnect") || strings.HasSuffix(id, ".terminate") || strings.Contains(id, "interrupt") || strings.Contains(id, "turn-off") || strings.Contains(id, "clear-unused-space"):
		return "control"
	default:
		return "change"
	}
}

func roleHasPermission(role models.Role, operationID string) bool {
	for _, permission := range role.Permissions {
		if permission.OperationID == operationID || permission.OperationID == "*" {
			return true
		}
	}
	return false
}

func roleScopeSummary(role models.Role, dimension string) string {
	items := map[string]struct{}{}
	for _, permission := range role.Permissions {
		value := permission.ClusterScope
		if dimension == "infobase" {
			value = permission.InfobaseScope
		}
		if strings.TrimSpace(value) == "" {
			value = "*"
		}
		items[value] = struct{}{}
	}
	if len(items) == 0 {
		return "*"
	}
	values := make([]string, 0, len(items))
	for item := range items {
		values = append(values, item)
	}
	sort.Strings(values)
	return strings.Join(values, ", ")
}

func roleScopeSets(role models.Role) []roleScopeSetView {
	grouped := map[string]*roleScopeSetView{}
	order := make([]string, 0)
	for _, permission := range role.Permissions {
		key := permission.ClusterScope + "\x00" + permission.InfobaseScope
		set, ok := grouped[key]
		if !ok {
			set = &roleScopeSetView{
				ClusterScope:  permission.ClusterScope,
				InfobaseScope: permission.InfobaseScope,
				OperationIDs:  map[string]bool{},
			}
			grouped[key] = set
			order = append(order, key)
		}
		set.OperationIDs[permission.OperationID] = true
	}
	sort.Strings(order)
	result := make([]roleScopeSetView, 0, len(order))
	for _, key := range order {
		result = append(result, *grouped[key])
	}
	return result
}

func scopeSetHas(set roleScopeSetView, operationID string) bool {
	return set.OperationIDs[operationID]
}

func splitScopes(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	})
	if len(parts) == 0 {
		return []string{"*"}
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	if len(result) == 0 {
		return []string{"*"}
	}
	return result
}

func buildFavoriteValues(form map[string][]string) map[string]string {
	values := map[string]string{}
	for key, item := range form {
		if len(item) == 0 {
			continue
		}
		if key == "csrf_token" || key == "action" || key == "favorite_name" || key == "favorite_connection_id" || key == "favorite_id" || key == "connection_profile_id" || key == "wizard_args_json" {
			continue
		}
		if key == "host" || key == "admin_port" {
			continue
		}
		if isSensitiveField(key) {
			continue
		}
		values[key] = item[0]
	}
	return values
}

func buildFormState(form map[string][]string) map[string]string {
	values := map[string]string{}
	for key, item := range form {
		if len(item) == 0 {
			continue
		}
		if key == "csrf_token" || key == "action" || key == "favorite_id" || key == "wizard_args_json" {
			continue
		}
		if isSensitiveField(key) {
			continue
		}
		values[key] = item[0]
	}
	return values
}

func isSensitiveField(name string) bool {
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

func undiscoveredToolchains(found []config.ToolchainSeed, existing []models.ToolchainProfile) []config.ToolchainSeed {
	seen := map[string]struct{}{}
	for _, item := range existing {
		key := strings.TrimSpace(item.RACPath) + "\x00" + strings.TrimSpace(item.RASPath)
		seen[key] = struct{}{}
	}
	result := make([]config.ToolchainSeed, 0, len(found))
	for _, item := range found {
		key := strings.TrimSpace(item.RACPath) + "\x00" + strings.TrimSpace(item.RASPath)
		if _, ok := seen[key]; ok {
			continue
		}
		result = append(result, item)
	}
	return result
}
