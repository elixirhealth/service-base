package storage

import (
	"cloud.google.com/go/datastore"
	"context"
	"os"
	"os/exec"
	"syscall"
	"github.com/drausin/libri/libri/common/errors"
	"log"
)

const (
	datastoreEmulatorHostEnv = "DATASTORE_EMULATOR_HOST"
	datastoreEmulatorAddr    = "localhost:2002"
	dummyDatastoreProject    = "dummy-datastore-test"
)

// DatastoreClient is a interface wrapper for a *datastore.Client to facilitate mocking in tests.
type DatastoreClient interface {
	Put(key *datastore.Key, value interface{}) (*datastore.Key, error)
	Get(key *datastore.Key, dest interface{}) error
	Delete(keys []*datastore.Key) error
	Count(ctx context.Context, q *datastore.Query) (int, error)
	Run(ctx context.Context, q *datastore.Query) *datastore.Iterator
}

// DatastoreClientImpl implements DatastoreClient.
type DatastoreClientImpl struct {
	Inner *datastore.Client
}

// Get wraps datastore.Client.Get(...)
func (c *DatastoreClientImpl) Get(key *datastore.Key, dest interface{}) error {
	return c.Inner.Get(context.Background(), key, dest)
}

// Put wraps datastore.Client.Put(...)
func (c *DatastoreClientImpl) Put(key *datastore.Key, value interface{}) (*datastore.Key, error) {
	return c.Inner.Put(context.Background(), key, value)
}

// Delete wraps datastore.Client.Delete(...)
func (c *DatastoreClientImpl) Delete(keys []*datastore.Key) error {
	return c.Inner.DeleteMulti(context.Background(), keys)
}

// Count wraps datastore.Client.Count(...)
func (c *DatastoreClientImpl) Count(ctx context.Context, q *datastore.Query) (int, error) {
	return c.Inner.Count(ctx, q)
}

// Run wraps datastore.Client.Run(...)
func (c *DatastoreClientImpl) Run(ctx context.Context, q *datastore.Query) *datastore.Iterator {
	return c.Inner.Run(ctx, q)
}

// DatastoreIterator is an interface wrapper for a *datastore.Iterator to facilitate mocking in
// tests.
type DatastoreIterator interface {
	Init(iter *datastore.Iterator)
	Next(dst interface{}) (*datastore.Key, error)
}

// DatastoreIteratorImpl implements DatastoreIterator.
type DatastoreIteratorImpl struct {
	inner *datastore.Iterator
}

// Init initializes the iterator with the given one.
func (i *DatastoreIteratorImpl) Init(iter *datastore.Iterator) {
	i.inner = iter
}

// Next wraps datastore.Iterator.Next(...)
func (i *DatastoreIteratorImpl) Next(dst interface{}) (*datastore.Key, error) {
	return i.inner.Next(dst)
}

// StartDatastoreEmulator starts the DataStore emulator.
func StartDatastoreEmulator(dataDir string) *os.Process {
	// nolint: gas
	cmd := exec.Command("gcloud", "beta", "emulators", "datastore", "start",
		"--no-store-on-disk",
		"--host-port", datastoreEmulatorAddr,
		"--project", dummyDatastoreProject,
		"--data-dir", dataDir,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	errors.MaybePanic(err)
	err = os.Setenv(datastoreEmulatorHostEnv, datastoreEmulatorAddr)
	errors.MaybePanic(err)
	return cmd.Process
}

// StopDatastoreEmulator stops the DataStore emulator.
func StopDatastoreEmulator(process *os.Process) {
	pgid, err := syscall.Getpgid(process.Pid)
	errors.MaybePanic(err)
	err = syscall.Kill(-pgid, syscall.SIGKILL)
	errors.MaybePanic(err)
	log.Printf("stopped child processes under pid %d\n", pgid)
}
