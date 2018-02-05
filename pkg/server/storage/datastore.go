package storage

import (
	"cloud.google.com/go/datastore"
	"context"
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
