package email

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NIROOZbx/notification-engine/engine/notification/provider"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/rs/zerolog"
)

type sesProvider struct {
	log    zerolog.Logger
	client *http.Client
}

func NewSESProvider(log zerolog.Logger, httpClient *http.Client) *sesProvider {
	return &sesProvider{
		log:    log,
		client: httpClient,
	}
}

func (s *sesProvider) Name() string { return "ses" }

func (s *sesProvider) Channel() string { return "email" }

func (s *sesProvider) RequiredFields() []string {
	return []string{"access_key_id", "secret_access_key", "region", "from_email"}
}

func (s *sesProvider) RequiredContent() []string { return []string{"subject", "body"} }

func (s *sesProvider) Send(ctx context.Context, msg provider.Message, configs map[string]string) error {
	accessKey := configs["access_key_id"]
	secretKey := configs["secret_access_key"]
	region := configs["region"]
	fromEmail := configs["from_email"]
	fromName := configs["from_name"]
	subject := msg.Content["subject"].(string)
	body := msg.Content["body"].(string)

	s.log.Debug().
		Str("region", region).
		Str("from", fromEmail).
		Str("access_key_prefix", accessKey[:min(4, len(accessKey))]).
		Msg("attempting SES send with config")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithHTTPClient(s.client),
	)
	if err != nil {
		return fmt.Errorf("failed to load aws config: %w", err)
	}

	svc := ses.NewFromConfig(cfg)

	source := fromEmail
	if fromName != "" {
		source = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	}

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{msg.To},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body),
				},
				Text: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body),
				},
			},
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(source),
	}

	result, err := svc.SendEmail(ctx, input)
	if err != nil {
		s.log.Error().
			Err(err).
			Str("to", msg.To).
			Msg("SES SendEmail failed")
		return fmt.Errorf("ses error: %w", err)
	}

	s.log.Info().
		Str("message_id", *result.MessageId).
		Str("to", msg.To).
		Msg("Email successfully sent via SES (v2)")

	return nil
}
