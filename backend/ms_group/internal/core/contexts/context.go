package contexts

import (
	"context"
	"uuid"
)

type contextKey string

const groupIDContextKey = contextKey("group_id")

func SetGroupID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, groupIDContextKey, id)
}

func GetGroupID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(groupIDContextKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil()
}
