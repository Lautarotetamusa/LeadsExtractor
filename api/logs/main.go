package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnectionSettings struct {
    User string
    Pass string
    Host string
    Port int16
}

func MongoConnect(c context.Context, s *MongoConnectionSettings) *mongo.Client {
    ctx, cancel := context.WithTimeout(c, 15*time.Second)
    defer cancel()

    connectionStr := fmt.Sprintf("mongodb://%s:%s@%s:%d", s.User, s.Pass, s.Host, s.Port)

    clientOptions := options.Client().ApplyURI(connectionStr)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal("imposible conectar con mongodb", err)
    }
    
    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("error pinging mongodb", err)
    }

    slog.Info("connected to mongodb")
    return client
}

type MongoWriter struct {
    collection *mongo.Collection
}

func NewMongoWriter(c *mongo.Collection) *MongoWriter {
    return &MongoWriter{
        collection: c,
    }
}

func (mw *MongoWriter) Write(p []byte) (n int, err error) {
    var logEntry any
    if err := json.Unmarshal(p, &logEntry); err != nil {
        return 0, nil
    }

    _, err = mw.collection.InsertOne(context.TODO(), logEntry)
    if err != nil {
        return
    }
    return len(p), nil
}
