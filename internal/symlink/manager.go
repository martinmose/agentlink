package symlink

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkStatus represents the status of a symlink
type LinkStatus int

const (
	StatusOK LinkStatus = iota
	StatusMissing
	StatusWrongTarget
	StatusNotSymlink
	StatusBroken
)

func (s LinkStatus) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusMissing:
		return "missing"
	case StatusWrongTarget:
		return "wrong target"
	case StatusNotSymlink:
		return "not a symlink"
	case StatusBroken:
		return "broken"
	default:
		return "unknown"
	}
}

// LinkInfo contains information about a symlink
type LinkInfo struct {
	Path         string
	Target       string
	ExpectedPath string
	Status       LinkStatus
	Error        error
}

// Manager handles symlink operations
type Manager struct {
	dryRun  bool
	force   bool
	verbose bool
}

// NewManager creates a new symlink manager
func NewManager(dryRun, force, verbose bool) *Manager {
	return &Manager{
		dryRun:  dryRun,
		force:   force,
		verbose: verbose,
	}
}

// ValidateSource checks if the source file exists and is a regular file
func (m *Manager) ValidateSource(sourcePath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source file %s does not exist", sourcePath)
		}
		return fmt.Errorf("failed to stat source file %s: %w", sourcePath, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		if !m.force {
			return fmt.Errorf("source file %s is a symlink (use --force to override)", sourcePath)
		}
	}

	if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("source file %s is not a regular file", sourcePath)
	}

	return nil
}

// CheckLink checks the status of a symlink
func (m *Manager) CheckLink(linkPath, expectedTarget string) *LinkInfo {
	info := &LinkInfo{
		Path:         linkPath,
		ExpectedPath: expectedTarget,
	}

	// Check if link exists
	linkInfo, err := os.Lstat(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			info.Status = StatusMissing
			return info
		}
		info.Error = err
		info.Status = StatusBroken
		return info
	}

	// Check if it's a symlink
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		info.Status = StatusNotSymlink
		return info
	}

	// Get the target
	target, err := os.Readlink(linkPath)
	if err != nil {
		info.Error = err
		info.Status = StatusBroken
		return info
	}

	info.Target = target

	// Make target path absolute for comparison
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(linkPath), target)
	}
	target = filepath.Clean(target)
	expectedClean := filepath.Clean(expectedTarget)

	if target == expectedClean {
		info.Status = StatusOK
	} else {
		info.Status = StatusWrongTarget
	}

	return info
}

// CreateLink creates or fixes a symlink
func (m *Manager) CreateLink(linkPath, targetPath string) error {
	if m.dryRun {
		return nil // Don't actually create in dry-run mode
	}

	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", linkPath, err)
	}

	// Calculate relative path from link to target
	relTarget, err := filepath.Rel(filepath.Dir(linkPath), targetPath)
	if err != nil {
		return fmt.Errorf("failed to calculate relative path: %w", err)
	}

	// Create the symlink
	if err := os.Symlink(relTarget, linkPath); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", linkPath, relTarget, err)
	}

	return nil
}

// RemoveLink removes a symlink if it's managed by agentlink
func (m *Manager) RemoveLink(linkPath, expectedTarget string) error {
	if m.dryRun {
		return nil
	}

	info := m.CheckLink(linkPath, expectedTarget)
	if info.Status == StatusOK {
		if err := os.Remove(linkPath); err != nil {
			return fmt.Errorf("failed to remove symlink %s: %w", linkPath, err)
		}
	}

	return nil
}

// FixLink creates or fixes a symlink based on its current status
func (m *Manager) FixLink(linkPath, targetPath string) (string, error) {
	info := m.CheckLink(linkPath, targetPath)

	switch info.Status {
	case StatusOK:
		return "skip", nil

	case StatusMissing:
		if err := m.CreateLink(linkPath, targetPath); err != nil {
			return "", err
		}
		return "create", nil

	case StatusWrongTarget:
		if !m.force {
			return "", fmt.Errorf("symlink %s points to wrong target %s (expected %s), use --force to fix", linkPath, info.Target, targetPath)
		}
		if err := os.Remove(linkPath); err != nil {
			return "", fmt.Errorf("failed to remove wrong symlink %s: %w", linkPath, err)
		}
		if err := m.CreateLink(linkPath, targetPath); err != nil {
			return "", err
		}
		return "fix", nil

	case StatusNotSymlink:
		if !m.force {
			return "", fmt.Errorf("file %s exists and is not a symlink, use --force to replace", linkPath)
		}
		if err := os.RemoveAll(linkPath); err != nil {
			return "", fmt.Errorf("failed to remove existing file %s: %w", linkPath, err)
		}
		if err := m.CreateLink(linkPath, targetPath); err != nil {
			return "", err
		}
		return "replace", nil

	case StatusBroken:
		if err := os.Remove(linkPath); err != nil {
			return "", fmt.Errorf("failed to remove broken symlink %s: %w", linkPath, err)
		}
		if err := m.CreateLink(linkPath, targetPath); err != nil {
			return "", err
		}
		return "fix broken", nil

	default:
		return "", fmt.Errorf("unknown link status for %s", linkPath)
	}
}