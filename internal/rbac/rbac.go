package rbac

import "rasgui/internal/models"

func Allowed(user models.User, operationID, clusterScope, infobaseScope string) bool {
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.OperationID != "*" && permission.OperationID != operationID {
				continue
			}
			if !scopeMatch(permission.ClusterScope, clusterScope) {
				continue
			}
			if !scopeMatch(permission.InfobaseScope, infobaseScope) {
				continue
			}
			return true
		}
	}
	return false
}

func scopeMatch(rule, value string) bool {
	if rule == "" && value == "" {
		return true
	}
	if rule == "*" {
		return true
	}
	return rule == value
}
