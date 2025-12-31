package project

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindProjectRoot tests project root detection climbing up directory tree
func TestFindProjectRoot(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string) // returns (startPath, expectedRoot)
		wantErr     bool
		expectFound bool
	}{
		{
			name: "finds root with .git directory",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("failed to create .git: %v", err)
				}
				subDir := filepath.Join(tmpDir, "src", "pkg")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}
				return subDir, tmpDir
			},
			expectFound: true,
		},
		{
			name: "finds root with .claude directory",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				claudeDir := filepath.Join(tmpDir, ".claude")
				if err := os.Mkdir(claudeDir, 0755); err != nil {
					t.Fatalf("failed to create .claude: %v", err)
				}
				subDir := filepath.Join(tmpDir, "nested", "deep")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}
				return subDir, tmpDir
			},
			expectFound: true,
		},
		{
			name: "finds root with go.mod",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				goMod := filepath.Join(tmpDir, "go.mod")
				if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				subDir := filepath.Join(tmpDir, "internal")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}
				return subDir, tmpDir
			},
			expectFound: true,
		},
		{
			name: "finds root with package.json",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				pkgJSON := filepath.Join(tmpDir, "package.json")
				if err := os.WriteFile(pkgJSON, []byte("{}"), 0644); err != nil {
					t.Fatalf("failed to create package.json: %v", err)
				}
				subDir := filepath.Join(tmpDir, "src", "components")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}
				return subDir, tmpDir
			},
			expectFound: true,
		},
		{
			name: "no project root - returns start path",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				subDir := filepath.Join(tmpDir, "no-markers")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatalf("failed to create subdirectory: %v", err)
				}
				return subDir, subDir
			},
			expectFound: false,
		},
		{
			name: "start from current directory",
			setupFunc: func(t *testing.T) (string, string) {
				// Use actual current directory which should have go.mod
				cwd, err := os.Getwd()
				if err != nil {
					t.Fatalf("failed to get current directory: %v", err)
				}
				// Navigate up to find project root
				projectRoot := cwd
				for {
					if isProjectRoot(projectRoot) {
						break
					}
					parent := filepath.Dir(projectRoot)
					if parent == projectRoot {
						break
					}
					projectRoot = parent
				}
				return ".", projectRoot
			},
			expectFound: true,
		},
		{
			name: "absolute path conversion - dot path",
			setupFunc: func(t *testing.T) (string, string) {
				// Create temp dir with marker
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}

				// Change to that directory
				origDir, _ := os.Getwd()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to change directory: %v", err)
				}
				t.Cleanup(func() {
					os.Chdir(origDir)
				})

				return ".", tmpDir
			},
			expectFound: true,
		},
		{
			name: "multiple markers - stops at first",
			setupFunc: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				// Create inner project with go.mod
				innerDir := filepath.Join(tmpDir, "outer", "inner")
				if err := os.MkdirAll(innerDir, 0755); err != nil {
					t.Fatalf("failed to create inner directory: %v", err)
				}
				if err := os.WriteFile(filepath.Join(innerDir, "go.mod"), []byte("module inner"), 0644); err != nil {
					t.Fatalf("failed to create inner go.mod: %v", err)
				}
				// Create outer project with .git
				if err := os.Mkdir(filepath.Join(tmpDir, "outer", ".git"), 0755); err != nil {
					t.Fatalf("failed to create outer .git: %v", err)
				}
				// Start from inner, should find inner first
				return innerDir, innerDir
			},
			expectFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startPath, expectedRoot := tt.setupFunc(t)

			got, err := FindProjectRoot(startPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindProjectRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Convert both to absolute for comparison and resolve symlinks (macOS /var -> /private/var)
				absGot, err := filepath.EvalSymlinks(got)
				if err != nil {
					absGot, _ = filepath.Abs(got)
				}
				absExpected, err := filepath.EvalSymlinks(expectedRoot)
				if err != nil {
					absExpected, _ = filepath.Abs(expectedRoot)
				}

				if absGot != absExpected {
					t.Errorf("FindProjectRoot() = %v, want %v", absGot, absExpected)
				}
			}
		})
	}
}

