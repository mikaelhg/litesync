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

func NewSqliteDatastore(filename string) (*SqliteDatastore, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return &SqliteDatastore{db: db}, nil
}

const createTableQuery = `
CREATE TABLE IF NOT EXISTS sync_entities (
    id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    "version" INTEGER,
    mtime INTEGER,
    specifics BLOB,
    datatype_mtime TEXT,
    unique_position BLOB,
    parent_id TEXT,
    "name" TEXT,
    "non_unique_name" TEXT,
    "deleted" BOOLEAN,
    "folder" BOOLEAN,
    PRIMARY KEY (client_id, id)
)
`

type execFunc func(tx *sql.Tx) (sql.Result, error)

func (d *SqliteDatastore) execInTransaction(proxied execFunc) (*sql.Result, error) {
	tx, err := d.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	txRes, txErr := proxied(tx)
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &txRes, txErr
}

func (d *SqliteDatastore) CreateTable() error {
	_, err := d.execInTransaction(func(tx *sql.Tx) (sql.Result, error) {
		return tx.Exec(createTableQuery)
	})
	if err != nil {
		return err
	}
	return nil
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
		_ = braveds.NewServerClientUniqueTagItem(se.ClientID,
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

const getClientItemCountQuery = `
SELECT COUNT(*)
FROM sync_entities
WHERE client_id = ?
`

func (d SqliteDatastore) GetClientItemCount(clientID string) (int, error) {
	fail := func(err error) (int, error) {
		return -1, fmt.Errorf("GetClientItemCount: %v", err)
	}
	tx, err := d.db.BeginTx(context.Background(), nil)
	if err != nil {
		return fail(err)
	}
	var count int = -1
	row := tx.QueryRow(getClientItemCountQuery, clientID)
	row.Scan(&count)
	if err = tx.Commit(); err != nil {
		return fail(err)
	}
	return count, nil
}

func (d SqliteDatastore) GetUpdatesForType(dataType int, clientToken int64, fetchFolders bool, clientID string, maxSize int64) (bool, []braveds.SyncEntity, error) {
	return false, nil, nil
}

func (d SqliteDatastore) HasServerDefinedUniqueTag(clientID string, tag string) (bool, error) {
	return false, nil
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
