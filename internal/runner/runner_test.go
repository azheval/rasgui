package runner

import (
	"reflect"
	"testing"

	"rasgui/internal/catalog"
)

func TestBuildArgsPreservesClusterInsertHostParameter(t *testing.T) {
	operation, ok := catalog.Find("rac.cluster.insert")
	if !ok {
		t.Fatal("operation not found")
	}

	args, err := buildArgs(operation, map[string]string{
		"host":         "admin-host",
		"admin_port":   "1545",
		"cluster-host": "cluster-host",
		"cluster-port": "1541",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"cluster", "insert", "--host=cluster-host", "--port=1541", "admin-host:1545"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args mismatch\nwant: %#v\n got: %#v", want, args)
	}
}

func TestBuildArgsIncludesSessionTerminateMessage(t *testing.T) {
	operation, ok := catalog.Find("rac.session.terminate")
	if !ok {
		t.Fatal("operation not found")
	}

	args, err := buildArgs(operation, map[string]string{
		"cluster":       "cluster-id",
		"session":       "session-id",
		"error-message": "Maintenance window",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"session", "terminate", "--cluster=cluster-id", "--session=session-id", "--error-message=Maintenance window", "localhost"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args mismatch\nwant: %#v\n got: %#v", want, args)
	}
}

func TestBuildArgsIncludesConnectionProcess(t *testing.T) {
	operation, ok := catalog.Find("rac.connection.disconnect")
	if !ok {
		t.Fatal("operation not found")
	}

	args, err := buildArgs(operation, map[string]string{
		"cluster":       "cluster-id",
		"process":       "process-id",
		"connection":    "connection-id",
		"infobase-user": "ib-admin",
		"infobase-pwd":  "secret",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"connection", "disconnect", "--cluster=cluster-id", "--process=process-id", "--connection=connection-id", "--infobase-user=ib-admin", "--infobase-pwd=secret", "localhost"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args mismatch\nwant: %#v\n got: %#v", want, args)
	}
}

func TestBuildArgsSplitsQuotedExtraArgs(t *testing.T) {
	operation, ok := catalog.Find("rac.infobase.update")
	if !ok {
		t.Fatal("operation not found")
	}

	args, err := buildArgs(operation, map[string]string{
		"cluster":    "cluster-id",
		"infobase":   "ib-id",
		"extra_args": `--denied-message="Maintenance window" --permission-code='let me in'`,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"infobase", "update", "--cluster=cluster-id", "--infobase=ib-id", "localhost", "--denied-message=Maintenance window", "--permission-code=let me in"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args mismatch\nwant: %#v\n got: %#v", want, args)
	}
}
