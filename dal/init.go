package dal

import (
	"context"

	"github.com/DeepLangAI/wcd/dal/mongo"
)

func Init() {
	ctx := context.Background()
	mongo.Init(ctx)
}
