package core

import (
	"context"

	"github.com/NIROOZbx/notification-engine/engine/notification/models"
)

type Strategy interface {
	SkipBillingCheck() bool
	SkipOptOut() bool
	ResolveContact(ctx context.Context, repo Repository, ic *ingestContext) (*Contact, error)
}

type ingestContext struct {
	workspaceID string
	envID       string
	payload     *models.TriggerPayload
	template    *Template
	contact     *Contact
	channelKey  string
	ch          *TemplateChannel
	strategy    Strategy
}
type normalStrategy struct{}


func (s *normalStrategy) SkipBillingCheck() bool { return false }
func (s *normalStrategy) SkipOptOut() bool       { return false }
func (s *normalStrategy) ResolveContact(ctx context.Context, repo Repository, ic *ingestContext) (*Contact, error) {
	return repo.GetContactByExternalUserAndChannel(ctx, ic.workspaceID, ic.envID, ic.payload.ExternalUserID, ic.ch.Channel)
}

type systemStrategy struct {
	recipientEmail string
}

func (s *systemStrategy) SkipBillingCheck() bool { return true }
func (s *systemStrategy) SkipOptOut() bool       { return true }
func (s *systemStrategy) ResolveContact(_ context.Context, _ Repository, _ *ingestContext) (*Contact, error) {
	return &Contact{ContactValue: s.recipientEmail}, nil
}