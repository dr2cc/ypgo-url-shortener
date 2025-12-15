// Package generator is used for generating hash from string.
package generator // => ./internal/app/services/generator

// interface URLGenerator,поведение (метод)- (base_64_hash_generator)generator.GenerateIDFromString
type URLGenerator interface {
	GenerateIDFromString(url string) (string, error)
}
