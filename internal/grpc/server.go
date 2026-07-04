// Package grpc предоставляет gRPC-сервер для сервиса сокращения URL.
package grpc

import (
	"context"
	"errors"

	pb "PolyakovEvg/go-musthave-shortener-tpl/api/shortener"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/auth"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/repository"
	"PolyakovEvg/go-musthave-shortener-tpl/internal/service/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// contextKey — тип для ключей контекста.
type contextKey struct{}

// userIDKey — ключ для хранения идентификатора пользователя в контексте.
var userIDKey = contextKey{}

// UserIDFromContext извлекает идентификатор пользователя из контекста запроса.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// ShortenerServer реализует gRPC-сервис ShortenerService.
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	urlSvc *url.URLService
	auth   *auth.Manager
}

// NewShortenerServer создаёт новый gRPC-сервер с указанным сервисом URL.
func NewShortenerServer(urlSvc *url.URLService, auth *auth.Manager) *ShortenerServer {
	return &ShortenerServer{
		urlSvc: urlSvc,
		auth:   auth,
	}
}

// ShortenURL сокращает URL и возвращает короткий URL.
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is empty")
	}

	userID, _ := UserIDFromContext(ctx)
	shortURL, err := s.urlSvc.SaveShorten(userID, req.Url)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return &pb.URLShortenResponse{Result: shortURL}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to save url: %v", err)
	}

	return &pb.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL возвращает оригинальный URL по короткому идентификатору.
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is empty")
	}

	url, err := s.urlSvc.GetOriginal(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "url not found")
	}

	if url.IsDeleted {
		return nil, status.Error(codes.NotFound, "url deleted")
	}

	return &pb.URLExpandResponse{Result: url.OriginalURL}, nil
}

// ListUserURLs возвращает все URL, созданные пользователем.
func (s *ShortenerServer) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, _ := UserIDFromContext(ctx)
	urls, err := s.urlSvc.GetUserURLs(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user urls: %v", err)
	}

	if len(urls) == 0 {
		return &pb.UserURLsResponse{}, nil
	}

	pbURLs := make([]*pb.URLData, len(urls))
	for i, url := range urls {
		pbURLs[i] = &pb.URLData{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}
	}

	return &pb.UserURLsResponse{Url: pbURLs}, nil
}

// Register регистрирует gRPC-сервер на сервере с указанным auth manager.
func Register(server *grpc.Server, urlService *url.URLService, authMgr *auth.Manager) {
	pb.RegisterShortenerServiceServer(server, NewShortenerServer(urlService, authMgr))
}
