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

type DatastoreClientImpl struct {
	inner *datastore.Client
}

func (c *DatastoreClientImpl) Get(key *datastore.Key, dest interface{}) error {
	return c.inner.Get(context.Background(), key, dest)
}

func (c *DatastoreClientImpl) Put(key *datastore.Key, value interface{}) (*datastore.Key, error) {
	return c.inner.Put(context.Background(), key, value)
}

func (c *DatastoreClientImpl) Delete(keys []*datastore.Key) error {
	return c.inner.DeleteMulti(context.Background(), keys)
}

func (c *DatastoreClientImpl) Count(ctx context.Context, q *datastore.Query) (int, error) {
	return c.inner.Count(ctx, q)
}

func (c *DatastoreClientImpl) Run(ctx context.Context, q *datastore.Query) *datastore.Iterator {
	return c.inner.Run(ctx, q)
}

