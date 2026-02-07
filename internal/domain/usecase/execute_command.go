package usecase

import "minishell/internal/domain/entity"

type CommandExecutor interface {
	ExecuteStatement(stmt *entity.Statement) error
}

type BuiltinChecker interface {
	IsBuiltin(name string) bool
}

type BuiltinExecutor interface {
	ExecuteBuiltin(name string, args []string) error
}
