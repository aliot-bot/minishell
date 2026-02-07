package builtins

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type BuiltinService struct{}

func NewBuiltinService() *BuiltinService {
	return &BuiltinService{}
}

func (s *BuiltinService) IsBuiltin(name string) bool {
	switch name {
	case "cd", "pwd", "echo", "kill", "ps":
		return true
	}
	return false
}

func (s *BuiltinService) ExecuteBuiltin(name string, args []string) error {
	switch name {
	case "cd":
		return s.cd(args)
	case "pwd":
		return s.pwd(args)
	case "echo":
		return s.echo(args)
	case "kill":
		return s.kill(args)
	case "ps":
		return s.ps(args)
	default:
		return fmt.Errorf("unknown builtin: %s", name)
	}
}

func (s *BuiltinService) cd(args []string) error {
	var dir string
	if len(args) > 0 {
		dir = args[0]
	} else {
		dir = os.Getenv("HOME")
		if dir == "" {
			return fmt.Errorf("cd: HOME not set")
		}
	}

	err := os.Chdir(dir)
	if err != nil {
		return fmt.Errorf("cd: %v", err)
	}
	return nil
}

func (s *BuiltinService) pwd(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %v", err)
	}
	fmt.Println(dir)
	return nil
}

func (s *BuiltinService) echo(args []string) error {
	output := strings.Join(args, " ")
	fmt.Println(output)
	return nil
}

func (s *BuiltinService) kill(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("kill: usage: kill <pid>")
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("kill: invalid pid: %s", args[0])
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("kill: %v", err)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("kill: %v", err)
	}

	return nil
}

func (s *BuiltinService) ps(args []string) error {
	cmd := exec.Command("ps", "aux")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type SignalHandler struct{}

func NewSignalHandler() *SignalHandler {
	return &SignalHandler{}
}

func (h *SignalHandler) Setup() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	go func() {
		for {
			<-sigChan
		}
	}()
}
