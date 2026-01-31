package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/pgvector/pgvector-go"
)

type ClusterPoint struct {
	ID        int64
	Embedding pgvector.Vector
}

func (obj *Database) ClusterSource(ctx context.Context, limit int) ([]*ClusterPoint, error) {
	if limit <= 0 {
		limit = 10000
	}
	const request = `
	SELECT id, embedding
	FROM hackernews
	WHERE deleted = false
	LIMIT $1
`

	rows, err := obj.DB.QueryContext(ctx, request, limit)
	if err != nil {
		return nil, fmt.Errorf("cluster source query: %w", err)
	}
	defer rows.Close()

	out := make([]*ClusterPoint, 0, limit)
	for rows.Next() {
		var point ClusterPoint
		if err = rows.Scan(&point.ID, &point.Embedding); err != nil {
			return nil, fmt.Errorf("cluster source scan: %w", err)
		}
		out = append(out, &point)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cluster source rows: %w", err)
	}
	return out, nil
}

func (obj *Database) UpdateClusterIDs(ctx context.Context, ids []int64, clusterIDs []int32) error {
	lenIDs := len(ids)
	if lenIDs == 0 {
		return nil
	}
	lenClusterIDs := len(clusterIDs)
	if lenIDs != lenClusterIDs {
		return fmt.Errorf("ids len %d != cluser IDs len %d", lenIDs, lenClusterIDs)
	}

	var builder strings.Builder
	builder.Grow(128 + lenIDs*16)

	builder.WriteString(`
	UPDATE hackernews AS h
	SET cluster_id = v.cluster_id
	FROM (VALUES `)

	args := make([]any, 0, lenIDs*2)
	argNum := 1
	for i := range lenIDs {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(fmt.Sprintf("($%d::bigint,$%d::int)", argNum, argNum+1))
		args = append(args, ids[i], clusterIDs[i])
		argNum += 2
	}
	builder.WriteString(`) AS v(id, cluster_id)
	WHERE h.id = v.id;
`)

	if _, err := obj.DB.ExecContext(ctx, builder.String(), args...); err != nil {
		return fmt.Errorf("update cluster IDs: %w", err)
	}
	return nil
}
