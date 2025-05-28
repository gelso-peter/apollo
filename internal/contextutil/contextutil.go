package contextutil

import "context"

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
