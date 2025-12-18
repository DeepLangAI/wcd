package manage

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/DeepLangAI/wcd/biz/model/wcd_manage"
	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/dal/mongo"
	"github.com/DeepLangAI/wcd/utils"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type RuleManageService struct {
	ctx context.Context
}

func NewRuleManageService(ctx context.Context) *RuleManageService {
	return &RuleManageService{ctx: ctx}
}

func (r *RuleManageService) getOrderFunc(orderBy string) utils.CmpFunc[*wcd_manage.SiteRuleData] {
	defaultCmp := func(a, b *wcd_manage.SiteRuleData) int {
		if a.Host < b.Host {
			return -1
		} else if a.Host == b.Host {
			return 0
		} else {
			return 1
		}
	}
	if orderBy == "" {
		// host
		return defaultCmp
	}
	if orderBy == "-update_time" {
		return func(a, b *wcd_manage.SiteRuleData) int {
			if a.UpdateTime > b.UpdateTime {
				return -1
			} else if a.UpdateTime == b.UpdateTime {
				return 0
			} else {
				return 1
			}
		}
	}
	if orderBy == "-create_time" {
		return func(a, b *wcd_manage.SiteRuleData) int {
			if a.CreateTime > b.CreateTime {
				return -1
			} else if a.CreateTime == b.CreateTime {
				return 0
			} else {
				return 1
			}
		}
	}
	return defaultCmp
}

func (r *RuleManageService) List(req wcd_manage.SiteRuleListReq) ([]*wcd_manage.SiteRuleData, error) {
	dal := mongo.SiteRuleModelDal
	data := []*wcd_manage.SiteRuleData{}
	rules, err := dal.ListAll(r.ctx)
	if err != nil {
		hlog.CtxErrorf(r.ctx, "list site rule failed, err: %v", err)
		return nil, err
	}
	if req.Query != "" {
		rulesFiltered := utils.Filter(rules, func(rule mongo.SiteRuleModel) bool {
			return strings.Contains(rule.HostName, req.Query) || strings.Contains(rule.Host, req.Query)
		})
		if len(rulesFiltered) == 0 {
			rulesFiltered = utils.Filter(rules, func(rule mongo.SiteRuleModel) bool {
				return rule.Match(req.Query)
			})
		}

		rules = rulesFiltered
	}
	siteEditing := map[string]int{}
	data = utils.Map(rules, func(rule mongo.SiteRuleModel) *wcd_manage.SiteRuleData {
		if rule.Stage == consts.RuleStageTesting {
			siteEditing[rule.Host] = 1
		}
		return rule.ToThrift()
	})

	utils.Sort(data, func(a, b *wcd_manage.SiteRuleData) int {
		// 先按有没有编辑中的状态排序
		if siteEditing[a.Host] != 0 && siteEditing[b.Host] != 0 {
			return 0
		} else if siteEditing[a.Host] != 0 {
			return -1
		} else if siteEditing[b.Host] != 0 {
			return 1
		} else {
			return 0
		}
	}, r.getOrderFunc(req.GetOrderBy()))

	return data, nil
}

func (r *RuleManageService) Detail(req wcd_manage.SiteRuleDetailReq) (*wcd_manage.SiteRuleData, error) {
	dal := mongo.SiteRuleModelDal
	rule, err := dal.FindOne(r.ctx, req.Host, consts.RuleStage(req.Stage))
	if err != nil {
		hlog.CtxErrorf(r.ctx, "find site rule failed, host: %s, err: %v", req.Host, err)
		return nil, err
	}
	return rule.ToThrift(), nil
}

func (r *RuleManageService) NewTestingRule(req wcd_manage.CreateSiteRuleReq) (*wcd_manage.SiteRuleData, error) {
	dal := mongo.SiteRuleModelDal
	// 1. 先判断是否存在测试规则
	testingExists := dal.Exists(r.ctx, req.Host, consts.RuleStageTesting)
	if testingExists {
		return nil, errors.New("testing rule already exists")
	}
	// 2. 创建测试规则
	rule := &mongo.SiteRuleModel{
		Host:       req.Host,
		HostName:   req.Name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Status:     consts.StatusValid,
		Stage:      consts.RuleStageTesting,
	}
	err := dal.UpsertMany(r.ctx, []*mongo.SiteRuleModel{rule})
	if err != nil {
		hlog.CtxErrorf(r.ctx, "create site rule failed, host: %s, err: %v", req.Host, err)
		return nil, err
	}
	return rule.ToThrift(), nil
}

func (r *RuleManageService) ProdToTest(req wcd_manage.ExportProdRuleToTestReq) (*wcd_manage.SiteRuleData, error) {
	dal := mongo.SiteRuleModelDal
	// 1. 先判断是否存在测试规则
	testingExists := dal.Exists(r.ctx, req.Host, consts.RuleStageTesting)
	if testingExists {
		return nil, errors.New("testing rule already exists")
	}

	prodRule, err := dal.FindOne(r.ctx, req.Host, consts.RuleStageProd)
	if err != nil {
		return nil, errors.New("prod rule not found")
	}
	// 2. 创建测试规则
	rule := &mongo.SiteRuleModel{}
	utils.DumpAndLoad(r.ctx, prodRule, rule)
	rule.Status = consts.StatusValid
	rule.Stage = consts.RuleStageTesting
	err = dal.UpsertMany(r.ctx, []*mongo.SiteRuleModel{rule})
	if err != nil {
		hlog.CtxErrorf(r.ctx, "create site rule failed, host: %s, err: %v", req.Host, err)
		return nil, err
	}
	return rule.ToThrift(), nil
}

