package state

import (
	"fmt"

	memdb "github.com/hashicorp/go-memdb"

	"github.com/walkergriggs/openstate/openstate/structs"
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
				Indexer: &memdb.StringFieldIndex{Field: "ID"},
			},
		},
	}
}

func (s *StateStore) InsertDefinition(def *structs.Definition) error {
	txn := s.DB.Txn(true)
	defer txn.Abort()

	existing, err := txn.First("definition", "id", def.Name)
	if err != nil {
		return err
	}

	if existing != nil {
		return fmt.Errorf("Definition with name %s already exists.", def.Name)
	}

	if err := txn.Insert("definition", def); err != nil {
		return err
	}

	txn.Commit()

	return nil
}

func (s *StateStore) GetDefinitions() ([]*structs.Definition, error) {
	txn := s.DB.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("definition", "id")
	if err != nil {
		return nil, err
	}

	defs := make([]*structs.Definition, 0)

	for obj := it.Next(); obj != nil; obj = it.Next() {
		defs = append(defs, obj.(*structs.Definition))
	}

	return defs, nil
}

func (s *StateStore) GetDefinitionByName(name string) (*structs.Definition, error) {
	txn := s.DB.Txn(false)
	defer txn.Abort()

	obj, err := txn.First("definition", "id", name)
	if err != nil {
		return nil, err
	}

	if obj != nil {
		return obj.(*structs.Definition), nil
	}

	return nil, nil
}

func (s *StateStore) InsertInstance(instance *structs.Instance) error {
	txn := s.DB.Txn(true)
	defer txn.Abort()

	existing, err := txn.First("instance", "id", instance.ID)
	if err != nil {
		return err
	}

	if existing != nil {
		return fmt.Errorf("Instance with ID %s already exists.", instance.ID)
	}

	if err := txn.Insert("instance", instance); err != nil {
		return err
	}

	txn.Commit()

	return nil
}

func (s *StateStore) GetInstances() ([]*structs.Instance, error) {
	txn := s.DB.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("instance", "id")
	if err != nil {
		return nil, err
	}

	instances := make([]*structs.Instance, 0)

	for obj := it.Next(); obj != nil; obj = it.Next() {
		instances = append(instances, obj.(*structs.Instance))
	}

	return instances, nil
}

func (s *StateStore) GetInstanceByID(id string) (*structs.Instance, error) {
	txn := s.DB.Txn(false)
	defer txn.Abort()

	obj, err := txn.First("instance", "id", id)
	if err != nil {
		return nil, err
	}

	if obj != nil {
		return obj.(*structs.Instance), nil
	}

	return nil, nil
}
