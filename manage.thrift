namespace go wcd_manage
include "../idl/wcd.thrift"

struct EmptyReq{
}


// 查看各站点规则列表
struct SiteRuleListReq{
    1: string query
    2: wcd.RuleStageGroupEnum stage_group
    3: optional string order_by
}
struct SiteRuleListResp{
    1: i32 code
    2: string msg
    3: list<SiteRuleData> data
}

struct SiteRuleData{
    1: string id
    2: string host
    3: string name
    4: list<string> bodies
    5: list<string> noises

    6: string title
    7: string pub_time
    8: string author
    9: wcd.RuleStageType stage
    10: list<string> reserved_nodes
    11: bool no_semantic_denoise
    12: bool need_browser_crawl
    13: bool body_use_rule_only

    14: string create_time
    15: string update_time
}

// 查看各站点规则详情
struct SiteRuleDetailReq{
    1: string host
    2: wcd.RuleStageType stage
}

struct SiteRuleDetailResp{
    1: i32 code
    2: string msg
    3: SiteRuleData data
}

// 新建规则，默认处于测试状态
struct CreateSiteRuleReq{
    1: string host
    2: string name
}
struct CreateSiteRuleResp{
    1: i32 code
    2: string msg
    3: SiteRuleData data
}

// 更新规则
struct UpdateSiteRuleResp{
    1: i32 code
    2: string msg
    3: SiteRuleData data
}

// 从正式规则导出一个测试规则
struct ExportProdRuleToTestReq{
    1: string host
}
struct ExportProdRuleToTestResp{
    1: i32 code
    2: string msg
    3: SiteRuleData data
}

// 发布测试状态的规则至正式
struct PublishSiteRuleReq{
    1: list<string> host
}
struct PublishSiteRuleResp{
    1: i32 code
    2: string msg
    3: SiteRuleData data
}

// 删除规则
struct DeleteSiteRuleReq{
    1: string host
    2: wcd.RuleStageType stage
}
struct DeleteSiteRuleResp{
    1: i32 code
    2: string msg
}


// 导出所有站点规则
struct ExportSiteRulesResp{
    1: i32 code
    2: string msg
    3: list<SiteRuleData> data // 当前导出任务id
}

// 导入站点规则
struct ImportSiteRulesResp{
    1: i32 code
    2: string msg
}


service RuleFactory{
    // 查看各站点规则列表
    SiteRuleListResp SiteRuleList(1: SiteRuleListReq req)(
        api.post="/api/v1/site_rule/list"
    )
    // 查看各站点规则详情
    SiteRuleDetailResp SiteRuleDetail(1: SiteRuleDetailReq req)(
        api.post="/api/v1/site_rule/detail"
    )
    // 新建规则，默认处于测试状态
    CreateSiteRuleResp CreateSiteRule(1: CreateSiteRuleReq req)(
        api.post="/api/v1/site_rule/testing/new"
    )
    // 更新规则，只能更新测试规则
    UpdateSiteRuleResp UpdateSiteRule(1: SiteRuleData req)(
        api.post="/api/v1/site_rule/testing/update"
    )
    // 从正式规则导出一个测试规则
    ExportProdRuleToTestResp CreateSiteRuleFromProd(1: ExportProdRuleToTestReq req)(
        api.post="/api/v1/site_rule/prod/to_test"
    )
    // 发布测试状态的规则至正式
    PublishSiteRuleResp PublishSiteRule(1: PublishSiteRuleReq req)(
        api.post="/api/v1/site_rule/testing/publish"
    )
    // 删除规则
    DeleteSiteRuleResp DeleteSiteRule(1: DeleteSiteRuleReq req)(
        api.post="/api/v1/site_rule/delete"
    )
    // 导出所有站点规则
    ExportSiteRulesResp ExportSiteRules(1: EmptyReq req)(
        api.get="/api/v1/site_rule/export"
    )
    // 导出所有站点规则
    ImportSiteRulesResp ImportSiteRules(1: EmptyReq req)(
        api.post="/api/v1/site_rule/import"
    )
}