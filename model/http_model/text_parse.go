package http_model

import (
	"github.com/DeepLangAI/wcd/biz/model/text_parse"
	"github.com/DeepLangAI/wcd/biz/model/wcd"
	"github.com/DeepLangAI/wcd/utils"
)

type PdfPosition struct {
	PositionId int32     `json:"position_id"`
	BBox       []float32 `json:"bbox"`
	PageNumber int32     `json:"page_number"`
	Shape      []float32 `json:"shape"`
}

type LabelPosition struct {
	Atoms       []*text_parse.AtomicTxt `json:"atoms,omitempty"`
	PdfPosition []*PdfPosition          `json:"pdf_position,omitempty"`
}

type TextParseInfo struct {
	LabelInfo
}

type LabelInfo struct {
	Txt          string            `json:"txt"`
	Position     *LabelPosition    `json:"position"`
	Tags         []string          `json:"tags"`            // 原始标签
	Label        string            `json:"label,omitempty"` // 模型打标
	Meta         *wcd.SentenceMeta `json:"meta"`
	WebSegmentId int32             `json:"web_segment_id"`
}

func NewLabelInfoFromSentence(sentence *wcd.AtomicSentence) *LabelInfo {
	atoms := utils.Map(sentence.Atoms, func(atom *wcd.AtomicText) *text_parse.AtomicTxt {
		return &text_parse.AtomicTxt{
			Txt:        &atom.Text,
			PositionID: atom.PositionID,
			X:          atom.Xpath,
		}
	})

	labelMeta := &wcd.SentenceMeta{}
	if sentence != nil {
		labelMeta.URL = sentence.Meta.GetURL()
		labelMeta.TableHTML = sentence.Meta.GetTableHTML()
	}
	return &LabelInfo{
		Txt:   sentence.Text,
		Meta:  labelMeta,
		Label: "O", // 默认传O，让text-parse来修改
		Position: &LabelPosition{
			Atoms: atoms,
		},
		Tags:         sentence.Tags,
		WebSegmentId: sentence.SegmentID,
	}
}

type LabelModelReq struct {
	ArticleMeta wcd.ArticleMeta `json:"article_meta"` // web需要
	EntryId     string          `json:"entry_id"`
	Type        string          `json:"type"`
	Infos       []*LabelInfo    `json:"infos"`
}

const (
	LabelModelCode_Success    = 0
	LabelModelCode_InfosEmpty = 1001 // infos为空
	LabelModelCode_AllEduO    = 1002 // 所有infos都是edu_o标签
	LabelModelCode_Unk        = 2001 // 未知错误
)

type LabelModelResp struct {
	Code         int              `json:"code"`
	Msg          string           `json:"msg"`
	Type         string           `json:"type"`
	Infos        []*TextParseInfo `json:"infos"`
	ArticleMeta  *wcd.ArticleMeta `json:"article_meta"`
	LabelVersion string           `json:"version"`
}
