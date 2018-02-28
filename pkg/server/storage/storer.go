package storage

// Type indicates the storage backend type.
type Type int

const (
	// Unspecified indicates when the storage type is not specified (and thus should take the
	// default value).
	Unspecified Type = iota

	// Memory indicates an ephemeral, in-memory (and thus not highly available) storage. This
	// storage layer should generally only be used during testing and not in production.
	Memory

	// DataStore indicates a (highly available) storage backed by GCP DataStore.
	DataStore

	// Postgres indicates a storage backed by a Postgres DB.
	Postgres
)

func (t Type) String() string {
	switch t {
	case Memory:
		return "Memory"
	case DataStore:
		return "DataStore"
	case Postgres:
		return "Postgres"
	default:
		return "Unspecified"
	}
}