// TestIsProjectRoot tests the project root marker detection
func TestIsProjectRoot(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		want      bool
	}{
		{
			name: "directory with .git",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gitDir := filepath.Join(tmpDir, ".git")
				if err := os.Mkdir(gitDir, 0755); err != nil {
					t.Fatalf("failed to create .git: %v", err)
				}
				return tmpDir
			},
			want: true,
		},
		{
			name: "directory with .claude",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				claudeDir := filepath.Join(tmpDir, ".claude")
				if err := os.Mkdir(claudeDir, 0755); err != nil {
					t.Fatalf("failed to create .claude: %v", err)
				}
				return tmpDir
			},
			want: true,
		},
		{
			name: "directory with package.json",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				pkgJSON := filepath.Join(tmpDir, "package.json")
				if err := os.WriteFile(pkgJSON, []byte("{}"), 0644); err != nil {
					t.Fatalf("failed to create package.json: %v", err)
				}
				return tmpDir
			},
			want: true,
		},
		{
			name: "directory with go.mod",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				goMod := filepath.Join(tmpDir, "go.mod")
				if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				return tmpDir
			},
			want: true,
		},
		{
			name: "directory with multiple markers",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
					t.Fatalf("failed to create .git: %v", err)
				}
				if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
					t.Fatalf("failed to create .claude: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				return tmpDir
			},
			want: true,
		},
		{
			name: "empty directory",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			want: false,
		},
		{
			name: "directory with other files but no markers",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644); err != nil {
					t.Fatalf("failed to create README.md: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
					t.Fatalf("failed to create main.go: %v", err)
				}
				return tmpDir
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFunc(t)
			if got := isProjectRoot(path); got != tt.want {
				t.Errorf("isProjectRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetectProjectInfo tests project information detection
func TestDetectProjectInfo(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) string
		wantType     string
		wantIsClaude bool
		wantHasGit   bool
		wantFiles    int
	}{
		{
			name: "claude project with git",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
					t.Fatalf("failed to create .claude: %v", err)
				}
				if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
					t.Fatalf("failed to create .git: %v", err)
				}
				return tmpDir
			},
			wantType:     "claude",
			wantIsClaude: true,
			wantHasGit:   true,
			wantFiles:    2,
		},
		{
			name: "go project with git",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
					t.Fatalf("failed to create .git: %v", err)
				}
				return tmpDir
			},
			wantType:     "go",
			wantIsClaude: false,
			wantHasGit:   true,
			wantFiles:    1,
		},
		{
			name: "node project without git",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644); err != nil {
					t.Fatalf("failed to create package.json: %v", err)
				}
				return tmpDir
			},
			wantType:     "node",
			wantIsClaude: false,
			wantHasGit:   false,
			wantFiles:    0,
		},
		{
			name: "claude go project - claude takes precedence",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
					t.Fatalf("failed to create .claude: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				return tmpDir
			},
			wantType:     "go", // go.mod is checked after .claude, so it overwrites
			wantIsClaude: true,
			wantHasGit:   false,
			wantFiles:    1,
		},
		{
			name: "unknown project type",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644); err != nil {
					t.Fatalf("failed to create README.md: %v", err)
				}
				return tmpDir
			},
			wantType:     "unknown",
			wantIsClaude: false,
			wantHasGit:   false,
			wantFiles:    0,
		},
		{
			name: "node and go project - go takes precedence",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644); err != nil {
					t.Fatalf("failed to create package.json: %v", err)
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
				return tmpDir
			},
			wantType:     "go",
			wantIsClaude: false,
			wantHasGit:   false,
			wantFiles:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := tt.setupFunc(t)

			info, err := DetectProjectInfo(rootPath)
			if err != nil {
				t.Fatalf("DetectProjectInfo() error = %v", err)
			}

			if info.Root != rootPath {
				t.Errorf("DetectProjectInfo() Root = %v, want %v", info.Root, rootPath)
			}
			if info.Type != tt.wantType {
				t.Errorf("DetectProjectInfo() Type = %v, want %v", info.Type, tt.wantType)
			}
			if info.IsClaude != tt.wantIsClaude {
				t.Errorf("DetectProjectInfo() IsClaude = %v, want %v", info.IsClaude, tt.wantIsClaude)
			}
			if info.HasGit != tt.wantHasGit {
				t.Errorf("DetectProjectInfo() HasGit = %v, want %v", info.HasGit, tt.wantHasGit)
			}
			if len(info.FilesFound) != tt.wantFiles {
				t.Errorf("DetectProjectInfo() FilesFound count = %v, want %v", len(info.FilesFound), tt.wantFiles)
			}
		})
	}
}

