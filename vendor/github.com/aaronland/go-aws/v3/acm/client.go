package acm

import (
	"context"

	"github.com/aaronland/go-aws/v3/auth"
	aws_acm "github.com/aws/aws-sdk-go-v2/service/acm"
)

func NewClient(ctx context.Context, uri string) (*aws_acm.Client, error) {

	cfg, err := auth.NewConfig(ctx, uri)

	if err != nil {
		return nil, err
	}

	return aws_acm.NewFromConfig(cfg), nil
}
