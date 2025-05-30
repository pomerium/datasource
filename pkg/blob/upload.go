package blob

import (
	"context"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog/log"

	"github.com/pomerium/datasource/internal/httputil"
)

// UploadBundle uploads a bundle of data to blob storage.
func UploadBundle(ctx context.Context, urlstr string, bundle map[string]any) error {
	log.Ctx(ctx).Debug().Msg("uploading bundle")
	return upload(ctx, urlstr, "bundle.zip", func(w io.Writer) error {
		err := httputil.EncodeBundle(w, bundle)
		if err != nil {
			return fmt.Errorf("error writing bundle to bucket file: %w", err)
		}

		return nil
	})
}

// UploadState uploads state data to blob storage.
func UploadState(ctx context.Context, urlstr string, callback func(dst io.Writer) error) error {
	log.Ctx(ctx).Debug().Msg("uploading state")
	return upload(ctx, urlstr, "state.zst", func(w io.Writer) error {
		zw, err := zstd.NewWriter(w)
		if err != nil {
			return fmt.Errorf("error creating zstd writer for bucket file: %w", err)
		}

		err = callback(zw)
		if err != nil {
			return fmt.Errorf("error writing state to bucket file: %w", err)
		}

		err = zw.Close()
		if err != nil {
			return fmt.Errorf("error closing zstd writer: %w", err)
		}

		return nil
	})
}

func upload(ctx context.Context, urlstr, fileName string, callback func(w io.Writer) error) error {
	bucket, err := openBucket(ctx, urlstr)
	if err != nil {
		return fmt.Errorf("error opening bucket: %w", err)
	}
	defer bucket.Close()

	file, err := bucket.NewWriter(ctx, fileName, nil)
	if err != nil {
		return fmt.Errorf("error opening bucket file: %w", err)
	}

	err = callback(file)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("error writing bucket file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("error closing bucket file: %w", err)
	}

	return nil
}
