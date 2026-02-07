package entity

type Command struct {
	Name       string
	Args       []string
	InputFile  string
	OutputFile string
}

type Pipeline struct {
	Commands []Command
}

type Statement struct {
	Pipeline Pipeline
	NextOp   string
	NextStmt *Statement
}
