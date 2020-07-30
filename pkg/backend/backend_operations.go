package backend

import (
	"fmt"
	"strings"

	"gocloud.dev/blob"

	// Blank import is being made so that this library can connect to multiple cloud  if required.
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

var (
	bucketPrefix   = map[string]string{"aws": "s3://", "gcp": "gs://", "azure": "azblob://", "fs": "file://"}
	cloudSUpported = []string{"gcp", "aws", "azure", "fs"}
)

// ConnectBucket will establish connection to the bucket of appropriate cloud so that the content can be accessed.
func (b *Store) ConnectBucket() (*blob.Bucket, error) {
	if err := b.ValidateBucketURL(); err != nil {
		return nil, err
	}

	bucket, err := blob.OpenBucket(b.ctx, b.Bucket)
	if err != nil {
		return nil, fmt.Errorf("could not open bucket: %v", err)
	}
	fmt.Println(bucket)
	return bucket, nil
}

// ReadBucket reads through the configured bucket.
func (b *Store) ReadBucket() error {
	if err := b.ValidateBucketURL(); err != nil {
		return err
	}
	r, err := b.blobConn.NewReader(b.ctx, "foo.txt", nil)
	if err != nil {
		return err
	}
	defer r.Close()
	return nil
}

// CopyBucketContent either downloads or uploads the asset to specified cloud.
func (b *Store) CopyBucketContent() error {
	if err := b.ValidateBucketURL(); err != nil {
		return err
	}

	if err := b.blobConn.Copy(b.ctx, b.TargetPath, b.sourcePath, &blob.CopyOptions{}); err != nil {
		return err
	}
	defer b.blobConn.Close()
	return nil
}

// ValidateBucketURL validates the bucket passed to Unpackker.
func (b *Store) ValidateBucketURL() error {
	if !b.validateCloud() {
		return fmt.Errorf("at the moment unpackker does not support backend for the cloud %s configured", b.Cloud)
	}

	if !b.validateURL(b.Bucket) {
		b.createBucketURL()
	}
	return nil
}

func (b *Store) validateCloud() bool {
	for _, cloud := range cloudSUpported {
		if b.Cloud == cloud {
			return true
		}
	}
	return false
}

func (b *Store) createBucketURL() {
	b.Bucket = (bucketPrefix[b.Cloud] + b.Bucket)
}

func (b *Store) validateURL(url string) bool {
	if strings.HasPrefix(b.Bucket, bucketPrefix[b.Bucket]) {
		return true
	}
	return false
}
