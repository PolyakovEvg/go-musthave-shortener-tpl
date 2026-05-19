package audit

import (
	"PolyakovEvg/go-musthave-shortener-tpl/internal/middleware/logger"
	"fmt"
)

type AuditService struct {
	logger    *logger.Logger
	observers []Observer
}

func NewAuditService(logger *logger.Logger) *AuditService {
	return &AuditService{
		logger:    logger,
		observers: make([]Observer, 0),
	}
}

func (s *AuditService) Register(observer Observer) {
	s.observers = append(s.observers, observer)
	s.logger.Zap.Infow("Registered observer",
		"type", fmt.Sprintf("%T", observer),
	)
}

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
