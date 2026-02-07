package main

import (
	"fmt"
	"minishell/internal/domain/usecase"
	"minishell/internal/infrastructure/builtins"
	"minishell/internal/infrastructure/env"
	"minishell/internal/infrastructure/executor"
	"minishell/internal/infrastructure/parser"
	"minishell/internal/infrastructure/reader"
	"os"
	"os/exec"
)

type Shell struct {
	reader   usecase.InputReader
	parser   usecase.CommandParser
	executor usecase.CommandExecutor
}

func NewShell(reader usecase.InputReader, parser usecase.CommandParser, executor usecase.CommandExecutor) *Shell {
	return &Shell{
		reader:   reader,
		parser:   parser,
		executor: executor,
	}
}

func (s *Shell) Run() {
	for {
		line, ok := s.reader.ReadCommand()
		if !ok {
			fmt.Println("exit")
			os.Exit(0)
		}

		if line == "" {
			continue
		}

		if line == "exit" {
			os.Exit(0)
		}

		stmt, err := s.parser.Parse(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			continue
		}

		if stmt == nil {
			continue
		}

		if err := s.executor.ExecuteStatement(stmt); err != nil {
			if exitErr, ok := err.(*exec.ExitError); !ok || (exitErr.ExitCode() != 1 && exitErr.ExitCode() != 2) {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		}
	}
}

func main() {
	envExpander := env.NewExpander()
	commandParser := parser.NewCommandParser(envExpander)
	builtinService := builtins.NewBuiltinService()
	commandExecutor := executor.NewCommandExecutor(builtinService, builtinService)
	inputReader := reader.NewInputReader()
	signalHandler := builtins.NewSignalHandler()

	signalHandler.Setup()

	shell := NewShell(inputReader, commandParser, commandExecutor)
	shell.Run()
}
