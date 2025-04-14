-- +goose Up
CREATE TABLE datasets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    public BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(user_id, name)
);

CREATE TABLE dataset_fields (
    id UUID PRIMARY KEY,
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    data_type TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(dataset_id, name)
);

CREATE TABLE dataset_records (
    id UUID PRIMARY KEY,
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE record_values (
    record_id UUID NOT NULL REFERENCES dataset_records(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES dataset_fields(id) ON DELETE CASCADE,
    value TEXT,
    PRIMARY KEY(record_id, field_id)
);

CREATE INDEX idx_dataset_fields_dataset_id ON dataset_fields(dataset_id);
CREATE INDEX idx_dataset_records_dataset_id ON dataset_records(dataset_id);

-- +goose Down
DROP INDEX idx_dataset_fields_dataset_id;
DROP INDEX idx_dataset_records_dataset_id;
DROP TABLE record_values;
DROP TABLE dataset_records;
DROP TABLE dataset_fields;
DROP TABLE datasets;
