// Package blob contains functions for working with blob storage.
package blob

import (
	"context"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
	_ "gocloud.dev/blob/s3blob"
)

func openBucket(ctx context.Context, urlstr string) (*blob.Bucket, error) {
	return blob.OpenBucket(ctx, urlstr)
}
