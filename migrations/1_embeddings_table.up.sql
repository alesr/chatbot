CREATE EXTENSION vector;

CREATE TABLE embeddings (
    id  VARCHAR(255) PRIMARY KEY, 
    user_id VARCHAR(255) NOT NULL,
    collection_id  VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    tokens INTEGER NOT NULL,
    vector vector(1536) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX embeddings_collection_id_idx ON embeddings(collection_id);
CREATE INDEX embeddings_user_id_collection_id_idx ON embeddings(user_id, collection_id);
