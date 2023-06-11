package internal

import (
	"context"
	"database/sql"
	"fmt"

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

const insertSyncEntityQuery = `
INSERT INTO sync_entities (
	id,
	client_id,
	"version",
	mtime,
	specifics,
	datatype_mtime,
	unique_position,
	parent_id,
	"name",
	"non_unique_name",
	"deleted",
	"folder"
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

func (d *SqliteDatastore) InsertSyncEntity(se *braveds.SyncEntity) (bool, error) {
	fail := func(err error) (bool, error) {
		return false, fmt.Errorf("InsertSyncEntity: %v", err)
	}
	tx, err := d.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()
	if se.ClientDefinedUniqueTag != nil {
		// Additional item for ensuring tag's uniqueness for a specific client.
		item := braveds.NewServerClientUniqueTagItem(se.ClientID,
			*se.ClientDefinedUniqueTag, false)
		// Normal sync item
	} else {
		_, err = tx.Exec(insertSyncEntityQuery,
			se.ID, se.ClientID, se.Version, se.Mtime, se.Specifics, se.DataTypeMtime,
			se.UniquePosition, se.ParentID, se.Name, se.NonUniqueName, se.Deleted, se.Folder)
		if err != nil {
			return fail(err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fail(err)
	}
	return true, nil
}

func (d *SqliteDatastore) InsertSyncEntitiesWithServerTags(entities []*braveds.SyncEntity) error {
	fail := func(err error) error {
		return fmt.Errorf("InsertSyncEntitiesWithServerTags: %v", err)
	}
	tx, err := d.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()
	for _, se := range entities {
		_, err = tx.Exec(insertSyncEntityQuery,
			se.ID, se.ClientID, se.Version, se.Mtime, se.Specifics, se.DataTypeMtime,
			se.UniquePosition, se.ParentID, se.Name, se.NonUniqueName, se.Deleted, se.Folder)
		if err != nil {
			return fail(err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fail(err)
	}
	return nil
}

const updateSyncEntityQuery = `
`

func (d *SqliteDatastore) UpdateSyncEntity(se *braveds.SyncEntity, oldVersion int64) (conflict bool, delete bool, err error) {
	fail := func(err error) (bool, bool, error) {
		return false, false, fmt.Errorf("InserUpdateSyncEntitytSyncEntity: %v", err)
	}
	tx, err := d.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fail(err)
	}
	_, err = tx.Exec(updateSyncEntityQuery,
		se.ID, se.ClientID, se.Version, se.Mtime, se.Specifics, se.DataTypeMtime,
		se.UniquePosition, se.ParentID, se.Name, se.NonUniqueName, se.Deleted, se.Folder)
	if err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return fail(err)
	}
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
