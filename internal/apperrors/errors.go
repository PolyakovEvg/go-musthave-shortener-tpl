// Package apperrors предоставляет ошибки уровня приложения.
// Этот пакет служит абстракцией между транспортным слоем (HTTP/gRPC) и слоем репозитория.
// Хендлеры и gRPC-сервер должны использовать ошибки из этого пакета, а не из repository.
package apperrors

import "errors"

// ErrURLConflict возникает при попытке сохранить URL, который уже существует.
var ErrURLConflict = errors.New("url already exists")

// ErrURLNotFound возникает, когда запрашиваемый URL не найден.
var ErrURLNotFound = errors.New("url not found")

// ErrEmptyURL возникает при попытке сохранить пустой URL.
var ErrEmptyURL = errors.New("url is empty")

// ErrEmptyBatch возникает при попытке сохранить пустой пакет URL.
var ErrEmptyBatch = errors.New("empty batch")

// ErrEmptyUserID возникает при передаче пустого идентификатора пользователя.
var ErrEmptyUserID = errors.New("empty user id")

// ErrDeletedURL возникает при обращении к удалённому URL.
var ErrDeletedURL = errors.New("url is deleted")
