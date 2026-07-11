// Package awsqueue publishes JSON messages to an SQS queue when one is configured, and no-ops otherwise. It
// matches the platform's self-service stream contract (ADR-073): the Environment Composition provisions an
// SSE-enabled queue and publishes its URL into the <svc>-resources ConfigMap (e.g. EVENTS_QUEUE_URL). Creds
// come from EKS Pod Identity; AWS_REGION is set by the manifest. Noop mode keeps local dev / tests simple —
// an unconfigured queue is not an error, the event is just dropped.
package awsqueue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Publisher sends a message body to a queue.
type Publisher interface {
	Send(ctx context.Context, body string) error
	Backend() string
}

// Open returns an SQS publisher when queueURL is non-empty, else a no-op publisher.
func Open(ctx context.Context, queueURL string) (Publisher, error) {
	if queueURL == "" {
		return noop{}, nil
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &sqsPublisher{cl: sqs.NewFromConfig(cfg), url: queueURL}, nil
}

type sqsPublisher struct {
	cl  *sqs.Client
	url string
}

func (p *sqsPublisher) Backend() string { return "sqs" }

func (p *sqsPublisher) Send(ctx context.Context, body string) error {
	_, err := p.cl.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.url),
		MessageBody: aws.String(body),
	})
	return err
}

type noop struct{}

func (noop) Backend() string                        { return "noop" }
func (noop) Send(_ context.Context, _ string) error { return nil }
