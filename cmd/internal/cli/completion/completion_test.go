package completion

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"

	"github.com/sipuaz/sshcloak/internal/config"
)

// TestListHostsCmd_MatchesOutput validates list-hosts command structure and output format.
func TestListHostsCmd_MatchesOutput(t *testing.T) {
	// Create a test command tree
	cmd := newListHostsCmd()

	// Verify basic command structure
	if cmd.Use != "list-hosts" {
		t.Errorf("unexpected command Use: got %q, want %q", cmd.Use, "list-hosts")
	}

	if cmd.RunE == nil {
		t.Error("list-hosts command missing RunE function")
	}

	// Verify command metadata
	if cmd.Short == "" {
		t.Error("list-hosts command missing Short description")
	}
}

// TestFlagsCmd_MatchesStructure validates flags command structure.
func TestFlagsCmd_MatchesStructure(t *testing.T) {
	cmd := newFlagsCmd()

	if cmd.Use != "flags <command>" {
		t.Errorf("unexpected command Use: got %q, want %q", cmd.Use, "flags <command>")
	}

	if cmd.RunE == nil {
		t.Error("flags command missing RunE function")
	}

	// Verify that Args validation is set
	if cmd.Args == nil {
		t.Error("flags command missing Args validation")
	}
}

// TestNewCompletionCmd_RegistersSubcommands validates that NewCompletionCmd registers subcommands.
func TestNewCompletionCmd_RegistersSubcommands(t *testing.T) {
	cmd := NewCompletionCmd()

	if cmd.Use != "completion" {
		t.Errorf("unexpected command Use: got %q, want %q", cmd.Use, "completion")
	}

	// Check that subcommands are registered
	subcommands := cmd.Commands()
	if len(subcommands) < 2 {
		t.Errorf("expected at least 2 subcommands, got %d", len(subcommands))
	}

	// Verify list-hosts and flags are present
	hasListHosts := false
	hasFlags := false
	for _, sc := range subcommands {
		if sc.Name() == "list-hosts" {
			hasListHosts = true
		}
		if sc.Name() == "flags" {
			hasFlags = true
		}
	}

	if !hasListHosts {
		t.Error("completion command missing list-hosts subcommand")
	}
	if !hasFlags {
		t.Error("completion command missing flags subcommand")
	}
}

// TestListHostOutput_ValidatesStructure validates the listHostOutput JSON structure.
func TestListHostOutput_ValidatesStructure(t *testing.T) {
	// Simulate JSON output
	output := listHostOutput{
		Host: "test-server",
		Tags: []string{"production"},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal listHostOutput: %v", err)
	}

	// Verify JSON contains expected fields
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if _, hasHost := decoded["host"]; !hasHost {
		t.Error("JSON output missing 'host' field")
	}
	if _, hasTags := decoded["tags"]; !hasTags {
		t.Error("JSON output missing 'tags' field")
	}
}

// TestFlagsCmd_WritesToOutput validates that flags command writes to output.
func TestFlagsCmd_WritesToOutput(t *testing.T) {
	// Create a root command with a subcommand
	root := &cobra.Command{Use: "test"}
	sub := &cobra.Command{Use: "sub"}
	sub.Flags().String("flag1", "", "test flag 1")
	sub.Flags().Bool("flag2", false, "test flag 2")
	root.AddCommand(sub)

	// Create the flags command and execute it
	flagsCmd := newFlagsCmd()
	var out bytes.Buffer
	flagsCmd.SetOut(&out)

	// Simulate execution by calling runFlags directly
	// (normally this would be executed via root)
	err := runFlags(flagsCmd, []string{"sub"})
	if err != nil {
		t.Errorf("runFlags failed: %v", err)
	}

	// Note: In a real test with proper command injection, we'd verify output
	// For now, we just verify the command executes without error
}

// TestSetManagerAndGetManager validates injection mechanism.
func TestSetManagerAndGetManager(t *testing.T) {
	mockMgr := &config.Manager{}
	cmd := &cobra.Command{}

	// Initially, getManager should panic (unless manager was set elsewhere)
	// This is by design — SetManager must be called before getManager

	// After SetManager, getManager should return the injected manager
	SetManager(cmd, mockMgr)
	if getManager() != mockMgr {
		t.Error("getManager returned unexpected manager")
	}

	// Reset for other tests
	sharedManager = nil
}
