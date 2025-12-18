namespace go wcd

struct EmptyReq{
}

enum RuleStageType{
    Unk = -1
    Testing = 0
    Production = 1
}
enum RuleStageGroupEnum{
    ProdOnly = 0
    TestingOnly = 1
    ProdPrior = 2
    TestingPrior = 3
}

struct WcdParseReq{
    1: string url
    2: string html
    3: optional bool reparse // 是否强制重新解析，不走缓存
    4: optional RuleStageGroupEnum rule_stage_group // 规则组，默认为ProdOnly
}

struct WcdParseResp{
    1: i32 code
    2: string msg

    3: string url
    4: string text
    5: list<string> images // 正文图片url列表
    6: string readable_html // 正文html，给阅读器
    7: string title
    8: string author
    9: string content_source
    10: string pub_time
    11: string model_input_str
    12: string model_result_str
    13: bool worthless // 是否无意义
    14: i32 worth_type
    15: string wcd_request_id
    19: optional string raw_html // 原始html
    21: optional ArticleAuthorMeta author_meta // 作者信息
    22: optional string site_icon // 网站图标
    23: optional string description // 网页描述
    24: optional string surface_image // 封面图
}

struct AtomicText{
	1:string      text
	2:i32         position_id
	3:string      xpath // xpath
	7:bool        immutable
	8:list<string>tags// 原始标签
	9:bool        tail
	10:i32        segment_id
	11:map<string, string> attrs
}

struct SentenceMeta{
    1: string url
    2: string table_html
    3: list<i32> captions
    4: list<string> caption_txts
    5: i32 type
    6: string title_type
    7: string title_index
    8: string pure_title
}
struct AtomicSentence{
    1: string text
    2: list<AtomicText> atoms
	3: i32 segment_id
	4: list<string> tags// 原始标签
	5: SentenceMeta meta
}

struct TextParseLabelSentence{
    1: string text
    2: list<AtomicText> atoms
	3: i32 segment_id
    4: string label
}

struct SegmentReq{
    1: string html
    2: string url
    3: optional RuleStageGroupEnum rule_stage_group
}

struct ArticleAuthorMeta{
    1: string name
    2: string profile_url
    3: string description
    4: string uid
}

struct ArticleMeta{
    1: string url
    2: string title
    3: string publish_time
    4: string author
    5: string content_source
    6: optional ArticleAuthorMeta author_meta
    7: optional string site_icon
    8: optional string description
    9: optional string surface_image
}

struct SegmentResp{
    1: i32 code
    2: string msg

    3: list<AtomicSentence> sentences
    4: string html
    5: string operation_id
    6: ArticleMeta article_meta
    7: map<string, string> images_with_position_id
}

struct DistillReq{
    3: list<TextParseLabelSentence> sentences
    4: string html
    5: string url
    6: ArticleMeta article_meta
}
struct DistillResp{
    1: i32 code
    2: string msg

    3: list<i32> sentence_ids
    4: string html // 去噪后html
    5: string text // 解析纯文本
    6: list<string> images // 正文图片url
    7: bool worthless // 是否无意义
    8: i32 worth_type
}


struct BaseParseReq{
    3: string url // 网页：网页链接；pdf则是pdf的oss链接
    4: optional string html
    5: optional string file_name

    6: optional bool with_raw_html // 是否返回原始html
    7: optional RuleStageGroupEnum rule_stage_group // 规则组，默认为ProdOnly
    8: optional bool skip_cache // 解析时是否强制跳过缓存
    9: optional bool save_crawl_html // 解析后是否保存抓取的html
}

service WcdService{
    // 基础解析。pdf：切句。web：抓取、切句、去噪
    WcdParseResp BaseParse(1: BaseParseReq req)(api.post="/base-parse")

    // 切分+去噪
    WcdParseResp WcdParse(1: WcdParseReq req)(api.post="/wcd/parse")
    // 切分网页
    SegmentResp Segment(1: SegmentReq req)(api.post="/wcd/segment")
    // 去噪
    DistillResp Distill(1: DistillReq req)(api.post="/wcd/distill")
}