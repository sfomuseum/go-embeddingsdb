package writer

import (
	"context"
	"fmt"
	"log/slog"

	aa_s3 "github.com/aaronland/gocloud/blob/s3"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"gocloud.dev/blob"
)

// NewWriterWithACL returns a new `blob.Writer` instance that has been configured with the relevant
// `blob.WriterOptions` to ensure that files written to S3 will be done using AWS ACL permissions
// defined in 'acl'.
func NewWriterWithACL(ctx context.Context, bucket *blob.Bucket, path string, str_acl string) (*blob.Writer, error) {

	logger := slog.Default()
	logger = logger.With("path", path)
	logger = logger.With("acl", str_acl)

	logger.Debug("Create new writer")

	before := func(asFunc func(any) bool) error {

		req := &transfermanager.UploadObjectInput{}
		ok := asFunc(&req)

		logger.Debug("Is s3.UploadObjectInput", "ok", ok)

		if ok {

			canned_acl, err := aa_s3.StringACLToTransferManagerObjectCannedACL(str_acl)

			if err != nil {
				logger.Error("Failed to derive canned ACL", "error", err)
				return fmt.Errorf("Failed to derive canned ACL, %w", err)
			}

			logger.Debug("Set ACL", "acl", canned_acl)
			req.ACL = canned_acl
		}

		return nil
	}

	wr_opts := &blob.WriterOptions{
		BeforeWrite: before,
	}

	wr, err := bucket.NewWriter(ctx, path, wr_opts)

	if err != nil {
		logger.Error("Failed to create new writer", "error", err)
		return nil, fmt.Errorf("Failed to create writer for %s, %w", path, err)
	}

	logger.Debug("Return new writer")
	return wr, nil
}
