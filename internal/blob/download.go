package blob

import (
	"context"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog/log"
)

// DownloadState downloads state data from blob storage.
func DownloadState(ctx context.Context, urlstr string) ([]byte, error) {
	log.Ctx(ctx).Debug().Msg("downloading state")

	bucket, err := openBucket(ctx, urlstr)
	if err != nil {
		return nil, fmt.Errorf("error opening bucket: %w", err)
	}
	defer bucket.Close()

	file, err := bucket.NewReader(ctx, "state.zst", nil)
	if err != nil {
		return nil, fmt.Errorf("error opening bucket file: %w", err)
	}

	zr, err := zstd.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("error creating zstd reader: %w", err)
	}

	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("error reading zstd reader: %w", err)
	}

	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing bucket file: %w", err)
	}

	return data, nil
}
