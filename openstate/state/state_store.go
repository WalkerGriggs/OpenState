package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
)

type schemaFactory func() *memdb.TableSchema

type Config struct{}

type StateStore struct {
	config *Config
	DB     *memdb.MemDB
}

func NewStateStore(config *Config) (*StateStore, error) {
	s := &StateStore{
		config: config,
	}

	var err error

	s.DB, err = s.setupDB()
	if err != nil {
		return nil, fmt.Errorf("Failed to setup memdb: %v", err)
	}

	return s, nil
}

func (s *StateStore) setupDB() (*memdb.MemDB, error) {
	return memdb.NewMemDB(stateStoreSchema())
}

func stateStoreSchema() *memdb.DBSchema {
	db := &memdb.DBSchema{
		Tables: make(map[string]*memdb.TableSchema),
	}

	factories := []schemaFactory{
		definitionTableSchema,
		instanceTableSchema,
	}

	for _, fn := range factories {
		// TODO check for duplicate table schema
		schema := fn()
		db.Tables[schema.Name] = schema
	}

	return db
}

func definitionTableSchema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: "definition",
		Indexes: map[string]*memdb.IndexSchema{
			"id": &memdb.IndexSchema{
				Name:    "id",
				Unique:  true,
				Indexer: &memdb.StringFieldIndex{Field: "Name"},
			},
		},
	}
}

func instanceTableSchema() *memdb.TableSchema {
	return &memdb.TableSchema{
		Name: "instance",
		Indexes: map[string]*memdb.IndexSchema{
			"id": &memdb.IndexSchema{
				Name:    "id",
				Unique:  true,
				Indexer: &memdb.StringFieldIndex{Field: "Name"},
			},
		},
	}
}

func InsertDefinition() error {

}
