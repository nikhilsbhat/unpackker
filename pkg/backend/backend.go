// Package backend is responsible for storing and fetching asset from or to the designated location.
package backend

import (
	"context"
	"fmt"

	"gocloud.dev/blob"
)

// Store helps one to specify where the artifact should be tranported, default to local.
type Store struct {
	// // Type of backend where the artifact to be stored.
	// Type string `json:"type" yaml:"type"`
	// Name the cloud of the bucket for artifact.
	Cloud string `json:"cloud" yaml:"cloud"`
	// Name the Bucket in appropriate cloud for artifact store.
	Bucket string `json:"bucket" yaml:"bucket"`
	// Path where the asset has to be fetched to.
	TargetPath string `json:"targetpath" yaml:"targetpath"`
	sourcePath string
	blobConn   *blob.Bucket
	ctx        context.Context
}

// New retunrns new config of Store.
func New() *Store {
	return &Store{}
}

// Backend initializes backend for Unpackker to store the packed asset.
func (b *Store) Backend() error {
	if err := b.validate(); err != nil {
		return err
	}
	if b.Cloud == "fs" {
		path := b.TargetPath
		b = New()
		b.Cloud = "fs"
		b.TargetPath = path
		return nil
	}
	return fmt.Errorf("currently we support only filesystem")
}

// StoreAsset stores the packed asset at specified location
func (b *Store) StoreAsset() error {
	if b.Cloud == "fs" {
		return nil
	}
	return nil
}

// FetchAsset stores the packed asset at specified location
func (b *Store) FetchAsset() error {
	if b.Cloud == "fs" {
		return nil
	}
	return nil
}

func (b *Store) validate() error {
	if len(b.Cloud) == 0 {
		b.Cloud = "fs"
	}
	return nil
}
