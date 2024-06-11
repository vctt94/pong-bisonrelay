package bot

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func attachClientIDToContext(ctx context.Context, clientID string) context.Context {
	md := metadata.New(map[string]string{
		"client-id": clientID,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
