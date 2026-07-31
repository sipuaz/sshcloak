package cli

import (
	"bytes"
	"strings"
	"testing"
)

func executeRootForHelpTest(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	cmd := newRootCmd()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestRootHelpFlagShowsHelp(t *testing.T) {
	stdout, stderr, err := executeRootForHelpTest(t, "--help")
	if err != nil {
		t.Fatalf("expected nil error, got %v\nstdout=%q\nstderr=%q", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected usage in stdout, got %q", stdout)
	}
	if !strings.Contains(stdout, "sshcloak") {
		t.Fatalf("expected root help output, got %q", stdout)
	}
}

func TestConnectHelpFlagShowsHelp(t *testing.T) {
	stdout, stderr, err := executeRootForHelpTest(t, "connect", "--help")
	if err != nil {
		t.Fatalf("expected nil error, got %v\nstdout=%q\nstderr=%q", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "sshcloak connect") {
		t.Fatalf("expected connect help output, got %q", stdout)
	}
}

func TestInitHelpFlagShowsHelp(t *testing.T) {
	stdout, stderr, err := executeRootForHelpTest(t, "init", "--help")
	if err != nil {
		t.Fatalf("expected nil error, got %v\nstdout=%q\nstderr=%q", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "sshcloak init") {
		t.Fatalf("expected init help output, got %q", stdout)
	}
}

func TestVersionHelpFlagShowsHelp(t *testing.T) {
	stdout, stderr, err := executeRootForHelpTest(t, "version", "--help")
	if err != nil {
		t.Fatalf("expected nil error, got %v\nstdout=%q\nstderr=%q", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "sshcloak version") {
		t.Fatalf("expected version help output, got %q", stdout)
	}
}
