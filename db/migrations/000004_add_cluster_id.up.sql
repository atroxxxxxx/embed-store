ALTER TABLE hackernews
ADD COLUMN IF NOT EXISTS cluster_id INT;

CREATE INDEX IF NOT EXISTS hackernews_cluster_id_idx
ON hackernews (cluster_id);
