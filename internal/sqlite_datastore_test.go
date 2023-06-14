package internal_test

import (
	"testing"

	"github.com/brave/go-sync/datastore"
	"github.com/mikaelhg/litesync/internal"
	"github.com/stretchr/testify/assert"
)

func TestInsertion(t *testing.T) {
	ds := internal.NewSqliteDatastore(":memory:")
	ds.CreateTable()

	one := int64(1)

	se := datastore.SyncEntity{
		ClientID: "foobar",
		ID:       "xyzzy",
		Version:  &one,
	}

	ok, err := ds.InsertSyncEntity(&se)
	assert.True(t, ok)
	assert.NoError(t, err, "can't have an error here")

	count, err := ds.GetClientItemCount(se.ClientID)
	assert.NoError(t, err, "can't have an error here")
	assert.Equal(t, count, 1, "we inserted a single row")
}
