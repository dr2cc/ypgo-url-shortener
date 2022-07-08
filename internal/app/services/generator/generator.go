package generator

type URLGenerator interface {
	GenerateIDFromString(url string) (string, error)
}
