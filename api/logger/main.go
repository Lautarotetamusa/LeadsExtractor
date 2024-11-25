package logger

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoWriter struct {
    client *mongo.Client
}

func (mw *MongoWriter) SetClient(c *mongo.Client) {
    mw.client = c
}

func (mw *MongoWriter) Write(p []byte) (n int, err error) {
    c := mw.client.Database("mydatabase").Collection("log")

    var logEntry any
    if err := json.Unmarshal(p, &logEntry); err != nil {
        return 0, nil
    }

    _, err = c.InsertOne(context.TODO(), logEntry)
    if err != nil {
        return
    }
    return len(p), nil
}
