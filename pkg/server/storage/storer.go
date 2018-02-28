package storage

import "go.uber.org/zap/zapcore"

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

var (
	// DefaultStorage is the default storage type.
	DefaultStorage = Memory
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


// Parameters defines the parameters of the Storer.
type Parameters struct {
	Type               Type
}

// NewDefaultParameters returns a *Parameters object with default values.
func NewDefaultParameters() *Parameters {
	return &Parameters{
		Type:               DefaultStorage,
	}
}

// MarshalLogObject writes the parameters to the given object encoder.
func (p *Parameters) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	oe.AddString(logStorageType, p.Type.String())
	return nil
}


