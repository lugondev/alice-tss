package peer

import (
	"encoding/json"
	"github.com/dgraph-io/badger"
)

type BadgerFSM struct {
	db *badger.DB
}

// Get fetch data from badgerDB
func (b BadgerFSM) Get(key string) (interface{}, error) {
	var keyByte = []byte(key)
	var data interface{}

	txn := b.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get(keyByte)
	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	var value = make([]byte, 0)
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})

	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	if value != nil && len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}

	if err != nil {
		data = map[string]interface{}{}
	}

	return data, err
}

// Set store data to badgerDB
func (b BadgerFSM) Set(key string, value interface{}) error {
	var data = make([]byte, 0)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if data == nil || len(data) <= 0 {
		return nil
	}

	txn := b.db.NewTransaction(true)
	err = txn.Set([]byte(key), data)
	if err != nil {
		txn.Discard()
		return err
	}

	return txn.Commit()
}

// SetArr store [data] to badgerDB
func (b BadgerFSM) SetArr(key string, value interface{}) error {
	var data = make([]byte, 0)
	data, err := json.Marshal([]interface{}{value})
	if err != nil {
		return err
	}

	if data == nil || len(data) <= 0 {
		return nil
	}
	existValue, err := b.Get(key)

	if existValue != nil && err == nil {
		//fmt.Println("existValue:", existValue)
		//fmt.Println("value:", value)
		newData := append(existValue.([]interface{}), value)
		data, err = json.Marshal(newData)
		if err != nil {
			return err
		}

	}
	txn := b.db.NewTransaction(true)

	err = txn.Set([]byte(key), data)
	if err != nil {
		txn.Discard()
		return err
	}

	return txn.Commit()
}

// Delete remove data from badgerDB
func (b BadgerFSM) Delete(key string) error {
	var keyByte = []byte(key)

	txn := b.db.NewTransaction(true)
	err := txn.Delete(keyByte)
	if err != nil {
		return err
	}

	return txn.Commit()
}

// NewBadger implementation using badgerDB
func NewBadger(badgerDB *badger.DB) *BadgerFSM {
	return &BadgerFSM{
		db: badgerDB,
	}
}
