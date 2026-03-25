package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/pepnova-9/old-message-dispatcher/domain" // imports package domain
)

// DB represents the database port for the worker.
type DB interface {
	FindPendingCampaign(ctx context.Context) (*domain.Campaign, error)
	MarkCampaignCompleted(ctx context.Context, id uuid.UUID) error
	SaveSmsDispatch(ctx context.Context, dispatch *domain.SmsDispatch) error
}
