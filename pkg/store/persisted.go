package store

type Database interface {
	AddInitHook(hook DbInitHook)
	AddPutHook(hook DbPutHook)
	Init() error
	Get(t string, id string) ([]byte, error)
	Put(t string, id string, data []byte) (string, error)
	Delete(t string, id string) error
	Close()
}

type DbInitHook func(Database) error
type DbPutHook func(t string, id string, value []byte)
