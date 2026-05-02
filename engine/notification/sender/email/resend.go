package email

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
	"github.com/resend/resend-go/v3"
	"github.com/rs/zerolog"
)

type resendProvider struct {
	log    zerolog.Logger
	client *http.Client
}

func NewResendProvider(log zerolog.Logger, httpClient *http.Client) *resendProvider {
	return &resendProvider{
		log:    log,
		client: httpClient,
	}
}

func (r *resendProvider) Name() string { return "resend" }

func (r *resendProvider) Channel() string { return "email" }

func (r *resendProvider) RequiredFields() []string { return []string{"api_key", "from_email"} }

func (r *resendProvider) RequiredContent() []string { return []string{"subject", "body"} }

func (r *resendProvider) Send(ctx context.Context, msg provider.Message, config map[string]string) (string, error) {
	apiKey := config["api_key"]
	fromEmail := config["from_email"]
	fromName := config["from_name"]
	subject := msg.Content["subject"].(string)
	body := msg.Content["body"].(string)

	from := fromEmail
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{msg.To},
		Subject: subject,
		Html:    body,
	}

	response, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		r.log.Error().Err(err).Str("to", msg.To).Msg("Resend API rejected the request")
		return "", fmt.Errorf("failed to send email via resend: %v", err)
	}

	r.log.Info().
		Str("to", msg.To).
		Str("message_id", response.Id).
		Msg("Email successfully handed off to Resend")

	return response.Id, nil
}
