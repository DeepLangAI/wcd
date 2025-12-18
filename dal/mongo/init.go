package mongo

import (
	"context"
	"fmt"

	"github.com/DeepLangAI/wcd/conf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	wcdDb *mongo.Database
)

func initParseDb(ctx context.Context) {
	mongoConfig := conf.GetConfig().Mongo
	clientOptions := options.Client().ApplyURI(mongoConfig.Addr)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(fmt.Sprintf("initialize mongodb Connect failed, err: %v", err))
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		panic(fmt.Sprintf("initialize mongodb Ping failed, err: %v", err))
	}
	wcdDb = client.Database(mongoConfig.DbName)
}

func Init(ctx context.Context) {
	initParseDb(ctx)
}
