package main

import (
	"log/slog"
	"os"

	"github.com/osbuild/logging/pkg/cloudwatch"
)

func main() {
	aws_region := os.Getenv("AWS_REGION")
	aws_key := os.Getenv("AWS_ACCESS_KEY_ID")
	aws_secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	aws_session := os.Getenv("AWS_SESSION")
	aws_log_group := os.Getenv("CLOUDWATCH_GROUP")
	aws_log_stream := os.Getenv("CLOUDWATCH_STREAM")

	cw, err := cloudwatch.New(cloudwatch.CloudwatchConfig{
		Level:        slog.LevelDebug,
		AddSource:    true,
		AWSRegion:    aws_region,
		AWSKey:       aws_key,
		AWSSecret:    aws_secret,
		AWSSession:   aws_session,
		AWSLogGroup:  aws_log_group,
		AWSLogStream: aws_log_stream,
	})
	if err != nil {
		panic(err)
	}
	defer cw.Close()

	log := slog.New(cw)
	log.Debug("a testing message", "k1", "v1")
}
