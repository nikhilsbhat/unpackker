// Package backend is responsible for storing and fetching asset from or to the designated location.
package backend

import (
	"context"
	"fmt"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nikhilsbhat/neuron/cli/ui"
)

// Store helps one to specify where the artifact should be tranported, default to local.
type Store struct {
	// Name of the asset which has to be either uploaded or downloaded.
	Name string `json:"name" yaml:"name"`
	// Cloud name of the bucket to which asset belongs to.
	Cloud string `json:"cloud" yaml:"cloud"`
	// Bucket name in appropriate cloud for asset store.
	Bucket string `json:"bucket" yaml:"bucket"`
	// Folder under which the asset has to be placed.
	Folder string `json:"folder" yaml:"folder"`
	// Path to the asset which has to be store on to cloud.
	Path string `json:"path" yaml:"path"`
	// TargetPath refers to path where the asset has to be fetched to.
	TargetPath string `json:"targetpath" yaml:"targetpath"`
	// Path to cloud credentail file, 'service-account.json' incase of gcp.
	CredentialPath string `json:"credspath" yaml:"credspath"`
	// CredentialType of for cloud config. Unpackker supports two type, default and file type.
	// It defaults to default config.
	CredentialType string `json:"credstype" yaml:"credstype"`
	// SkipRemoteCheck would skip feature which avoids pushing asset which is present at backend.
	SkipRemoteCheck bool `json:"skipremotecheck" yaml:"skipremotecheck"`
	// Region where the bucket resides.
	Region string `json:"region" yaml:"region"`
	// Metadata of the asset that would be stored.
	MetaData   map[string]string `json:"metadata" yaml:"metadata"`
	sourcePath string
	gcpCreds   *gcpCredentials
	awsCreds   *awsCredentials
	// azureCreds can be added when unpackker supports azure as backend.
}

type gcpCredentials struct {
	ctx         context.Context
	gcpClient   *storage.Client
	gcpBlobConn *storage.ObjectHandle
}

type awsCredentials struct {
	awsClient   *session.Session
	awsBlobConn *s3.S3
}

// New returns new config of Store.
func New() *Store {
	return &Store{}
}

// newGCPCreds returns new instance of gcpCredentials.
func newGCPCreds() *gcpCredentials {
	return &gcpCredentials{}
}

// newAWSCreds returns new instance of awsCredentials.
func newAWSCreds() *awsCredentials {
	return &awsCredentials{}
}

// InitBackend initializes backend for Unpackker to store or retrieve the packed asset.
func (b *Store) InitBackend() error {
	if err := b.validate(); err != nil {
		return err
	}
	if b.Cloud == "fs" {
		return nil
	}
	if err := b.getClient(); err != nil {
		return err
	}
	return nil
}

// StoreAsset stores the packed asset at specified location.
// Make sure that InitBackend is invoked before calling this.
func (b *Store) StoreAsset() error {
	if b.Cloud == "fs" {
		return nil
	}

	b.Name = filepath.Join(b.Folder, b.Name)
	if err := b.connectBucket(); err != nil {
		return err
	}

	object, err := b.objectExists()
	if err != nil {
		if err == storage.ErrObjectNotExist {
			fmt.Println(ui.Info("Looks like object is not present we are creating one\n"))
		} else {
			return err
		}
	}
	if !b.SkipRemoteCheck {
		if object {
			return fmt.Errorf("asset with current version already exists at the backend as %s", b.Name)
		}
	}

	if err := b.storeAsset(); err != nil {
		return err
	}
	return nil
}

// FetchAsset stores the packed asset at specified location.
// Make sure that InitBackend is invoked before calling this.
func (b *Store) FetchAsset() error {
	if b.Cloud == "fs" {
		b.TargetPath = b.getTargetPath()
		return nil
	}
	if err := b.connectBucket(); err != nil {
		return err
	}

	b.TargetPath = b.getTargetPath()
	if err := b.fetchAsset(); err != nil {
		return err
	}
	return nil
}

func (b *Store) validate() error {
	if len(b.Cloud) == 0 {
		b.Cloud = "fs"
	}
	if (len(b.CredentialType) == 0) || (len(b.CredentialPath) == 0) || (len(b.CredentialPath) == 0) {
		b.CredentialType = "default"
	}
	return nil
}

func (b *Store) getTargetPath() string {
	return fmt.Sprintf("%s/%s", b.TargetPath, b.Name)
}
