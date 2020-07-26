package backend

import "fmt"

// Store helps one to specify where the artifact should be tranported, default to local.
type Store struct {
	// Type of backend where the artifact to be stored.
	Type string `json:"type" yaml:"type"`
	// Name the cloud of the bucket for artifact.
	Cloud string `json:"cloud" yaml:"cloud"`
	// Name the Bucket in appropriate cloud for artifact store.
	Bucket string `json:"bucket" yaml:"bucket"`
	// Path where the asset has to be fetched to.
	Path string `json:"path" yaml:"path"`
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
	if b.Type == "fs" {
		path := b.Path
		b = New()
		b.Type = "fs"
		b.Path = path
		return nil
	}
	return fmt.Errorf("Currently we support only filesystem")
}

// StoreAsset stores the packed asset at specified location
func (b *Store) StoreAsset() error {
	return nil
}

// FetchAsset stores the packed asset at specified location
func (b *Store) FetchAsset() {

}

func (b *Store) validate() error {
	if len(b.Type) == 0 {
		b.Type = "fs"
	}
	return nil
}
