package builtins

import (
	"os"
	"os/exec"
	"testing"
)

func TestBuiltinService_IsBuiltin(t *testing.T) {
	service := NewBuiltinService()

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"cd is builtin", "cd", true},
		{"pwd is builtin", "pwd", true},
		{"echo is builtin", "echo", true},
		{"kill is builtin", "kill", true},
		{"ps is builtin", "ps", true},
		{"ls is not builtin", "ls", false},
		{"grep is not builtin", "grep", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsBuiltin(tt.command)
			if result != tt.expected {
				t.Errorf("IsBuiltin(%s) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestBuiltinService_Echo(t *testing.T) {
	service := NewBuiltinService()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"echo with args", []string{"hello", "world"}, false},
		{"echo empty", []string{}, false},
		{"echo single", []string{"test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ExecuteBuiltin("echo", tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBuiltin(echo) error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuiltinService_Pwd(t *testing.T) {
	service := NewBuiltinService()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	err = service.ExecuteBuiltin("pwd", []string{})
	if err != nil {
		t.Errorf("ExecuteBuiltin(pwd) error = %v", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if currentDir != originalDir {
		t.Errorf("pwd should not change directory")
	}
}

func TestBuiltinService_Cd(t *testing.T) {
	service := NewBuiltinService()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		t.Skip("HOME not set, skipping cd test")
	}

	err = service.ExecuteBuiltin("cd", []string{homeDir})
	if err != nil {
		t.Errorf("ExecuteBuiltin(cd) error = %v", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if currentDir != homeDir {
		t.Errorf("cd failed: expected %s, got %s", homeDir, currentDir)
	}
}

func TestBuiltinService_CdInvalid(t *testing.T) {
	service := NewBuiltinService()

	err := service.ExecuteBuiltin("cd", []string{"/nonexistent/directory/12345"})
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}

func TestBuiltinService_KillInvalid(t *testing.T) {
	service := NewBuiltinService()

	tests := []struct {
		name string
		args []string
	}{
		{"kill without args", []string{}},
		{"kill with invalid pid", []string{"invalid"}},
		{"kill with nonexistent pid", []string{"999999"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ExecuteBuiltin("kill", tt.args)
			if err == nil && len(tt.args) > 0 {
				t.Error("expected error for invalid kill command")
			}
		})
	}
}

func TestBuiltinService_UnknownBuiltin(t *testing.T) {
	service := NewBuiltinService()

	err := service.ExecuteBuiltin("unknown", []string{})
	if err == nil {
		t.Error("expected error for unknown builtin")
	}
}

func TestBuiltinService_Ps(t *testing.T) {
	if _, err := exec.LookPath("ps"); err != nil {
		t.Skip("ps command not available")
	}

	service := NewBuiltinService()

	err := service.ExecuteBuiltin("ps", []string{})
	if err != nil {
		t.Errorf("ExecuteBuiltin(ps) error = %v", err)
	}
}

