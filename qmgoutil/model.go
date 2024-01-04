package qmgoutil

type QmgoModel interface {
	DBConfig() Config
	DatabaseName() string
	CollectionName() string
}
