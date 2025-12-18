package mongo

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/DeepLangAI/wcd/biz/model/wcd"

	"github.com/DeepLangAI/wcd/biz/model/wcd_manage"
	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TableNameSiteRule = "site_rule"

type SiteRuleModel struct {
	Host     string `bson:"host"`
	HostName string `bson:"host_name"`

	Bodies            []string `bson:"bodies"`
	Noises            []string `bson:"noises"`
	Author            string   `bson:"author"`
	PubTime           string   `bson:"pub_time"`
	Title             string   `bson:"title"`
	ReservedNodes     []string `bson:"reserved_nodes"`
	NoSemanticDenoise bool     `bson:"no_semantic_denoise"` // 无须按语义去噪
	NeedBrowserCrawl  bool     `bson:"need_browser_crawl"`  // 需要浏览器爬取
	BodyUseRuleOnly   bool     `bson:"body_use_rule_only"`  // 仅使用规则提取正文

	CreateTime time.Time        `bson:"create_time"`
	UpdateTime time.Time        `bson:"update_time"`
	Status     consts.DbStatus  `bson:"status"`
	Stage      consts.RuleStage `bson:"stage"`
}

func (s *SiteRuleModel) cleanUrlHost(url string) string {
	prefixs := []string{"https://", "http://", "www."}
	for _, prefix := range prefixs {
		if strings.HasPrefix(url, prefix) {
			url = strings.TrimPrefix(url, prefix)
			return s.cleanUrlHost(url)
		}
	}
	return url
}

func (s *SiteRuleModel) Match(url string) bool {
	if url == "" {
		return false
	}
	hostName := s.cleanUrlHost(url)
	matched := false
	var err error
	if utils.Any([]string{"*", "+"}, func(token string) bool {
		return strings.Contains(s.Host, token)
	}) {
		matched, err = regexp.MatchString(s.Host, hostName)
		if err != nil {
			return false
		}
		if matched {
			return true
		}
	}
	if matched = strings.HasPrefix(hostName, s.Host); matched {
		return true
	}
	return matched
}

func (s *SiteRuleModel) ToThrift() *wcd_manage.SiteRuleData {
	return &wcd_manage.SiteRuleData{
		ID:                "",
		Host:              s.Host,
		Name:              s.HostName,
		Bodies:            s.Bodies,
		Noises:            s.Noises,
		Title:             s.Title,
		PubTime:           s.PubTime,
		Author:            s.Author,
		Stage:             wcd.RuleStageType(s.Stage),
		ReservedNodes:     s.ReservedNodes,
		NoSemanticDenoise: s.NoSemanticDenoise,
		NeedBrowserCrawl:  s.NeedBrowserCrawl,
		BodyUseRuleOnly:   s.BodyUseRuleOnly,
		CreateTime:        s.CreateTime.Format(time.DateTime),
		UpdateTime:        s.UpdateTime.Format(time.DateTime),
	}
}

func (s *SiteRuleModel) FromThrift(data *wcd_manage.SiteRuleData) *SiteRuleModel {
	model := &SiteRuleModel{
		Host:              data.Host,
		HostName:          data.Name,
		Bodies:            data.Bodies,
		Noises:            data.Noises,
		Title:             data.Title,
		PubTime:           data.PubTime,
		Author:            data.Author,
		ReservedNodes:     data.ReservedNodes,
		NoSemanticDenoise: data.NoSemanticDenoise,
		NeedBrowserCrawl:  data.NeedBrowserCrawl,
		BodyUseRuleOnly:   data.BodyUseRuleOnly,
		Stage:             consts.RuleStage(data.Stage),
	}
	if createTime, err := time.Parse(time.DateTime, data.CreateTime); err == nil {
		model.CreateTime = createTime
	} else {
		model.CreateTime = time.Now()
	}
	if updateTime, err := time.Parse(time.DateTime, data.UpdateTime); err == nil {
		model.UpdateTime = updateTime
	} else {
		model.UpdateTime = time.Now()
	}
	return model
}

var SiteRuleModelDal *siteRuleModelDal

type siteRuleModelDal struct{}

