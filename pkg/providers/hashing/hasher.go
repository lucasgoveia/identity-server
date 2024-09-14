package hashing

type Hasher interface {
	Hash(text string, options ...interface{}) (string, error)
	Verify(text, hashed string) (bool, error)
	GetAlgorithm() string
}
