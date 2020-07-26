package backend

// Store helps one to specify where the artifact should be tranported, default to local.
type Store struct {
	// Type of backend where the artifact to be stored.
	Type string `json:"type" yaml:"type"`
	// Name the cloud of the bucket for artifact.
	Cloud string `json:"cloud" yaml:"cloud"`
	// Name the Bucket in appropriate cloud for artifact store.
	Bucket string `json:"bucket" yaml:"bucket"`
}

// Backend initializes backend for Unpackker to store the packed asset.
func (b *Store) Backend() {

	if b.Type == "fs" {

	}
}

// StoreAsset stores the packed asset at specified location
func (b *Store) StoreAsset() {

}
