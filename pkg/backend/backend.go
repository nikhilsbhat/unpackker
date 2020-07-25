package backend

// Store helps one to specify where the artifact should be tranported, default to local.
type Store struct {
	// Type of backend where the artifact to be stored.
	Type string
	// Name the cloud of the bucket for artifact.
	Cloud string
	// Name the Bucket in appropriate cloud for artifact store.
	Bucket string
}

// Backend initializes backend for Unpackker to store the packed asset.
func (b *Store) Backend() {

	if b.Type == "fs" {

	}
}
