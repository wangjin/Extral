package app

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Service struct {
	ctx context.Context
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.ctx = ctx
	return nil
}

func (s *Service) ServiceShutdown() error {
	return nil
}
