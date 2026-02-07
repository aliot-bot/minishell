package usecase

type InputReader interface {
	ReadCommand() (string, bool)
}
