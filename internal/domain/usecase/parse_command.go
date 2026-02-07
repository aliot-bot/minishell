package usecase

import "minishell/internal/domain/entity"

type CommandParser interface {
	Parse(line string) (*entity.Statement, error)
}
