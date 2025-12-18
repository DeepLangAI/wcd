package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/DeepLangAI/wcd/consts"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const TableNameCrawlImage = "crawl_image"

type CrawlImageModel struct {
	SrcImgUrl string `bson:"src_img_url"`
	OssImgUrl string `bson:"oss_img_url"`

	CreateTime time.Time       `bson:"create_time"`
	UpdateTime time.Time       `bson:"update_time"`
	Status     consts.DbStatus `bson:"status"`
}

var CrawlImageModelDal *crawlImageModelDal

type crawlImageModelDal struct{}

func (c *crawlImageModelDal) checkLegal(model CrawlImageModel) (legal bool, msg string) {
	if model.SrcImgUrl == "" {
		return false, "src img url is empty"
	}
	if model.OssImgUrl == "" {
		return false, "oss img url is empty"
	}
	return true, ""
}

func (c *crawlImageModelDal) SaveOne(ctx context.Context, model CrawlImageModel) error {
	if legal, msg := c.checkLegal(model); !legal {
		hlog.CtxErrorf(ctx, "before save one, check model illegal: %s", msg)
		return errors.New(msg)
	}
	one, err := wcdDb.Collection(TableNameCrawlImage).InsertOne(ctx, model)
	if err != nil {
		return err
	}
	if one.InsertedID == nil {
		return err
	}
	return err
}

func (c *crawlImageModelDal) UpsertMany(ctx context.Context, models []CrawlImageModel) error {
	hlog.CtxInfof(ctx, "trying to upsert %v crawled images", len(models))
	// 1. 校验所有模型合法性
	for i, model := range models {
		if legal, msg := c.checkLegal(model); !legal {
			hlog.CtxErrorf(ctx, "before save many, model[%d] illegal: %s", i, msg)
			return errors.New(msg)
		}
	}

	// 2. 构造批量写操作（upsert by src_img_url）
	now := time.Now()
	writeModels := make([]mongo.WriteModel, 0, len(models))
	for _, model := range models {
		// 过滤条件：根据 src_img_url 匹配
		filter := bson.M{"src_img_url": model.SrcImgUrl}

		// 更新操作：
		// - $set：更新通用字段（包括 update_time）
		// - $setOnInsert：仅在插入时设置 create_time（避免更新时覆盖原有 create_time）
		update := bson.M{
			"$set": bson.M{
				"oss_img_url": model.OssImgUrl,
				"update_time": now,
				"status":      model.Status,
			},
			"$setOnInsert": bson.M{
				"create_time": now, // 插入时设置创建时间
			},
		}

		// 构造 UpdateOne 操作（upsert: true）
		writeModels = append(writeModels, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	// 3. 执行批量写操作
	_, err := wcdDb.Collection(TableNameCrawlImage).BulkWrite(ctx, writeModels)
	if err != nil {
		hlog.CtxErrorf(ctx, "bulk write failed: %v", err)
		return err
	}

	// 可选：记录操作结果（如需要）
	hlog.CtxInfof(ctx, "bulk write %v crawled images success", len(models))
	return nil
}

func (c *crawlImageModelDal) FindBySrcImgUrls(ctx context.Context, imgUrls []string) ([]*CrawlImageModel, error) {
	hlog.CtxInfof(ctx, "trying to find %v crawled images by src img urls: %v", len(imgUrls), imgUrls)
	filter := bson.D{
		{Key: "status", Value: consts.StatusValid},
		{Key: "src_img_url", Value: bson.D{{Key: "$in", Value: imgUrls}}},
	}
	cur, err := wcdDb.Collection(TableNameCrawlImage).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var models []*CrawlImageModel
	for cur.Next(ctx) {
		var model CrawlImageModel
		if err := cur.Decode(&model); err != nil {
			hlog.CtxErrorf(ctx, "decode error: %v", err)
			return nil, err
		}
		models = append(models, &model)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	hlog.CtxInfof(ctx, "find crawled %v images by src img urls: %v", len(models), imgUrls)
	return models, nil
}
