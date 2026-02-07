package usecase

type EnvExpander interface {
	Expand(s string) string
}
