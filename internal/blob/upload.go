package blob

import (
	"context"
	"fmt"

	"github.com/pomerium/datasource/internal/httputil"
)

// UploadBundle uploads a bundle of data to blob storage.
func UploadBundle(ctx context.Context, urlstr string, bundle map[string]any) error {
	bucket, err := openBucket(ctx, urlstr)
	if err != nil {
		return fmt.Errorf("error opening bucket: %w", err)
	}
	defer bucket.Close()

	file, err := bucket.NewWriter(ctx, "bundle.zip", nil)
	if err != nil {
		return fmt.Errorf("error opening bucket file: %w", err)
	}

	err = httputil.EncodeBundle(file, bundle)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("error writing bundle to bucket file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("error closing bucket file: %w", err)
	}

	return nil
}
