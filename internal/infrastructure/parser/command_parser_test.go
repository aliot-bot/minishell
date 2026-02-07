package parser

import (
	"minishell/internal/domain/entity"
	"minishell/internal/infrastructure/env"
	"testing"
)

func TestCommandParser_Parse(t *testing.T) {
	expander := env.NewExpander()
	parser := NewCommandParser(expander)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, interface{})
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				if result != nil {
					stmt := result.(*entity.Statement)
					if stmt != nil {
						t.Errorf("expected nil for empty input")
					}
				}
			},
		},
		{
			name:    "simple command",
			input:   "echo hello",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if len(stmt.Pipeline.Commands) != 1 {
					t.Errorf("expected 1 command, got %d", len(stmt.Pipeline.Commands))
				}
				if stmt.Pipeline.Commands[0].Name != "echo" {
					t.Errorf("expected command name 'echo', got '%s'", stmt.Pipeline.Commands[0].Name)
				}
			},
		},
		{
			name:    "pipeline",
			input:   "echo hello | grep hello",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if len(stmt.Pipeline.Commands) != 2 {
					t.Errorf("expected 2 commands, got %d", len(stmt.Pipeline.Commands))
				}
			},
		},
		{
			name:    "conditional and",
			input:   "echo first && echo second",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if stmt.NextOp != "&&" {
					t.Errorf("expected && operator, got '%s'", stmt.NextOp)
				}
				if stmt.NextStmt == nil {
					t.Error("expected NextStmt to be set")
				}
			},
		},
		{
			name:    "conditional or",
			input:   "false || echo fallback",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if stmt.NextOp != "||" {
					t.Errorf("expected || operator, got '%s'", stmt.NextOp)
				}
			},
		},
		{
			name:    "input redirection",
			input:   "cat < file.txt",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if stmt.Pipeline.Commands[0].InputFile != "file.txt" {
					t.Errorf("expected InputFile 'file.txt', got '%s'", stmt.Pipeline.Commands[0].InputFile)
				}
			},
		},
		{
			name:    "output redirection",
			input:   "echo hello > output.txt",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if stmt.Pipeline.Commands[0].OutputFile != "output.txt" {
					t.Errorf("expected OutputFile 'output.txt', got '%s'", stmt.Pipeline.Commands[0].OutputFile)
				}
			},
		},
		{
			name:    "quoted arguments",
			input:   `echo "hello world"`,
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if len(stmt.Pipeline.Commands[0].Args) != 1 {
					t.Errorf("expected 1 arg, got %d", len(stmt.Pipeline.Commands[0].Args))
				}
			},
		},
		{
			name:    "environment variable",
			input:   "echo $HOME",
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				stmt := result.(*entity.Statement)
				if len(stmt.Pipeline.Commands[0].Args) == 0 {
					t.Error("expected args to be expanded")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestCommandParser_ParseComplex(t *testing.T) {
	expander := env.NewExpander()
	parser := NewCommandParser(expander)

	complex := "echo 'test' | grep test && cat file.txt || echo 'fallback'"
	stmt, err := parser.Parse(complex)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if stmt == nil {
		t.Fatal("expected statement, got nil")
	}

	if len(stmt.Pipeline.Commands) != 2 {
		t.Errorf("expected 2 commands in first pipeline, got %d", len(stmt.Pipeline.Commands))
	}

	if stmt.NextOp != "&&" {
		t.Errorf("expected && operator, got '%s'", stmt.NextOp)
	}

	if stmt.NextStmt == nil {
		t.Fatal("expected NextStmt to be set")
	}

	if stmt.NextStmt.NextOp != "||" {
		t.Errorf("expected || operator in NextStmt, got '%s'", stmt.NextStmt.NextOp)
	}
}