func (r *RuleManageService) Publish(req wcd_manage.PublishSiteRuleReq) error {
	dal := mongo.SiteRuleModelDal
	testingRules, err := dal.FindManyTestingRules(r.ctx, req.Host)
	if err != nil {
		hlog.CtxErrorf(r.ctx, "find site rule failed, host: %s, err: %v", req.Host, err)
		return err
	}
	if len(testingRules) == 0 {
		return errors.New("testing rule not found")
	}

	prodRules := []*mongo.SiteRuleModel{}
	for _, rule := range testingRules {
		rule.Stage = consts.RuleStageProd
		prodRules = append(prodRules, &rule)
	}
	err = dal.UpsertMany(r.ctx, prodRules)
	if err != nil {
		hlog.CtxErrorf(r.ctx, "create site rule failed, host: %s, err: %v", req.Host, err)
		return err
	}

	err = dal.DeleteMany(r.ctx, testingRules)
	if err != nil {
		hlog.CtxErrorf(r.ctx, "delete site rule failed, host: %s, err: %v", req.Host, err)
		return err
	}
	return nil
}

func (r *RuleManageService) Delete(req wcd_manage.DeleteSiteRuleReq) error {
	dal := mongo.SiteRuleModelDal
	exists := dal.Exists(r.ctx, req.Host, consts.RuleStage(req.Stage))
	if !exists {
		hlog.CtxErrorf(r.ctx, "site rule not found, host: %s, stage: %s", req.Host, req.Stage)
		return errors.New("site rule not found")
	}
	err := dal.DeleteMany(r.ctx, []mongo.SiteRuleModel{
		{Host: req.Host, Stage: consts.RuleStage(req.Stage)},
	})
	if err != nil {
		hlog.CtxErrorf(r.ctx, "delete site rule failed, host: %s, err: %v", req.Host, err)
		return err
	}
	return nil
}

func (r *RuleManageService) Update(req wcd_manage.SiteRuleData) error {
	if consts.RuleStage(req.Stage) == consts.RuleStageProd {
		// 不允许更新线上规则
		return errors.New("prod rule can not be updated")
	}
	// 1. 先判断是否存在测试规则
	dal := mongo.SiteRuleModelDal
	oldModel, err := dal.FindOne(r.ctx, req.Host, consts.RuleStageTesting)
	if err != nil {
		hlog.CtxErrorf(r.ctx, "find site rule failed, host: %s, err: %v", req.Host, err)
		return err
	}
	if oldModel == nil {
		hlog.CtxErrorf(r.ctx, "site rule not found, host: %s, stage: %s", req.Host, req.Stage)
		return errors.New("site rule not found")
	}
	// host不允许修改
	//if req.Host != "" {
	//	oldModel.Host = req.Host
	//}
	oldModel.HostName = req.Name
	oldModel.Bodies = req.Bodies
	oldModel.Noises = req.Noises
	oldModel.Author = req.Author
	oldModel.PubTime = req.PubTime
	oldModel.Title = req.Title
	oldModel.ReservedNodes = req.ReservedNodes
	oldModel.NoSemanticDenoise = req.NoSemanticDenoise
	oldModel.NeedBrowserCrawl = req.NeedBrowserCrawl
	oldModel.BodyUseRuleOnly = req.BodyUseRuleOnly
	// 2. 更新测试规则
	err = dal.UpsertMany(r.ctx, []*mongo.SiteRuleModel{
		oldModel,
	})
	if err != nil {
		hlog.CtxErrorf(r.ctx, "update site rule failed, host: %s, err: %v", req.Host, err)
		return err
	}
	return nil
}

func (r *RuleManageService) Export(req wcd_manage.EmptyReq) ([]*wcd_manage.SiteRuleData, error) {
	dal := mongo.SiteRuleModelDal
	models, err := dal.FindMany(r.ctx, consts.RuleStageProd)
	if err != nil {
		hlog.CtxErrorf(r.ctx, "find site rule failed, err: %v", err)
		return nil, err
	}
	rules := utils.Map(models, func(model mongo.SiteRuleModel) *wcd_manage.SiteRuleData {
		return model.ToThrift()
	})
	return rules, nil
}

func (r *RuleManageService) ImportFromCtx(c *app.RequestContext) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	fhandle, err := file.Open()
	if err != nil {
		return err
	}
	defer fhandle.Close()
	// 读取文件内容
	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(fhandle)
	if err != nil {
		return err
	}
	data := []*wcd_manage.SiteRuleData{}
	err = sonic.Unmarshal(buffer.Bytes(), &data)
	if err != nil {
		return err
	}
	return r.Import(data)
}

func (r *RuleManageService) Import(data []*wcd_manage.SiteRuleData) error {
	dal := mongo.SiteRuleModelDal
	models := utils.Map(data, func(rule *wcd_manage.SiteRuleData) *mongo.SiteRuleModel {
		model := &mongo.SiteRuleModel{}
		return model.FromThrift(rule)
	})
	err := dal.SaveMany(r.ctx, models)
	if err != nil {
		hlog.CtxErrorf(r.ctx, "update site rule failed, err: %v", err)
		return err
	}
	return nil
}