func (s *siteRuleModelDal) FindManyByStageGroup(ctx context.Context, stageGroup wcd.RuleStageGroupEnum) ([]*SiteRuleModel, error) {
	filter := bson.D{
		{"status", consts.StatusValid},
	}
	cursor, err := wcdDb.Collection(TableNameSiteRule).Find(ctx, filter)
	if err != nil {
		hlog.CtxErrorf(ctx, "find site rule error: %v", err)
		return nil, err
	}
	var rules []*SiteRuleModel
	err = cursor.All(ctx, &rules)
	if err != nil {
		hlog.CtxErrorf(ctx, "decode site rule error: %v", err)
		return nil, err
	}

	finalRules := []*SiteRuleModel{}
	testingRules := map[string]*SiteRuleModel{}
	prodRules := map[string]*SiteRuleModel{}
	for _, rule := range rules {
		if rule.Stage == consts.RuleStageTesting {
			testingRules[rule.Host] = rule
		} else {
			prodRules[rule.Host] = rule
		}
	}

	if stageGroup == wcd.RuleStageGroupEnum_TestingOnly {
		for _, rule := range testingRules {
			finalRules = append(finalRules, rule)
		}
	} else if stageGroup == wcd.RuleStageGroupEnum_ProdOnly {
		for _, rule := range prodRules {
			finalRules = append(finalRules, rule)
		}
	} else if stageGroup == wcd.RuleStageGroupEnum_TestingPrior {
		for _, rule := range testingRules {
			finalRules = append(finalRules, rule)
		}

		for _, rule := range prodRules {
			if testingRules[rule.Host] == nil {
				finalRules = append(finalRules, rule)
			}
		}
	} else if stageGroup == wcd.RuleStageGroupEnum_ProdPrior {
		for _, rule := range prodRules {
			finalRules = append(finalRules, rule)
		}

		for _, rule := range testingRules {
			if prodRules[rule.Host] == nil {
				finalRules = append(finalRules, rule)
			}
		}
	}

	return finalRules, nil
}
func (s *siteRuleModelDal) FindMany(ctx context.Context, ruleStage consts.RuleStage) ([]SiteRuleModel, error) {
	filter := bson.D{
		{"stage", ruleStage},
		{"status", consts.StatusValid},
	}
	cursor, err := wcdDb.Collection(TableNameSiteRule).Find(ctx, filter)
	if err != nil {
		hlog.CtxErrorf(ctx, "find site rule error: %v", err)
		return nil, err
	}
	var rules []SiteRuleModel
	err = cursor.All(ctx, &rules)
	if err != nil {
		hlog.CtxErrorf(ctx, "decode site rule error: %v", err)
		return nil, err
	}
	return rules, err
}

