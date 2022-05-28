package generator

type Generator interface {
	GenerateIDFromString(url string) (string, error)
}
