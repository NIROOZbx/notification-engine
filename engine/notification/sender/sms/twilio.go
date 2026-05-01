package sms

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
	"github.com/rs/zerolog"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type twilioProvider struct {
	log    zerolog.Logger
	client *http.Client
}

func NewTwilioProvider(log zerolog.Logger, httpClient *http.Client) *twilioProvider {
	return &twilioProvider{
		log:    log,
		client: httpClient,
	}
}

func (p *twilioProvider) Name() string { return "twilio" }

func (p *twilioProvider) Channel() string { return "sms" }

func (p *twilioProvider) RequiredFields() []string {
	return []string{"account_sid", "auth_token", "from"}
}

func (p *twilioProvider) RequiredContent() []string {
	return []string{"body"}
}

func (p *twilioProvider) Send(ctx context.Context, msg provider.Message, config map[string]string) (string, error) {
	accountSid := config["account_sid"]
	authToken := config["auth_token"]
	from := config["from"]

	p.log.Debug().
		Str("account_sid", accountSid).
		Str("from", from).
		Msg("twilio credentials")

	body, ok := msg.Content["body"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid body content")
	}

	twilioClient := &client.Client{
		Credentials: client.NewCredentials(accountSid, authToken),
		HTTPClient:  p.client,
	}
	twilioClient.SetAccountSid(accountSid)

	restClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
		Client:   twilioClient,
	})

	params := &twilioApi.CreateMessageParams{}
	params.SetTo(msg.To)
	params.SetFrom(from)
	params.SetBody(body)

	resp, err := restClient.Api.CreateMessage(params)
	if err != nil {
		p.log.Error().
			Err(err).
			Str("to", msg.To).
			Msg("Twilio send request failed")
		return "", fmt.Errorf("twilio error: %w", err)
	}

	if resp.ErrorCode != nil {
		errMsg := ""
		if resp.ErrorMessage != nil {
			errMsg = *resp.ErrorMessage
		}
		p.log.Error().
			Interface("errorCode", resp.ErrorCode).
			Str("errorMessage", errMsg).
			Str("to", msg.To).
			Msg("Twilio API returned an error")
		return "", fmt.Errorf("twilio api error: %s (code %v)", errMsg, resp.ErrorCode)
	}

	p.log.Info().
		Str("sid", *resp.Sid).
		Str("to", msg.To).
		Msg("SMS successfully sent via Twilio")

	if resp.Sid == nil {
		p.log.Warn().Str("to", msg.To).Msg("twilio returned no SID — delivery tracking unavailable")
		return "", nil
	}
	return *resp.Sid, nil

}
