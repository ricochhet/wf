package drivers

import (
	"bytes"
	"context"

	"github.com/ricochhet/pkg/errutil"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoConnector struct {
	ctx    context.Context //nolint:containedctx // contained ctx is fine for single context application.
	client *mongo.Client
}

// NewMongoConnector creates a new Mongo client connection with the given URI.
func NewMongoConnector(ctx context.Context, uri string) (*MongoConnector, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &MongoConnector{ctx: ctx, client: client}, nil
}

// Disconnect disconnects the client.
func (c MongoConnector) Disconnect() error {
	if err := c.client.Disconnect(c.ctx); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// Get gets a database collection and returns a mongo.Collection.
func (c MongoConnector) Get(database, collection string) ([]byte, *mongo.Collection, error) {
	data := c.client.Database(database).Collection(collection)

	bytes, err := get(c.ctx, data)
	if err != nil {
		return nil, nil, errutil.WithFrame(err)
	}

	return bytes, data, nil
}

// Set sets a collection to the given data.
func (c MongoConnector) Set(collection *mongo.Collection, data []byte) error {
	if err := set(c.ctx, collection, data); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// get converts a mongo.Collection into a byte slice.
func get(ctx context.Context, collection *mongo.Collection) ([]byte, error) {
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, errutil.New("collection.Find", err)
	}
	defer cursor.Close(ctx)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		return nil, errutil.New("cursor.All", err)
	}

	// Convert []bson.D to []any for generic marshaling
	docs := make([]any, len(results))
	for i := range results {
		docs[i] = results[i]
	}

	return marshalExtJSONArray(docs)
}

// set drops the existing collection, and sets it to the given data.
func set(ctx context.Context, collection *mongo.Collection, data []byte) error {
	if err := collection.Drop(ctx); err != nil {
		return errutil.New("collection.Drop", err)
	}

	var rawDocs []bson.Raw
	if err := bson.UnmarshalExtJSON(data, false, &rawDocs); err == nil {
		docs := make([]any, len(rawDocs))
		for i, raw := range rawDocs {
			docs[i] = raw
		}

		_, err := collection.InsertMany(ctx, docs)
		if err != nil {
			return errutil.New("collection.InsertMany", err)
		}

		return nil
	}

	var singleDoc bson.Raw
	if err := bson.UnmarshalExtJSON(data, false, &singleDoc); err != nil {
		return errutil.New("bson.UnmarshalExtJSON", err)
	}

	_, err := collection.InsertOne(ctx, singleDoc)
	if err != nil {
		return errutil.New("collection.InsertOne", err)
	}

	return nil
}

// marshalExtJSONArray places all documents into a JSON array, and returns a byte slice.
func marshalExtJSONArray(docs []any) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')

	for cur, doc := range docs {
		if cur > 0 {
			buf.WriteByte(',')
		}

		b, err := bson.MarshalExtJSON(doc, false, false)
		if err != nil {
			return nil, errutil.WithFrame(err)
		}

		buf.Write(b)
	}

	buf.WriteByte(']')

	return buf.Bytes(), nil
}
