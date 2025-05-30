// Package blob contains functions for working with blob storage.
package blob

import (
	"context"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob" // support azure blob storage
	_ "gocloud.dev/blob/fileblob"  // support file blob storage
	_ "gocloud.dev/blob/gcsblob"   // support gcs blob storage
	_ "gocloud.dev/blob/memblob"   // support mem blob storage
	_ "gocloud.dev/blob/s3blob"    // support s3 blob storage
)

func openBucket(ctx context.Context, urlstr string) (*blob.Bucket, error) {
	return blob.OpenBucket(ctx, urlstr)
}
