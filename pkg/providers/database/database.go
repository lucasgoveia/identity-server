package database

type Database interface {
	Close() error
	GetProviderType() string
}
