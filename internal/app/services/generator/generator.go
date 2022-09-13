// Package generator is used for generating hash from string.
package generator

// URLGenerator generates hash from string.
type URLGenerator interface {
	GenerateIDFromString(url string) (string, error)
}
