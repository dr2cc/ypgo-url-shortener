package crypto

type Cryptographer interface {
	Encrypt(src []byte) ([]byte, error)
	Decrypt(src []byte) ([]byte, error)
}
