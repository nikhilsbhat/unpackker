package backend

import (
	"fmt"
	"strings"
)

var (
	bucketPrefix   = map[string]string{"aws": "s3://", "gcp": "gs://", "azure": "azblob://", "fs": "file://"}
	cloudSUpported = []string{"gcp", "aws", "azure", "fs"}
)

// ConnectBucket will establish connection to the bucket of appropriate cloud so that the content can be accessed.
func (b *Store) connectBucket() error {
	if err := b.validateBucketURL(); err != nil {
		return err
	}

	if b.Cloud == "gcp" {
		if err := b.gcpCreds.connectBucket(b.Bucket, b.Name); err != nil {
			return nil
		}
	} else if b.Cloud == "aws" {
		if err := b.awsCreds.connectBucket(b.Bucket, b.Name); err != nil {
			return nil
		}
	} else if b.Cloud == "azure" {
		return nil
	}
	return nil
}

// // ReadBucket reads through the configured bucket.
// func (b *Store) readBucket() error {
// 	if err := b.validateBucketURL(); err != nil {
// 		return err
// 	}
// 	// r, err := b.blobConn.NewReader(b.ctx, "foo.txt", nil)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// defer r.Close()
// 	// return nil
// 	return nil
// }

// storeAsset downloads the asset from specified cloud.
func (b *Store) storeAsset() error {
	if err := b.validateBucketURL(); err != nil {
		return err
	}

	if b.Cloud == "gcp" {
		if err := b.gcpCreds.storeAsset(b.Path); err != nil {
			return nil
		}
	} else if b.Cloud == "aws" {
		if err := b.awsCreds.storeAsset(b.Path); err != nil {
			return nil
		}
	} else if b.Cloud == "azure" {
		return nil
	}

	return nil
}

// fetchAsset downloads the asset from specified cloud.
func (b *Store) fetchAsset() error {
	if err := b.validateBucketURL(); err != nil {
		return err
	}

	if b.Cloud == "gcp" {
		if err := b.gcpCreds.fetchAsset(b.TargetPath); err != nil {
			return nil
		}
	} else if b.Cloud == "aws" {
		if err := b.awsCreds.fetchAsset(b.TargetPath); err != nil {
			return nil
		}
	} else if b.Cloud == "azure" {
		return nil
	}

	return nil
}

// ValidateBucketURL validates the bucket passed to Unpackker.
func (b *Store) validateBucketURL() error {
	if !b.validateCloud() {
		return fmt.Errorf("at the moment unpackker does not support backend for the cloud %s configured", b.Cloud)
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
	if strings.HasPrefix(b.Bucket, bucketPrefix[b.Cloud]) {
		return true
	}
	return false
}
