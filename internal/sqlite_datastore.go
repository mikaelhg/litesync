package internal

import "github.com/brave/go-sync/datastore"

type SqliteDatastore struct {
	datastore.Datastore
}

func NewSqliteDatastore() *SqliteDatastore {
	return &SqliteDatastore{}
}

func (d SqliteDatastore) InsertSyncEntity(entity *datastore.SyncEntity) (bool, error) {
	return false, nil
}

func (d SqliteDatastore) InsertSyncEntitiesWithServerTags(entities []*datastore.SyncEntity) error {
	return nil
}

func (d SqliteDatastore) UpdateSyncEntity(entity *datastore.SyncEntity, oldVersion int64) (conflict bool, delete bool, err error) {
	return false, false, nil
}

func (d SqliteDatastore) GetUpdatesForType(dataType int, clientToken int64, fetchFolders bool, clientID string, maxSize int64) (bool, []datastore.SyncEntity, error) {
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

func (d SqliteDatastore) ClearServerData(clientID string) ([]datastore.SyncEntity, error) {
	return nil, nil
}

func (d SqliteDatastore) DisableSyncChain(clientID string) error {
	return nil
}

func (d SqliteDatastore) IsSyncChainDisabled(clientID string) (bool, error) {
	return false, nil
}
