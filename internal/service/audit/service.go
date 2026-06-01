package audit

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/logger"
	"fmt"
)

// AuditService управляет отправкой событий аудита наблюдателям.
// Поддерживает регистрацию нескольких наблюдателей и асинхронную отправку событий.
type AuditService struct {
	logger    *logger.Logger
	observers []Observer
}

// NewAuditService создаёт новый сервис аудита с указанным логгером.
func NewAuditService(logger *logger.Logger) *AuditService {
	return &AuditService{
		logger:    logger,
		observers: make([]Observer, 0),
	}
}

// Register добавляет наблюдателя для получения событий аудита.
func (s *AuditService) Register(observer Observer) {
	s.observers = append(s.observers, observer)
	s.logger.Zap.Infow("Registered observer",
		"type", fmt.Sprintf("%T", observer),
	)
}

// Notify отправляет событие аудита всем зарегистрированным наблюдателям.
// Отправка происходит асинхронно в отдельных горутинах.
func (s *AuditService) Notify(event AuditEvent) {
	for _, observer := range s.observers {
		go func(obs Observer) {
			if err := obs.Send(event); err != nil {
				s.logger.Zap.Errorw("Observer error",
					"type", fmt.Sprintf("%T", obs),
					"error", err,
				)
			}
		}(observer)
	}
}
