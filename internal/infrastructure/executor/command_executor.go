package executor

import (
	"fmt"
	"io"
	"minishell/internal/domain/entity"
	"minishell/internal/domain/usecase"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

type CommandExecutor struct {
	builtinChecker  usecase.BuiltinChecker
	builtinExecutor usecase.BuiltinExecutor
}

func NewCommandExecutor(builtinChecker usecase.BuiltinChecker, builtinExecutor usecase.BuiltinExecutor) *CommandExecutor {
	return &CommandExecutor{
		builtinChecker:  builtinChecker,
		builtinExecutor: builtinExecutor,
	}
}

func (e *CommandExecutor) ExecuteStatement(stmt *entity.Statement) error {
	if stmt == nil {
		return nil
	}

	err := e.ExecutePipeline(stmt.Pipeline)

	if stmt.NextOp != "" && stmt.NextStmt != nil {
		if stmt.NextOp == "&&" {
			if err == nil {
				return e.ExecuteStatement(stmt.NextStmt)
			}
		} else if stmt.NextOp == "||" {
			if err != nil {
				return e.ExecuteStatement(stmt.NextStmt)
			}
		}
	}

	return err
}

func (e *CommandExecutor) ExecutePipeline(pl entity.Pipeline) error {
	if len(pl.Commands) == 0 {
		return nil
	}

	if len(pl.Commands) == 1 {
		return e.ExecuteCommand(pl.Commands[0], nil, nil)
	}

	numCommands := len(pl.Commands)

	type pipePair struct {
		r *os.File
		w *os.File
	}
	pipes := make([]pipePair, numCommands-1)

	for i := 0; i < numCommands-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			return fmt.Errorf("pipe error: %v", err)
		}
		pipes[i] = pipePair{r: r, w: w}
	}

	defer func() {
		for _, p := range pipes {
			p.r.Close()
			p.w.Close()
		}
	}()

	errChan := make(chan error, numCommands)
	processChan := make(chan *exec.Cmd, numCommands)

	for i, cmd := range pl.Commands {
		go func(idx int, command entity.Command) {
			var stdin io.Reader
			var stdout io.Writer

			if idx > 0 {
				stdin = pipes[idx-1].r
			}

			if idx < numCommands-1 {
				stdout = pipes[idx].w
			}

			err := e.ExecuteCommandWithProcessTracking(command, stdin, stdout, processChan)

			if idx < numCommands-1 {
				pipes[idx].w.Close()
			}

			errChan <- err
		}(i, cmd)
	}

	var processes []*exec.Cmd
	var processesMutex sync.Mutex

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	go func() {
		for sig := range sigChan {
			processesMutex.Lock()
			for _, proc := range processes {
				if proc != nil && proc.Process != nil {
					syscall.Kill(-proc.Process.Pid, sig.(syscall.Signal))
				}
			}
			processesMutex.Unlock()
		}
	}()

	go func() {
		for i := 0; i < numCommands; i++ {
			proc := <-processChan
			if proc != nil {
				processesMutex.Lock()
				processes = append(processes, proc)
				processesMutex.Unlock()
			}
		}
	}()

	var lastErr error
	for i := 0; i < numCommands; i++ {
		if err := <-errChan; err != nil {
			lastErr = err
		}
	}

	signal.Stop(sigChan)
	close(sigChan)

	return lastErr
}

func (e *CommandExecutor) ExecuteCommand(cmd entity.Command, stdin io.Reader, stdout io.Writer) error {
	if e.builtinChecker.IsBuiltin(cmd.Name) {
		return e.executeBuiltinWithIO(cmd, stdin, stdout)
	}
	return e.executeExternalCommandWithTracking(cmd, stdin, stdout, nil)
}

func (e *CommandExecutor) executeBuiltinWithIO(cmd entity.Command, stdin io.Reader, stdout io.Writer) error {
	originalStdin, originalStdout := os.Stdin, os.Stdout
	defer func() {
		os.Stdin, os.Stdout = originalStdin, originalStdout
	}()

	switch {
	case stdin != nil:
		if file, ok := stdin.(*os.File); ok {
			os.Stdin = file
		}
	case cmd.InputFile != "":
		file, err := os.Open(cmd.InputFile)
		if err != nil {
			return fmt.Errorf("cannot open input file: %v", err)
		}
		defer file.Close()
		os.Stdin = file
	}

	switch {
	case stdout != nil:
		if file, ok := stdout.(*os.File); ok {
			os.Stdout = file
		}
	case cmd.OutputFile != "":
		file, err := os.Create(cmd.OutputFile)
		if err != nil {
			return fmt.Errorf("cannot create output file: %v", err)
		}
		defer file.Close()
		os.Stdout = file
	}

	return e.builtinExecutor.ExecuteBuiltin(cmd.Name, cmd.Args)
}

func (e *CommandExecutor) ExecuteCommandWithProcessTracking(cmd entity.Command, stdin io.Reader, stdout io.Writer, processChan chan<- *exec.Cmd) error {
	if e.builtinChecker.IsBuiltin(cmd.Name) {
		if processChan != nil {
			processChan <- nil
		}
		return e.executeBuiltinWithIO(cmd, stdin, stdout)
	}
	return e.executeExternalCommandWithTracking(cmd, stdin, stdout, processChan)
}

func (e *CommandExecutor) executeExternalCommandWithTracking(cmd entity.Command, stdin io.Reader, stdout io.Writer, processChan chan<- *exec.Cmd) error {
	execCmd := exec.Command(cmd.Name, cmd.Args...)

	switch {
	case stdin != nil:
		execCmd.Stdin = stdin
	case cmd.InputFile != "":
		file, err := os.Open(cmd.InputFile)
		if err != nil {
			return fmt.Errorf("cannot open input file: %v", err)
		}
		defer file.Close()
		execCmd.Stdin = file
	default:
		execCmd.Stdin = os.Stdin
	}

	switch {
	case stdout != nil:
		execCmd.Stdout = stdout
	case cmd.OutputFile != "":
		file, err := os.Create(cmd.OutputFile)
		if err != nil {
			return fmt.Errorf("cannot create output file: %v", err)
		}
		defer file.Close()
		execCmd.Stdout = file
	default:
		execCmd.Stdout = os.Stdout
	}

	execCmd.Stderr = os.Stderr
	execCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := execCmd.Start(); err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}

	if processChan != nil {
		processChan <- execCmd
	}

	if processChan == nil {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
		go func() {
			for sig := range sigChan {
				if execCmd.Process != nil {
					syscall.Kill(-execCmd.Process.Pid, sig.(syscall.Signal))
				}
			}
		}()
		err := execCmd.Wait()
		signal.Stop(sigChan)
		close(sigChan)
		return err
	}

	return execCmd.Wait()
}
