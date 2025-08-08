package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegrationBasicWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	
	// Change to test directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	
	// Create a fake .git directory
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatal(err)
	}
	
	// Build the binary
	binaryPath := filepath.Join(tmpDir, "agentlink")
	cmd := exec.Command("go", "build", "-o", binaryPath, filepath.Join(origDir, "cmd", "agentlink"))
	cmd.Dir = origDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	
	// Test 1: Init command
	cmd = exec.Command(binaryPath, "init")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "Created") {
		t.Errorf("Init output doesn't contain 'Created': %s", output)
	}
	
	// Verify config file was created
	if _, err := os.Stat(".agentlink.yaml"); err != nil {
		t.Error("Config file was not created")
	}
	
	// Test 2: Check command (should show problems since source doesn't exist)
	cmd = exec.Command(binaryPath, "check")
	output, err = cmd.CombinedOutput()
	if err == nil {
		t.Error("Check should have failed since source doesn't exist")
	}
	
	// Test 3: Create source file and sync
	sourceContent := "# Test Source\nThis is a test instruction file."
	if err := os.WriteFile("CLAUDE.md", []byte(sourceContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	cmd = exec.Command(binaryPath, "sync")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Sync failed: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "[create]") {
		t.Errorf("Sync output doesn't contain '[create]': %s", output)
	}
	
	// Test 4: Verify symlinks were created
	for _, link := range []string{"AGENTS.md", "OPENCODE.md"} {
		info, err := os.Lstat(link)
		if err != nil {
			t.Errorf("Link %s was not created: %v", link, err)
			continue
		}
		
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("File %s is not a symlink", link)
		}
		
		// Verify content is accessible through symlink
		content, err := os.ReadFile(link)
		if err != nil {
			t.Errorf("Cannot read through symlink %s: %v", link, err)
			continue
		}
		
		if string(content) != sourceContent {
			t.Errorf("Content through symlink %s doesn't match source", link)
		}
	}
	
	// Test 5: Check command (should pass now)
	cmd = exec.Command(binaryPath, "check")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Check failed after sync: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "All links are correctly configured") {
		t.Errorf("Check output doesn't indicate success: %s", output)
	}
	
	// Test 6: Sync again (should be idempotent)
	cmd = exec.Command(binaryPath, "sync")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Second sync failed: %v\nOutput: %s", err, output)
	}
	
	// Should not contain [create] this time
	if strings.Contains(string(output), "[create]") {
		t.Errorf("Second sync should not create new links: %s", output)
	}
	
	// Test 7: Clean command
	cmd = exec.Command(binaryPath, "clean", "--dry-run")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Clean dry-run failed: %v\nOutput: %s", err, output)
	}
	
	if !strings.Contains(string(output), "would remove") {
		t.Errorf("Clean dry-run output doesn't mention removal: %s", output)
	}
}

func TestIntegrationDoctorCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "agentlink")
	cmd := exec.Command("go", "build", "-o", binaryPath, "cmd/agentlink")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	
	// Run doctor command
	cmd = exec.Command(binaryPath, "doctor")
	output, _ := cmd.CombinedOutput()
	// Doctor might return non-zero exit code if there are warnings, but that's OK
	
	expectedStrings := []string{
		"Agentlink Doctor",
		"Operating System:",
		"Symlink Support:",
		"Binary Location:",
		"Configuration:",
		"Project Configuration:",
		"Global Configuration:",
	}
	
	outputStr := string(output)
	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Doctor output missing '%s':\n%s", expected, outputStr)
		}
	}
}

func TestIntegrationForceFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	
	// Build the binary
	binaryPath := filepath.Join(tmpDir, "agentlink")
	cmd := exec.Command("go", "build", "-o", binaryPath, filepath.Join(origDir, "cmd", "agentlink"))
	cmd.Dir = origDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	
	// Create .git and initialize project
	os.Mkdir(".git", 0755)
	
	cmd = exec.Command(binaryPath, "init")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	
	// Create source file
	if err := os.WriteFile("CLAUDE.md", []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create conflicting file
	if err := os.WriteFile("AGENTS.md", []byte("conflicting content"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Sync without force should fail
	cmd = exec.Command(binaryPath, "sync")
	if err := cmd.Run(); err == nil {
		t.Error("Sync should have failed due to conflicting file")
	}
	
	// Sync with force should succeed
	cmd = exec.Command(binaryPath, "sync", "--force")
	if err := cmd.Run(); err != nil {
		t.Errorf("Sync with --force failed: %v", err)
	}
	
	// Verify the file was replaced with a symlink
	info, err := os.Lstat("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("File was not replaced with symlink")
	}
}