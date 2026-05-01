package email

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
	"github.com/rs/zerolog"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type sendGridProvider struct {
	log    zerolog.Logger
	client *http.Client 
}

func NewSendGridProvider(log zerolog.Logger, httpClient *http.Client) *sendGridProvider {
	return &sendGridProvider{
		log:    log,
		client: httpClient,
	}
}

func (s *sendGridProvider) Name() string { return "sendgrid" }

func (s *sendGridProvider) Channel() string { return "email" }

func (s *sendGridProvider) RequiredFields() []string { return []string{"api_key", "from_email"} }

func (s *sendGridProvider) RequiredContent() []string { return []string{"subject", "body"} }

func (s *sendGridProvider) Send(ctx context.Context, msg provider.Message, config map[string]string)(string, error) {
	apiKey := config["api_key"]
	fromEmail := config["from_email"]
	fromName := config["from_name"]
	subject := msg.Content["subject"].(string)
	body := msg.Content["body"].(string)

	from := mail.NewEmail(fromName, fromEmail)
	to := mail.NewEmail("", msg.To)
	message := mail.NewSingleEmail(from, subject, to, body, body)

	client := sendgrid.NewSendClient(apiKey)

	response, err := client.SendWithContext(ctx,message)
	if err != nil {
		return "",fmt.Errorf("sendgrid error: %w", err)
	}

	if response.StatusCode >= 400 {
        s.log.Error().
            Int("status", response.StatusCode).
            Str("body", response.Body).
            Str("to", msg.To).
            Msg("SendGrid API rejected the request")
        return "",fmt.Errorf("sendgrid api error: status %d, body: %s", response.StatusCode, response.Body)
    }
	messageID := response.Headers["X-Message-Id"][0]

    s.log.Info().
        Str("to", msg.To).
        Int("status", response.StatusCode).
        Msg("Email successfully handed off to SendGrid")

	return messageID,nil

}
