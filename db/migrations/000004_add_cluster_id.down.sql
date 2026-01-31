DROP INDEX IF EXISTS hackernews_cluster_id_idx;

ALTER TABLE hackernews
DROP COLUMN IF EXISTS cluster_id;
