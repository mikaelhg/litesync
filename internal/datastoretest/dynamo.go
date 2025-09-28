package datastoretest

import (
	"database/sql"
	"fmt"

	"github.com/brave/go-sync/datastore"
	"github.com/mikaelhg/litesync/internal"
)

// DeleteTable deletes datastore.Table in dynamoDB.
func DeleteTable(dynamo *internal.SqliteDatastore) error {
	// _, err := dynamo.DeleteTable(
	// 	&dynamodb.DeleteTableInput{TableName: aws.String(datastore.Table)})
	// if err != nil {
	// 	if aerr, ok := err.(awserr.Error); ok {
	// 		// Return as successful if the table is not existed.
	// 		if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
	// 			return nil
	// 		}
	// 	} else {
	// 		return fmt.Errorf("error deleting table: %w", err)
	// 	}
	// }

	// return dynamo.WaitUntilTableNotExists(
	// 	&dynamodb.DescribeTableInput{TableName: aws.String(datastore.Table)})

	return nil
}

// CreateTable creates datastore.Table in dynamoDB.
func CreateTable(dynamo *internal.SqliteDatastore) error {
	// _, b, _, _ := runtime.Caller(0)
	// root := filepath.Join(filepath.Dir(b), "../../")
	// raw, err := os.ReadFile(filepath.Join(root, "schema/dynamodb/table.json"))
	// if err != nil {
	// 	return fmt.Errorf("error reading table.json: %w", err)
	// }

	// var input dynamodb.CreateTableInput
	// err = json.Unmarshal(raw, &input)
	// if err != nil {
	// 	return fmt.Errorf("error unmarshalling raw data from table.json: %w", err)
	// }
	// input.TableName = aws.String(datastore.Table)

	// _, err = dynamo.CreateTable(&input)
	// if err != nil {
	// 	return fmt.Errorf("error creating table: %w", err)
	// }

	// return dynamo.WaitUntilTableExists(
	// 	&dynamodb.DescribeTableInput{TableName: aws.String(datastore.Table)})

	return nil
}

// ResetTable deletes and creates datastore.Table in dynamoDB.
func ResetTable(dynamo *internal.SqliteDatastore) error {
	// if err := DeleteTable(dynamo); err != nil {
	// 	return fmt.Errorf("error deleting table to reset table: %w", err)
	// }
	// return CreateTable(dynamo)

	return nil
}

// ScanSyncEntities scans the SQLite table and returns all sync items.
func ScanSyncEntities(sqlite *internal.SqliteDatastore) ([]datastore.SyncEntity, error) {
	var syncItems []datastore.SyncEntity
	var scanErr error

	_, err := sqlite.ExecInTransaction(func(tx *sql.Tx) (sql.Result, error) {
		const query = `
            SELECT client_id, id, parent_id, version, mtime, ctime, name, non_unique_name,
                   server_defined_unique_tag, deleted, originator_cache_guid,
                   originator_client_item_id, specifics, data_type, folder,
                   client_defined_unique_tag, unique_position, data_type_mtime, expiration_time
            FROM sync_entities`

		rows, err := tx.Query(query)
		if err != nil {
			scanErr = fmt.Errorf("error querying sync entities: %w", err)
			return nil, scanErr
		}
		defer rows.Close()

		syncItems = []datastore.SyncEntity{}
		for rows.Next() {
			var entity datastore.SyncEntity
			err := rows.Scan(
				&entity.ClientID,
				&entity.ID,
				&entity.ParentID,
				&entity.Version,
				&entity.Mtime,
				&entity.Ctime,
				&entity.Name,
				&entity.NonUniqueName,
				&entity.ServerDefinedUniqueTag,
				&entity.Deleted,
				&entity.OriginatorCacheGUID,
				&entity.OriginatorClientItemID,
				&entity.Specifics,
				&entity.DataType,
				&entity.Folder,
				&entity.ClientDefinedUniqueTag,
				&entity.UniquePosition,
				&entity.DataTypeMtime,
				&entity.ExpirationTime,
			)
			if err != nil {
				scanErr = fmt.Errorf("error scanning sync entity: %w", err)
				return nil, scanErr
			}
			syncItems = append(syncItems, entity)
		}

		if err = rows.Err(); err != nil {
			scanErr = fmt.Errorf("error iterating sync entities: %w", err)
			return nil, scanErr
		}

		return nil, nil // No result to return for SELECT queries
	})

	if err != nil && scanErr == nil {
		return nil, err
	}
	if scanErr != nil {
		return nil, scanErr
	}

	return syncItems, nil
}

// ScanTagItems scans the SQLite table and returns all tag items.
func ScanTagItems(sqlite *internal.SqliteDatastore) ([]datastore.ServerClientUniqueTagItem, error) {
	const query = `
        SELECT client_id, id, mtime, ctime
        FROM sync_entities
        WHERE expiration_time IS NULL 
        AND version IS NULL`

	rows, err := sqlite.Db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying tag items: %w", err)
	}
	defer rows.Close()

	tagItems := []datastore.ServerClientUniqueTagItem{}
	for rows.Next() {
		var entity datastore.ServerClientUniqueTagItem
		err := rows.Scan(
			&entity.ClientID,
			&entity.ID,
			&entity.Mtime,
			&entity.Ctime,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning tag item: %w", err)
		}
		tagItems = append(tagItems, entity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag items: %w", err)
	}

	return tagItems, nil
}

// ScanClientItemCounts scans the dynamoDB table and returns all client item
// counts.
func ScanClientItemCounts(dynamo *internal.SqliteDatastore) ([]datastore.ClientItemCounts, error) {
	// filter := expression.AttributeExists(expression.Name("ItemCount"))
	// expr, err := expression.NewBuilder().WithFilter(filter).Build()
	// if err != nil {
	// 	return nil, fmt.Errorf("error building expression to scan item counts: %w", err)
	// }

	// input := &dynamodb.ScanInput{
	// 	ExpressionAttributeNames:  expr.Names(),
	// 	ExpressionAttributeValues: expr.Values(),
	// 	FilterExpression:          expr.Filter(),
	// 	TableName:                 aws.String(datastore.Table),
	// }
	// out, err := dynamo.Scan(input)
	// if err != nil {
	// 	return nil, fmt.Errorf("error doing scan for item counts: %w", err)
	// }
	// clientItemCounts := []datastore.ClientItemCounts{}
	// err = dynamodbattribute.UnmarshalListOfMaps(out.Items, &clientItemCounts)
	// if err != nil {
	// 	return nil, fmt.Errorf("error unmarshalling item counts: %w", err)
	// }

	// return clientItemCounts, nil

	return nil, nil
}
