package backend

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nikhilsbhat/unpackker/pkg/helper"
)

// bucketOpts implements various operations on cloud
type bucketOpts interface {
	getClient() error
	connectBucket() error
	fetchAsset() error
	storeAsset() error
}

// Operations related to cloud gcp

// connectBucket establishes connection to GCS bucket
func (c *gcpCredentials) connectBucket(bucket string, object string) error {
	if c.gcpClient == nil {
		return fmt.Errorf("unable to connect to bucket, gcp client not found")
	}

	c.gcpBlobConn = c.gcpClient.Bucket(bucket).Object(object)
	return nil
}

// fetchAsset makes sure that the asset is fetched from specified GCS bucket onto the specified location.
func (c *gcpCredentials) fetchAsset(path string) error {
	if c.gcpBlobConn == nil {
		return fmt.Errorf("unable to fetch asset, connection to bucket was not established")
	}
	if helper.Statfile(path) {
		return fmt.Errorf("asset already fetched in the specified path: %s", path)
	}
	defer c.gcpClient.Close()

	rc, err := c.gcpBlobConn.NewReader(c.ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	asset, err := helper.CreateFile(path)
	if err != nil {
		return err
	}

	if _, err := io.Copy(asset, rc); err != nil {
		return err
	}

	if err := rc.Close(); err != nil {
		return fmt.Errorf("Reader.Close: %v", err)
	}

	if err := asset.Close(); err != nil {
		return err
	}
	return nil
}

// storeAsset makes sure that that the asset is stored to specified GCS bucket.
func (c *gcpCredentials) storeAsset(path string) error {
	if c.gcpBlobConn == nil {
		return fmt.Errorf("unable to store asset, connection to bucket was not established")
	}
	if !helper.Statfile(path) {
		return fmt.Errorf("unable to find asset under specified path: %s", path)
	}
	defer c.gcpClient.Close()

	asset, err := helper.OpenFile(path)
	if err != nil {
		return err
	}

	wc := c.gcpBlobConn.NewWriter(c.ctx)

	if _, err := io.Copy(wc, asset); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	if err := asset.Close(); err != nil {
		return err
	}
	return nil
}

// Operations related to cloud aws

// connectBucket establishes connection to S3 bucket
func (c *awsCredentials) connectBucket(bucket string, object string) error {
	if c.awsClient == nil {
		return fmt.Errorf("unable to connect to bucket, aws client not found")
	}

	c.awsBlobConn = s3.New(c.awsClient)
	return nil
}

// fetchAsset makes sure that the asset is fetched from specified aws S3 bucket onto the specified location.
func (c *awsCredentials) fetchAsset(path string) error {
	if c.awsBlobConn == nil {
		return fmt.Errorf("unable to fetch asset, connection to bucket was not established")
	}
	return nil
}

// storeAsset makes sure that the asset is stored to specified S3 bucket.
func (c *awsCredentials) storeAsset(path string) error {
	return nil
}
