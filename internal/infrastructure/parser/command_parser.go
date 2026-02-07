package parser

import (
	"minishell/internal/domain/entity"
	"minishell/internal/domain/usecase"
	"strings"
)

type CommandParser struct {
	envExpander usecase.EnvExpander
}

func NewCommandParser(envExpander usecase.EnvExpander) *CommandParser {
	return &CommandParser{
		envExpander: envExpander,
	}
}

func (p *CommandParser) Parse(line string) (*entity.Statement, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}
	return p.parseStatement(line), nil
}

func (p *CommandParser) parseStatement(line string) *entity.Statement {
	tokens := p.tokenizeWithOperators(line)
	if len(tokens) == 0 {
		return nil
	}

	stmt := &entity.Statement{}
	current := stmt

	for i, token := range tokens {
		if token == "&&" || token == "||" {
			if i+1 < len(tokens) {
				nextStmt := &entity.Statement{}
				current.NextOp = token
				current.NextStmt = nextStmt
				current = nextStmt
			}
		} else {
			current.Pipeline = p.parsePipeline(token)
		}
	}

	return stmt
}

func (p *CommandParser) tokenizeWithOperators(line string) []string {
	var tokens []string
	var current strings.Builder

	for i := 0; i < len(line); i++ {
		if i < len(line)-1 && line[i] == '&' && line[i+1] == '&' {
			if current.Len() > 0 {
				tokens = append(tokens, strings.TrimSpace(current.String()))
				current.Reset()
			}
			tokens = append(tokens, "&&")
			i++
			continue
		}
		if i < len(line)-1 && line[i] == '|' && line[i+1] == '|' {
			if current.Len() > 0 {
				tokens = append(tokens, strings.TrimSpace(current.String()))
				current.Reset()
			}
			tokens = append(tokens, "||")
			i++
			continue
		}
		current.WriteByte(line[i])
	}

	if current.Len() > 0 {
		tokens = append(tokens, strings.TrimSpace(current.String()))
	}

	return tokens
}

func (p *CommandParser) parsePipeline(line string) entity.Pipeline {
	parts := p.splitByPipe(line)
	var commands []entity.Command

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cmd := p.parseCommand(part)
		commands = append(commands, cmd)
	}

	return entity.Pipeline{Commands: commands}
}

func (p *CommandParser) splitByPipe(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(line); i++ {
		char := line[i]

		if char == '\\' && i+1 < len(line) {
			current.WriteByte(char)
			i++
			current.WriteByte(line[i])
			continue
		}

		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(char)
			continue
		}

		if !inQuotes && char == '|' {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}

		current.WriteByte(char)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func (p *CommandParser) parseCommand(line string) entity.Command {
	cmd := entity.Command{}
	tokens := p.tokenizeCommand(line)

	if len(tokens) == 0 {
		return cmd
	}

	var args []string
	i := 0

	for i < len(tokens) {
		token := tokens[i]
		if token == "<" && i+1 < len(tokens) {
			cmd.InputFile = p.envExpander.Expand(tokens[i+1])
			i += 2
			continue
		}
		if token == ">" && i+1 < len(tokens) {
			cmd.OutputFile = p.envExpander.Expand(tokens[i+1])
			i += 2
			continue
		}
		args = append(args, p.envExpander.Expand(token))
		i++
	}

	if len(args) > 0 {
		cmd.Name = args[0]
		if len(args) > 1 {
			cmd.Args = args[1:]
		}
	}

	return cmd
}

func (p *CommandParser) tokenizeCommand(line string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	i := 0
	for i < len(line) {
		char := line[i]

		if char == '\\' && i+1 < len(line) {
			if !inQuotes || (inQuotes && quoteChar == '"') {
				i++
				current.WriteByte(line[i])
				i++
				continue
			}
		}

		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteByte(char)
			}
			i++
			continue
		}

		if !inQuotes && (char == ' ' || char == '\t') {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			i++
			continue
		}

		current.WriteByte(char)
		i++
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

