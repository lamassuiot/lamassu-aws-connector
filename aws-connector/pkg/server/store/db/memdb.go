package db

import (
	"context"
	"encoding/json"
	"os"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/lamassuiot/aws-connector/pkg/server/store"
)

type BadgerDB struct {
	db *badger.DB
}

const (
	AWSConfig = "CONFIG"
)

func AWSThingConfig(deviceID string) string {
	return "THINGS_CONFIG_" + deviceID
}

func NewInMemoryDB() (store.DB, error) {
	err := os.RemoveAll("/tmp/badger")
	if err != nil {
		return nil, err
	}

	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))

	if err != nil {
		return nil, err
	}
	return &BadgerDB{
		db: db,
	}, nil
}

func (b *BadgerDB) UpdateAWSIoTCoreConfig(ctx context.Context, newConfig interface{}) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		bytes, err := json.Marshal(newConfig)
		if err != nil {
			return err
		}

		e := badger.NewEntry([]byte(AWSConfig), []byte(bytes)).WithTTL(time.Hour)
		err = txn.SetEntry(e)

		return nil
	})

	return err
}

func (b *BadgerDB) GetAWSIoTCoreConfig(ctx context.Context) (map[string]interface{}, error) {
	var valCopy []byte

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(AWSConfig))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})

		return nil
	})

	if err != nil {
		return make(map[string]interface{}), err
	}

	valMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(string(valCopy)), &valMap); err != nil {
		return make(map[string]interface{}), err
	}

	return valMap, nil
}

func (b *BadgerDB) DeleteAWSIoTCoreConfig(ctx context.Context) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(AWSConfig))
		return err
	})

	return err
}

func (b *BadgerDB) GetAWSIoTCoreThingConfig(ctx context.Context, deviceID string) ([]interface{}, error) {
	var valCopy []byte

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(AWSThingConfig(deviceID)))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})

		return nil
	})

	if err != nil {
		return make([]interface{}, 0), err
	}

	valMap := []interface{}{}
	if err := json.Unmarshal([]byte(string(valCopy)), &valMap); err != nil {
		return make([]interface{}, 0), err
	}

	return valMap, nil
}

func (b *BadgerDB) UpdateAWSIoTCoreThingConfig(ctx context.Context, deviceID string, newConfig interface{}) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		bytes, err := json.Marshal(newConfig)
		if err != nil {
			return err
		}

		e := badger.NewEntry([]byte(AWSThingConfig(deviceID)), []byte(bytes)).WithTTL(time.Hour)
		err = txn.SetEntry(e)

		return nil
	})

	return err
}

func (b *BadgerDB) DeleteAWSIoTCoreThingConfig(ctx context.Context, deviceID string) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(AWSThingConfig(deviceID)))
		return err
	})

	return err
}
