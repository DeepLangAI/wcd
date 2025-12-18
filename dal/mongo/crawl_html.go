package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/DeepLangAI/wcd/consts"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TableNameCrawlHtml = "crawl_html"

type CrawlHtmlModel struct {
	Url  string `bson:"url"`
	Html string `bson:"html"`

	CreateTime time.Time       `bson:"create_time"`
	UpdateTime time.Time       `bson:"update_time"`
	Status     consts.DbStatus `bson:"status"`
}

var CrawlHtmlModelDal *crawlHtmlModelDal

type crawlHtmlModelDal struct{}

func (c *crawlHtmlModelDal) checkLegal(model CrawlHtmlModel) (legal bool, msg string) {
	if model.Url == "" {
		return false, "url is empty"
	}
	if model.Html == "" {
		return false, "html is empty"
	}
	return true, ""
}

func (c *crawlHtmlModelDal) SaveOne(ctx context.Context, model CrawlHtmlModel) error {
	if legal, msg := c.checkLegal(model); !legal {
		hlog.CtxErrorf(ctx, "before save one, check model illegal: %s", msg)
		return errors.New(msg)
	}
	one, err := wcdDb.Collection(TableNameCrawlHtml).InsertOne(ctx, model)
	if err != nil {
		return err
	}
	if one.InsertedID == nil {
		return err
	}
	return err
}

func (c *crawlHtmlModelDal) FindByUrl(ctx context.Context, url string, expireDuration time.Duration) (*CrawlHtmlModel, error) {
	hlog.CtxInfof(ctx, "trying to find cached html by url: %s", url)
	findOptions := options.FindOne()
	findOptions.Sort = bson.D{{Key: "create_time", Value: -1}} // 按照创建时间倒序

	filter := bson.D{
		{Key: "status", Value: consts.StatusValid},
		{Key: "url", Value: url},
		{Key: "create_time", Value: bson.M{
			"$gte": time.Now().Add(-1 * expireDuration), // 有时效性，只查询最近24小时的数据
		}},
	}
	one := wcdDb.Collection(TableNameCrawlHtml).FindOne(ctx, filter, findOptions)
	if err := one.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, one.Err()
	}

	var model *CrawlHtmlModel
	if err := one.Decode(&model); err != nil { // 解码失败
		hlog.CtxErrorf(ctx, "decode error: %v", err)
		return nil, err
	}
	hlog.CtxInfof(ctx, "find cached html by url: %s", url)
	return model, nil
}
