package internal

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	braveds "github.com/brave/go-sync/datastore"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteDatastore struct {
	braveds.Datastore
	Db *sql.DB
}

func NewSqliteDatastore(filename string) (*SqliteDatastore, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return &SqliteDatastore{Db: db}, nil
}

const createTableQuery = `
CREATE TABLE IF NOT EXISTS sync_entities (
     client_id TEXT NOT NULL,
     id TEXT NOT NULL,
     parent_id TEXT,
     version INTEGER,
     mtime INTEGER,
     ctime INTEGER,
     name TEXT,
     non_unique_name TEXT,
     server_defined_unique_tag TEXT,
     deleted BOOLEAN,
     originator_cache_guid TEXT,
     originator_client_item_id TEXT,
     specifics BLOB,
     data_type INTEGER,
     folder BOOLEAN,
     client_defined_unique_tag TEXT,
     unique_position BLOB,
     data_type_mtime TEXT,
     expiration_time INTEGER,
     PRIMARY KEY (client_id, id)
)
`

const createSyncEntityIndex = `
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_client_tag
ON sync_entities (client_id, client_defined_unique_tag)
WHERE client_defined_unique_tag IS NOT NULL
`

type execFunc func(tx *sql.Tx) (sql.Result, error)

func (d *SqliteDatastore) ExecInTransaction(proxied execFunc) (*sql.Result, error) {
	tx, err := d.Db.BeginTx(context.Background(), nil)
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
	_, err := d.ExecInTransaction(func(tx *sql.Tx) (sql.Result, error) {
		// Create table
		if _, err := tx.Exec(createTableQuery); err != nil {
			return nil, err
		}
		// Create index
		if _, err := tx.Exec(createSyncEntityIndex); err != nil {
			return nil, err
		}
		return nil, nil // or return the result of the last operation
	})
	return err
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

func (d *SqliteDatastore) InsertSyncEntity(entity *braveds.SyncEntity) (bool, error) {
	// First, try to insert the main sync entity
	conflict, err := d.insertMainSyncEntity(entity)
	if err != nil || conflict {
		return conflict, err
	}

	// If entity has a client defined unique tag, also insert the tag item
	if entity.ClientDefinedUniqueTag != nil {
		tagConflict, err := d.insertTagItem(entity)
		if err != nil {
			return false, err
		}
		if tagConflict {
			return true, nil
		}
	}

	return false, nil
}

func (d *SqliteDatastore) insertMainSyncEntity(entity *braveds.SyncEntity) (bool, error) {
	const query = `
        INSERT INTO sync_entities (
            client_id, id, parent_id, version, mtime, ctime, name, non_unique_name,
            server_defined_unique_tag, deleted, originator_cache_guid,
            originator_client_item_id, specifics, data_type, folder,
            client_defined_unique_tag, unique_position, data_type_mtime, expiration_time
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.Db.Exec(query,
		entity.ClientID,
		entity.ID,
		entity.ParentID,
		entity.Version,
		entity.Mtime,
		entity.Ctime,
		entity.Name,
		entity.NonUniqueName,
		entity.ServerDefinedUniqueTag,
		entity.Deleted,
		entity.OriginatorCacheGUID,
		entity.OriginatorClientItemID,
		entity.Specifics,
		entity.DataType,
		entity.Folder,
		entity.ClientDefinedUniqueTag,
		entity.UniquePosition,
		entity.DataTypeMtime,
		entity.ExpirationTime,
	)

	if err != nil {
		// Check if it's a conflict (duplicate key) error
		if isUniqueConstraintError(err) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (d *SqliteDatastore) insertTagItem(entity *braveds.SyncEntity) (bool, error) {
	if entity.ClientDefinedUniqueTag == nil {
		return false, nil
	}

	const query = `
        INSERT INTO sync_entities (
            client_id, id, mtime, ctime
        ) VALUES (?, ?, ?, ?)`

	// Set current time for mtime and ctime if not already set
	now := time.Now().Unix()
	mtime := entity.Mtime
	ctime := entity.Ctime
	if mtime == nil {
		mtime = &now
	}
	if ctime == nil {
		ctime = &now
	}

	_, err := d.Db.Exec(query,
		entity.ClientID,
		"Client#"+*entity.ClientDefinedUniqueTag, // Construct the tag ID
		mtime,
		ctime,
	)

	if err != nil {
		// Check if it's a conflict (duplicate key) error
		if isUniqueConstraintError(err) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite unique constraint violation error
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed") ||
		strings.Contains(strings.ToLower(err.Error()), "constraint failed")
}

func (d *SqliteDatastore) InsertSyncEntitiesWithServerTags(entities []*braveds.SyncEntity) error {
	fail := func(err error) error {
		return fmt.Errorf("InsertSyncEntitiesWithServerTags: %v", err)
	}
	tx, err := d.Db.BeginTx(context.Background(), nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	for _, se := range entities {
		if se.ServerDefinedUniqueTag != nil {
			// Check for existing tag item
			var exists bool
			err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM sync_entities WHERE client_id = ? AND id = ?)",
				se.ClientID, "Server#"+*se.ServerDefinedUniqueTag).Scan(&exists)
			if err != nil {
				return fail(err)
			}
			if exists {
				return fmt.Errorf("server tag already exists")
			}

			// Insert tag item
			_, err = tx.Exec(insertSyncEntityQuery,
				"Server#"+*se.ServerDefinedUniqueTag, se.ClientID, 0, se.Mtime, nil, nil,
				nil, nil, nil, nil, false, false)
			if err != nil {
				return fail(err)
			}
		}

		// Insert sync entity
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
UPDATE sync_entities
SET "version" = ?,
    mtime = ?,
    specifics = ?,
    datatype_mtime = ?,
    unique_position = ?,
    parent_id = ?,
    "name" = ?,
    "non_unique_name" = ?,
    "deleted" = ?,
    "folder" = ?
WHERE client_id = ? AND id = ? AND "version" = ?
`

func (d *SqliteDatastore) UpdateSyncEntity(se *braveds.SyncEntity, oldVersion int64) (conflict bool, delete bool, err error) {
	fail := func(err error) (bool, bool, error) {
		return false, false, fmt.Errorf("UpdateSyncEntity: %v", err)
	}

	tx, err := d.Db.BeginTx(context.Background(), nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(updateSyncEntityQuery,
		se.Version, se.Mtime, se.Specifics, se.DataTypeMtime,
		se.UniquePosition, se.ParentID, se.Name, se.NonUniqueName, se.Deleted, se.Folder,
		se.ClientID, se.ID, oldVersion)
	if err != nil {
		return fail(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fail(err)
	}

	if rowsAffected == 0 {
		return true, false, nil // Conflict
	}

	if se.Deleted != nil && *se.Deleted && se.ClientDefinedUniqueTag != nil {
		_, err = tx.Exec("DELETE FROM sync_entities WHERE client_id = ? AND id = ?", se.ClientID, "Client#"+*se.ClientDefinedUniqueTag)
		if err != nil {
			return fail(err)
		}
		delete = true
	}

	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return false, delete, nil
}

const getClientItemCountQuery = `
SELECT COUNT(*)
FROM sync_entities
WHERE client_id = ?
`

func (d SqliteDatastore) GetClientItemCount(clientID string) (*braveds.ClientItemCounts, error) {
	tx, err := d.Db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	var count braveds.ClientItemCounts
	row := tx.QueryRow(getClientItemCountQuery, clientID)
	row.Scan(&count)
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &count, nil
}

func (d SqliteDatastore) GetUpdatesForType(dataType int, clientToken int64, fetchFolders bool, clientID string, maxSize int64) (bool, []braveds.SyncEntity, error) {
	return false, nil, nil
}

func (d SqliteDatastore) HasServerDefinedUniqueTag(clientID string, tag string) (bool, error) {
	return false, nil
}

func (d SqliteDatastore) UpdateClientItemCount(counts *braveds.ClientItemCounts, newNormalItemCount int, newHistoryItemCount int) error {
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
