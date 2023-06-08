package internal

import (
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"

	braveds "github.com/brave/go-sync/datastore"
)

type SqliteDatastore struct {
	braveds.Datastore
	db *sql.DB
}

func NewSqliteDatastore() *SqliteDatastore {
	db, err := sql.Open("sqlite3", "./litesync.sqlite")
	if err != nil {
		panic(err)
	}
	return &SqliteDatastore{db: db}
}

//go:embed insert_sync_entity.sql
var insertSyncEntityQuery string

func (d *SqliteDatastore) InsertSyncEntity(entity *braveds.SyncEntity) (bool, error) {
	stmt, err := d.db.Prepare(insertSyncEntityQuery)
	return false, nil
}

func (d *SqliteDatastore) InsertSyncEntitiesWithServerTags(entities []*braveds.SyncEntity) error {
	for _, se := range entities {
		_, err := d.InsertSyncEntity(se)
		if err != nil {
			return err
		}
	}
	return nil
}

var updateSyncEntityQuery string

func (d *SqliteDatastore) UpdateSyncEntity(entity *braveds.SyncEntity, oldVersion int64) (conflict bool, delete bool, err error) {
	stmt, err := d.db.Prepare(updateSyncEntityQuery)
	return false, false, nil
}

func (d SqliteDatastore) GetUpdatesForType(dataType int, clientToken int64, fetchFolders bool, clientID string, maxSize int64) (bool, []braveds.SyncEntity, error) {
	return false, nil, nil
}

func (d SqliteDatastore) HasServerDefinedUniqueTag(clientID string, tag string) (bool, error) {
	return false, nil
}

func (d SqliteDatastore) GetClientItemCount(clientID string) (int, error) {
	return 0, nil
}

func (d SqliteDatastore) UpdateClientItemCount(clientID string, count int) error {
	return nil
}

func (d SqliteDatastore) ClearServerData(clientID string) ([]braveds.SyncEntity, error) {
	return nil, nil
}

func (d SqliteDatastore) DisableSyncChain(clientID string) error {
	return nil
}

func (d SqliteDatastore) IsSyncChainDisabled(clientID string) (bool, error) {
	return false, nil
}
