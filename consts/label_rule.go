package consts

type LabelRule struct {
	Label    string
	CleanTag bool
	NewTag   string
	NewAttr  string
}
type SubtreeLabelRule struct {
	LabelRule
	SameParagraph bool // 只能合并同一段落的子树
}

var LABEL_RULES = []LabelRule{
	{LABEL_NOISE, true, "", ""},
	{LABEL_REFERENCE, false, "", ATTR_PREFIX + "reference"},
	{LABEL_TITLE, true, "", ""},
	{LABEL_PUB_TIME, true, "", ""},
	{LABEL_AUTHOR, true, "", ""},
	{LABEL_SOURCE, true, "", ""},
}

var SUBTREE_RULES = []SubtreeLabelRule{
	{LabelRule{LABEL_INTRO, false, "", ATTR_PREFIX + "intro"}, false},
	{LabelRule{LABEL_ABSTRACT, false, "", ATTR_PREFIX + "intro"}, false},
	{LabelRule{LABEL_CATALOG, false, "", ATTR_PREFIX + "intro"}, false},
	//{LabelRule{LABEL_REFERENCE, false, "", ATTR_PREFIX + "reference"}, true},
	{LabelRule{LABEL_LEGEND, false, "", ATTR_PREFIX + "legend"}, false},

	{LabelRule{LABEL_TITLE_L1, false, "", ATTR_PREFIX + "h1"}, true},
	{LabelRule{LABEL_TITLE_L2, false, "", ATTR_PREFIX + "h2"}, true},
	{LabelRule{LABEL_TITLE_L3, false, "", ATTR_PREFIX + "h3"}, true},
	{LabelRule{LABEL_TITLE_L4, false, "", ATTR_PREFIX + "h4"}, true},
	{LabelRule{LABEL_TITLE_OTHER, false, "", ATTR_PREFIX + "h5"}, true},
}
