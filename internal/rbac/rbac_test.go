package rbac

import (
	"testing"

	"rasgui/internal/models"
)

func TestAllowed(t *testing.T) {
	user := models.User{
		Username: "operator",
		Roles: []models.Role{{
			Name: "ib-role",
			Permissions: []models.Permission{{
				OperationID:   "rac.infobase.update",
				ClusterScope:  "cluster-1",
				InfobaseScope: "ib-1",
			}},
		}},
	}

	if !Allowed(user, "rac.infobase.update", "cluster-1", "ib-1") {
		t.Fatal("expected permission granted")
	}
	if Allowed(user, "rac.infobase.update", "cluster-2", "ib-1") {
		t.Fatal("expected permission denied for another cluster")
	}
}