func (s *siteRuleModelDal) FindManyTestingRules(ctx context.Context, hosts []string) ([]SiteRuleModel, error) {
	var models []SiteRuleModel
	findOptions := options.Find()
	filter := bson.D{
		{Key: "status", Value: consts.StatusValid},
		{Key: "stage", Value: consts.RuleStageTesting},
	}
	if len(hosts) > 0 {
		filter = append(filter, bson.E{Key: "host", Value: bson.D{{Key: "$in", Value: hosts}}})
	}
	cursor, err := wcdDb.Collection(TableNameSiteRule).Find(ctx, filter, findOptions)
	if err != nil {
		hlog.CtxErrorf(ctx, "find site rule failed: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &models)
	if err != nil {
		hlog.CtxErrorf(ctx, "decode site rule failed: %v", err)
		return nil, err
	}
	return models, nil
}

func (s *siteRuleModelDal) ListAll(ctx context.Context) ([]SiteRuleModel, error) {
	filter := bson.D{
		//{"stage", ruleStage},
		{"status", consts.StatusValid},
	}
	cursor, err := wcdDb.Collection(TableNameSiteRule).Find(ctx, filter)
	if err != nil {
		hlog.CtxErrorf(ctx, "find site rule error: %v", err)
		return nil, err
	}
	var rules []SiteRuleModel
	err = cursor.All(ctx, &rules)
	if err != nil {
		hlog.CtxErrorf(ctx, "decode site rule error: %v", err)
		return nil, err
	}
	return rules, err
}

func (s *siteRuleModelDal) FindOne(ctx context.Context, host string, ruleStage consts.RuleStage) (*SiteRuleModel, error) {
	filter := bson.D{
		{"host", host},
		{"stage", ruleStage},
		{"status", consts.StatusValid},
	}
	result := wcdDb.Collection(TableNameSiteRule).FindOne(ctx, filter)
	if result.Err() != nil {
		// 如果没有找到，也会有错误
		hlog.CtxErrorf(ctx, "find site rule error: %v", result.Err())
		return nil, result.Err()
	}
	rule := &SiteRuleModel{}
	err := result.Decode(rule)
	if err != nil {
		hlog.CtxErrorf(ctx, "decode site rule error: %v", err)
		return nil, err
	}
	return rule, nil
}
func (s *siteRuleModelDal) Exists(ctx context.Context, host string, ruleStage consts.RuleStage) bool {
	rule, err := s.FindOne(ctx, host, ruleStage)
	if rule == nil || err != nil {
		return false
	}
	return true
}
func (s *siteRuleModelDal) SaveMany(ctx context.Context, models []*SiteRuleModel) error {
	// 也是upsert，但不会自动更新时间
	var operations []mongo.WriteModel
	for _, model := range models {
		filter := bson.D{
			{Key: "host", Value: model.Host},
			{Key: "stage", Value: model.Stage},
			{Key: "status", Value: consts.StatusValid},
		}
		update := bson.D{
			{Key: "$set", Value: model},
		}
		op := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		operations = append(operations, op)
	}
	if len(operations) == 0 {
		return nil
	}
	_, err := wcdDb.Collection(TableNameSiteRule).BulkWrite(ctx, operations)
	if err != nil {
		hlog.CtxErrorf(ctx, "bulkWrite failed, err: %v", err)
		return err
	}
	return nil
}
func (s *siteRuleModelDal) UpsertMany(ctx context.Context, models []*SiteRuleModel) error {
	var operations []mongo.WriteModel
	for _, model := range models {
		filter := bson.D{
			{Key: "host", Value: model.Host},
			{Key: "stage", Value: model.Stage},
			{Key: "status", Value: consts.StatusValid},
		}
		update := bson.D{
			{Key: "$set", Value: bson.D{
				//只需要更新的字段
				{Key: "host_name", Value: model.HostName},
				{Key: "bodies", Value: model.Bodies},
				{Key: "noises", Value: model.Noises},
				{Key: "author", Value: model.Author},
				{Key: "pub_time", Value: model.PubTime},
				{Key: "title", Value: model.Title},
				{Key: "reserved_nodes", Value: model.ReservedNodes},
				{Key: "no_semantic_denoise", Value: model.NoSemanticDenoise},
				{Key: "need_browser_crawl", Value: model.NeedBrowserCrawl},
				{Key: "body_use_rule_only", Value: model.BodyUseRuleOnly},

				{Key: "update_time", Value: time.Now()}, // 始终更新：更新时间
			}},
			{Key: "$setOnInsert", Value: bson.D{
				{Key: "create_time", Value: time.Now()}, // 仅在插入时设置
				//{Key: "id", Value: primitive.NewObjectID().Hex()},
			}},
		}
		op := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		operations = append(operations, op)
	}
	if len(operations) == 0 {
		return nil
	}
	_, err := wcdDb.Collection(TableNameSiteRule).BulkWrite(ctx, operations)
	if err != nil {
		hlog.CtxErrorf(ctx, "bulkWrite failed, err: %v", err)
		return err
	}
	return nil
}

func (s *siteRuleModelDal) DeleteMany(ctx context.Context, models []SiteRuleModel) error {
	var operations []mongo.WriteModel
	for _, model := range models {
		filter := bson.D{
			{Key: "host", Value: model.Host},
			{Key: "stage", Value: model.Stage},
			{Key: "status", Value: consts.StatusValid},
		}
		update := bson.D{
			{Key: "$set", Value: bson.D{
				//只需要更新的字段
				{Key: "status", Value: consts.StatusDeleted},
				{Key: "update_time", Value: time.Now()}, // 始终更新：更新时间
			}},
			{Key: "$setOnInsert", Value: bson.D{
				{Key: "create_time", Value: time.Now()}, // 仅在插入时设置
			}},
		}
		op := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		operations = append(operations, op)
	}
	if len(operations) == 0 {
		return nil
	}
	_, err := wcdDb.Collection(TableNameSiteRule).BulkWrite(ctx, operations)
	if err != nil {
		hlog.CtxErrorf(ctx, "bulkWrite failed, err: %v", err)
		return err
	}
	return nil
}
