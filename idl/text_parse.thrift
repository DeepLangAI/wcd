namespace go text_parse

include "wcd.thrift"

struct AtomicTxt {
	1:optional string   txt
	2:i32      position_id
	3:string   x // xpath
}

struct LabelPosition{
    1: list<AtomicTxt> atoms
}

// 简化版接口定义
struct LabelInfoReq{
    1: string txt
    2: string label
    3: wcd.SentenceMeta meta
}

struct LabelInfoResp{
    2: string label
    3: wcd.SentenceMeta meta
}

struct TextParseReq{
    1: wcd.ArticleMeta article_meta
    2: list<LabelInfoReq> labels
}

struct TextParseResp{
    1: i32 code
    2: string msg
    3: list<LabelInfoResp> labels
    4: wcd.ArticleMeta article_meta
}