package grpc

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor создаёт unary interceptor для аутентификации пользователей.
// Если authorization header отсутствует или невалиден, создаётся новый пользователь.
// Новый токен отправляется в response trailer.
func AuthInterceptor(authMgr *auth.Manager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			userID := uuid.New().String()
			token, err := authMgr.GenerateToken(userID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
			}

			trailer := metadata.New(map[string]string{"authorization": token})
			grpc.SetTrailer(ctx, trailer)

			ctx = context.WithValue(ctx, userIDKey, userID)
			return handler(ctx, req)
		}

		authValues := md.Get("authorization")
		if len(authValues) == 0 {
			userID := uuid.New().String()
			token, err := authMgr.GenerateToken(userID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
			}

			trailer := metadata.New(map[string]string{"authorization": token})
			grpc.SetTrailer(ctx, trailer)

			ctx = context.WithValue(ctx, userIDKey, userID)
			return handler(ctx, req)
		}

		userID, err := authMgr.ParseToken(authValues[0])
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				userID = uuid.New().String()
				token, genErr := authMgr.GenerateToken(userID)
				if genErr != nil {
					return nil, status.Errorf(codes.Internal, "failed to generate token: %v", genErr)
				}

				trailer := metadata.New(map[string]string{"authorization": token})
				grpc.SetTrailer(ctx, trailer)

				ctx = context.WithValue(ctx, userIDKey, userID)
				return handler(ctx, req)
			}
			return nil, status.Errorf(codes.Unauthenticated, "failed to parse token: %v", err)
		}

		ctx = context.WithValue(ctx, userIDKey, userID)
		return handler(ctx, req)
	}
}
