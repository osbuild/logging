package cloudwatch

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cww "github.com/lzap/cloudwatchwriter2"
)

// CloudwatchHandler is a slog.Handler that sends logs to AWS CloudWatch.
type CloudwatchHandler struct {
	*slog.JSONHandler

	client *cww.CloudWatchWriter
}

// CloudwatchConfig is the configuration for the Cloudwatch output.
type CloudwatchConfig struct {
	// Level is the logging level for this output.
	Level slog.Leveler

	// AddSource is a flag to add source to the log record.
	AddSource bool

	// AWSRegion is the AWS region.
	AWSRegion string

	// AWSKey is the AWS access key.
	AWSKey string

	// AWSSecret is the AWS secret key.
	AWSSecret string

	// AWSSession is an optional AWS session token.
	AWSSession string

	// AWSLogGroup is the AWS CloudWatch log group.
	AWSLogGroup string

	// AWSLogStream is the AWS CloudWatch log stream.
	AWSLogStream string
}

// New creates a new CloudwatchHandler.
func New(config CloudwatchConfig) (*CloudwatchHandler, error) {
	// configure AWS CloudWatch client
	options := cloudwatchlogs.Options{
		Region:      config.AWSRegion,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(config.AWSKey, config.AWSSecret, config.AWSSession)),
	}
	client, err := cww.NewWithClient(cloudwatchlogs.New(options), 500*time.Millisecond, config.AWSLogGroup, config.AWSLogStream)
	if err != nil {
		return nil, fmt.Errorf("cannot create new cloudwatch client: %w", err)
	}

	// configure slog handler
	opts := &slog.HandlerOptions{
		Level:       config.Level,
		AddSource:   config.AddSource,
		ReplaceAttr: replaceAttr,
	}
	cwh := &CloudwatchHandler{
		JSONHandler: slog.NewJSONHandler(client, opts),
		client:      client,
	}

	return cwh, nil
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	// timestamp is added by CloudWatch library automatically
	if groups == nil && a.Key == slog.TimeKey {
		return slog.Attr{}
	}

	return a
}

// Flush flushes all pending payloads to the CloudWatch client.
func (h *CloudwatchHandler) Close() {
	h.client.Close()
}
