package generator

type Generator interface {
	GenerateIdFromString(url string) (string, error)
}
