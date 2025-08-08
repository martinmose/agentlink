package symlink

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSource(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name     string
		setup    func() string
		wantErr  bool
	}{
		{
			name: "valid regular file",
			setup: func() string {
				file := filepath.Join(tmpDir, "valid.md")
				os.WriteFile(file, []byte("test"), 0644)
				return file
			},
			wantErr: false,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "missing.md")
			},
			wantErr: true,
		},
		{
			name: "symlink source (should fail without force)",
			setup: func() string {
				target := filepath.Join(tmpDir, "target.md")
				link := filepath.Join(tmpDir, "link.md")
				os.WriteFile(target, []byte("test"), 0644)
				os.Symlink(target, link)
				return link
			},
			wantErr: true,
		},
		{
			name: "directory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "dir")
				os.Mkdir(dir, 0755)
				return dir
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(false, false, false)
			sourcePath := tt.setup()
			
			err := manager.ValidateSource(sourcePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckLink(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(false, false, false)
	
	// Create source file
	source := filepath.Join(tmpDir, "source.md")
	os.WriteFile(source, []byte("test"), 0644)
	
	tests := []struct {
		name           string
		setup          func() string
		expectedStatus LinkStatus
	}{
		{
			name: "missing link",
			setup: func() string {
				return filepath.Join(tmpDir, "missing.md")
			},
			expectedStatus: StatusMissing,
		},
		{
			name: "correct symlink",
			setup: func() string {
				link := filepath.Join(tmpDir, "correct.md")
				os.Symlink("source.md", link) // relative link
				return link
			},
			expectedStatus: StatusOK,
		},
		{
			name: "wrong target",
			setup: func() string {
				other := filepath.Join(tmpDir, "other.md")
				link := filepath.Join(tmpDir, "wrong.md")
				os.WriteFile(other, []byte("other"), 0644)
				os.Symlink("other.md", link)
				return link
			},
			expectedStatus: StatusWrongTarget,
		},
		{
			name: "not a symlink",
			setup: func() string {
				file := filepath.Join(tmpDir, "regular.md")
				os.WriteFile(file, []byte("content"), 0644)
				return file
			},
			expectedStatus: StatusNotSymlink,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkPath := tt.setup()
			
			info := manager.CheckLink(linkPath, source)
			if info.Status != tt.expectedStatus {
				t.Errorf("CheckLink() status = %v, expected %v", info.Status, tt.expectedStatus)
			}
		})
	}
}

func TestCreateLink(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(false, false, false)
	
	// Create source file
	source := filepath.Join(tmpDir, "source.md")
	os.WriteFile(source, []byte("test"), 0644)
	
	// Create link
	link := filepath.Join(tmpDir, "test.md")
	err := manager.CreateLink(link, source)
	if err != nil {
		t.Fatalf("CreateLink() failed: %v", err)
	}
	
	// Verify link exists and points to source
	info := manager.CheckLink(link, source)
	if info.Status != StatusOK {
		t.Errorf("Created link has wrong status: %v", info.Status)
	}
	
	// Verify we can read through the link
	content, err := os.ReadFile(link)
	if err != nil {
		t.Fatalf("Cannot read through link: %v", err)
	}
	if string(content) != "test" {
		t.Errorf("Wrong content through link: got %s, expected test", string(content))
	}
}

func TestFixLink(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(false, true, false) // force enabled
	
	// Create source file
	source := filepath.Join(tmpDir, "source.md")
	os.WriteFile(source, []byte("test"), 0644)
	
	tests := []struct {
		name           string
		setup          func() string
		expectedAction string
	}{
		{
			name: "create missing link",
			setup: func() string {
				return filepath.Join(tmpDir, "missing.md")
			},
			expectedAction: "create",
		},
		{
			name: "skip correct link",
			setup: func() string {
				link := filepath.Join(tmpDir, "correct.md")
				os.Symlink("source.md", link)
				return link
			},
			expectedAction: "skip",
		},
		{
			name: "fix wrong target",
			setup: func() string {
				other := filepath.Join(tmpDir, "other.md")
				link := filepath.Join(tmpDir, "wrong.md")
				os.WriteFile(other, []byte("other"), 0644)
				os.Symlink("other.md", link)
				return link
			},
			expectedAction: "fix",
		},
		{
			name: "replace regular file",
			setup: func() string {
				file := filepath.Join(tmpDir, "regular.md")
				os.WriteFile(file, []byte("content"), 0644)
				return file
			},
			expectedAction: "replace",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkPath := tt.setup()
			
			action, err := manager.FixLink(linkPath, source)
			if err != nil {
				t.Fatalf("FixLink() failed: %v", err)
			}
			
			if action != tt.expectedAction {
				t.Errorf("FixLink() action = %s, expected %s", action, tt.expectedAction)
			}
			
			// Verify the link is correct after fixing (except for skip case)
			if tt.expectedAction != "skip" {
				info := manager.CheckLink(linkPath, source)
				if info.Status != StatusOK {
					t.Errorf("Link not correct after fixing: status = %v", info.Status)
				}
			}
		})
	}
}

func TestDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(true, false, false) // dry-run enabled
	
	source := filepath.Join(tmpDir, "source.md")
	os.WriteFile(source, []byte("test"), 0644)
	
	link := filepath.Join(tmpDir, "test.md")
	
	// Create link in dry-run mode
	err := manager.CreateLink(link, source)
	if err != nil {
		t.Fatalf("CreateLink() in dry-run failed: %v", err)
	}
	
	// Link should not exist
	if _, err := os.Lstat(link); err == nil {
		t.Error("Link was created in dry-run mode")
	}
}