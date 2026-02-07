package executor

import (
	"minishell/internal/domain/entity"
	"minishell/internal/infrastructure/builtins"
	"os"
	"os/exec"
	"testing"
)

func TestCommandExecutor_ExecuteStatement(t *testing.T) {
	builtinService := builtins.NewBuiltinService()
	executor := NewCommandExecutor(builtinService, builtinService)

	tests := []struct {
		name    string
		stmt    *entity.Statement
		wantErr bool
	}{
		{
			name:    "nil statement",
			stmt:    nil,
			wantErr: false,
		},
		{
			name: "simple command",
			stmt: &entity.Statement{
				Pipeline: entity.Pipeline{
					Commands: []entity.Command{
						{Name: "echo", Args: []string{"test"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "conditional and success",
			stmt: &entity.Statement{
				Pipeline: entity.Pipeline{
					Commands: []entity.Command{
						{Name: "echo", Args: []string{"first"}},
					},
				},
				NextOp: "&&",
				NextStmt: &entity.Statement{
					Pipeline: entity.Pipeline{
						Commands: []entity.Command{
							{Name: "echo", Args: []string{"second"}},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ExecuteStatement(tt.stmt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteStatement() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandExecutor_ExecutePipeline(t *testing.T) {
	builtinService := builtins.NewBuiltinService()
	executor := NewCommandExecutor(builtinService, builtinService)

	if _, err := exec.LookPath("echo"); err != nil {
		t.Skip("echo command not available")
	}

	tests := []struct {
		name    string
		pipeline entity.Pipeline
		wantErr bool
	}{
		{
			name:    "empty pipeline",
			pipeline: entity.Pipeline{Commands: []entity.Command{}},
			wantErr: false,
		},
		{
			name: "single command",
			pipeline: entity.Pipeline{
				Commands: []entity.Command{
					{Name: "echo", Args: []string{"test"}},
				},
			},
			wantErr: false,
		},
		{
			name: "pipeline with two commands",
			pipeline: entity.Pipeline{
				Commands: []entity.Command{
					{Name: "echo", Args: []string{"hello"}},
					{Name: "echo", Args: []string{"world"}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ExecutePipeline(tt.pipeline)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecutePipeline() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandExecutor_ExecuteCommandWithRedirection(t *testing.T) {
	builtinService := builtins.NewBuiltinService()
	executor := NewCommandExecutor(builtinService, builtinService)

	tmpFile := "/tmp/minishell_test_output.txt"
	defer os.Remove(tmpFile)

	cmd := entity.Command{
		Name:       "echo",
		Args:       []string{"test output"},
		OutputFile: tmpFile,
	}

	err := executor.ExecuteCommand(cmd, nil, nil)
	if err != nil {
		t.Fatalf("ExecuteCommand() error = %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(content) != "test output\n" {
		t.Errorf("expected 'test output\\n', got %q", string(content))
	}
}

func TestCommandExecutor_ExternalCommand(t *testing.T) {
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("true command not available")
	}

	builtinService := builtins.NewBuiltinService()
	executor := NewCommandExecutor(builtinService, builtinService)

	cmd := entity.Command{
		Name: "true",
		Args: []string{},
	}

	err := executor.ExecuteCommand(cmd, nil, nil)
	if err != nil {
		t.Errorf("ExecuteCommand() error = %v", err)
	}
}

func TestCommandExecutor_ExternalCommandNotFound(t *testing.T) {
	builtinService := builtins.NewBuiltinService()
	executor := NewCommandExecutor(builtinService, builtinService)

	cmd := entity.Command{
		Name: "nonexistent_command_12345",
		Args: []string{},
	}

	err := executor.ExecuteCommand(cmd, nil, nil)
	if err == nil {
		t.Error("expected error for nonexistent command")
	}
}

