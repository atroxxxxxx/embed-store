CREATE INDEX IF NOT EXISTS hackernews_embedding_hnsw_idx
ON hackernews
USING hnsw (embedding vector_l2_ops)
WITH (
    m = 16,
    ef_construction = 64
);
