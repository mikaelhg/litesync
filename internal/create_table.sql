CREATE TABLE sync_entities (
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
);