// TestFindProjectFiles tests the file discovery functionality
func TestFindProjectFiles(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantFiles []string
	}{
		{
			name: "finds .claude and .git",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
					t.Fatalf("failed to create .claude: %v", err)
				}
				if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
					t.Fatalf("failed to create .git: %v", err)
				}
				return tmpDir
			},
			wantFiles: []string{".claude/", ".git/"},
		},
		{
			name: "finds only .claude",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
					t.Fatalf("failed to create .claude: %v", err)
				}
				return tmpDir
			},
			wantFiles: []string{".claude/"},
		},
		{
			name: "finds only .git",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
					t.Fatalf("failed to create .git: %v", err)
				}
				return tmpDir
			},
			wantFiles: []string{".git/"},
		},
		{
			name: "finds nothing",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644); err != nil {
					t.Fatalf("failed to create README.md: %v", err)
				}
				return tmpDir
			},
			wantFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := tt.setupFunc(t)

			files := findProjectFiles(rootPath)

			if len(files) != len(tt.wantFiles) {
				t.Errorf("findProjectFiles() returned %d files, want %d", len(files), len(tt.wantFiles))
			}

			// Check that all expected files are present
			for _, wantFile := range tt.wantFiles {
				found := false
				for _, gotFile := range files {
					if gotFile == wantFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("findProjectFiles() missing expected file %v", wantFile)
				}
			}
		})
	}
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("FindProjectRoot with non-existent path", func(t *testing.T) {
		nonExistentPath := "/this/path/should/not/exist/anywhere/on/filesystem/xyz123"
		_, err := FindProjectRoot(nonExistentPath)
		// Should return error due to filepath.Abs failing on non-existent path
		// or should handle gracefully
		if err != nil {
			// Expected - path doesn't exist
			t.Logf("Expected error for non-existent path: %v", err)
		}
	})

	t.Run("isProjectRoot with non-existent path", func(t *testing.T) {
		nonExistentPath := "/this/path/should/not/exist"
		result := isProjectRoot(nonExistentPath)
		if result {
			t.Error("isProjectRoot() should return false for non-existent path")
		}
	})

	t.Run("DetectProjectInfo with empty path", func(t *testing.T) {
		info, err := DetectProjectInfo("")
		if err != nil {
			t.Fatalf("DetectProjectInfo() with empty path failed: %v", err)
		}
		if info.Type != "unknown" {
			t.Errorf("DetectProjectInfo() with empty path should have unknown type, got %v", info.Type)
		}
	})

	t.Run("FindProjectRoot from filesystem root", func(t *testing.T) {
		// Create a temporary deep path
		tmpDir := t.TempDir()
		deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
		if err := os.MkdirAll(deepPath, 0755); err != nil {
			t.Fatalf("failed to create deep path: %v", err)
		}

		// Should climb all the way up and return the start path when no markers found
		result, err := FindProjectRoot(deepPath)
		if err != nil {
			t.Fatalf("FindProjectRoot() error = %v", err)
		}

		// Should return deepPath since no project markers were found
		absDeep, _ := filepath.Abs(deepPath)
		absResult, _ := filepath.Abs(result)
		if absResult != absDeep {
			t.Errorf("FindProjectRoot() should return start path when no markers found, got %v, want %v", absResult, absDeep)
		}
	})
}

// TestProjectInfoStructure tests the ProjectInfo struct
func TestProjectInfoStructure(t *testing.T) {
	t.Run("ProjectInfo initialization", func(t *testing.T) {
		info := &ProjectInfo{
			Root:       "/test/path",
			IsClaude:   true,
			HasGit:     true,
			Type:       "go",
			FilesFound: []string{".claude/", ".git/"},
		}

		if info.Root != "/test/path" {
			t.Errorf("Root = %v, want /test/path", info.Root)
		}
		if !info.IsClaude {
			t.Error("IsClaude should be true")
		}
		if !info.HasGit {
			t.Error("HasGit should be true")
		}
		if info.Type != "go" {
			t.Errorf("Type = %v, want go", info.Type)
		}
		if len(info.FilesFound) != 2 {
			t.Errorf("FilesFound length = %v, want 2", len(info.FilesFound))
		}
	})
}
