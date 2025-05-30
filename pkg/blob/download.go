package blob

import (
	"context"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog/log"
)

// DownloadState downloads state data from blob storage.
func DownloadState(ctx context.Context, urlstr string, callback func(src io.Reader) error) error {
	log.Ctx(ctx).Debug().Msg("downloading state")

	bucket, err := openBucket(ctx, urlstr)
	if err != nil {
		return fmt.Errorf("error opening bucket: %w", err)
	}
	defer bucket.Close()

	file, err := bucket.NewReader(ctx, "state.zst", nil)
	if err != nil {
		return fmt.Errorf("error opening bucket file: %w", err)
	}

	zr, err := zstd.NewReader(file)
	if err != nil {
		return fmt.Errorf("error creating zstd reader: %w", err)
	}

	err = callback(zr)
	zr.Close()
	if err != nil {
		return fmt.Errorf("error reading state: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("error closing bucket file: %w", err)
	}

	return nil
}
