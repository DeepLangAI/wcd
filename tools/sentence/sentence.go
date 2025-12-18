package sentence

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/DeepLangAI/wcd/consts"
	"github.com/DeepLangAI/wcd/utils"

	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/data"
	"github.com/neurosnap/sentences/english"
)

type SentenceOp struct {
	cutter *sentences.DefaultSentenceTokenizer
}

func NewSentenceOp() *SentenceOp {
	tokenizer, err := NewSentenceTokenizer(nil)
	if err != nil {
		panic(err)
	}
	return &SentenceOp{
		cutter: tokenizer,
	}
}

func (so *SentenceOp) Cut(s string) []*sentences.Sentence {
	tokenize := so.cutter.Tokenize(s)
	return tokenize
}

type PunctStrings struct{}

// NewPunctStrings creates a default set of properties
func NewPunctStrings() *PunctStrings {
	return &PunctStrings{}
}

// NonPunct regex string to detect non-punctuation.
func (p *PunctStrings) NonPunct() string {
	return `[^\W\d]`
}

// Punctuation characters
func (p *PunctStrings) Punctuation() string {
	return ";:,.!?；：，。！？"
}

// HasSentencePunct does the supplied text have a known sentence punctuation character?
func (p *PunctStrings) HasSentencePunct(text string) bool {
	endPunct := consts.END_PUNCT
	for _, char := range endPunct {
		for _, achar := range text {
			if char == achar {
				return true
			}
		}
	}

	return false
}

type WordTokenizer struct {
	sentences.DefaultWordTokenizer
}

func IsCjkPunct(r rune) bool {
	allPuncts := append(consts.CHINESE_SENTENCE_STOP_SIGN, consts.SENTENCE_STOP_EXT...)
	//allPuncts := consts.CHINESE_SENTENCE_STOP_SIGN
	allPuncts = append(allPuncts, consts.ENGLISH_SENTENCE_STOP_SIGN...)
	if utils.Contains(
		allPuncts,
		string(r),
	) {
		return true
	}
	return false
	//switch r {
	//case '。', '；', '！', '？', ';', '：', '…':
	//	return true
	//}
	//return false
}

func NewWordTokenizer(p sentences.PunctStrings) *WordTokenizer {
	word := &WordTokenizer{}
	word.PunctStrings = p

	return word
}

func (p *WordTokenizer) Tokenize(text string, onlyPeriodContext bool) []*sentences.Token {
	textLength := len(text)

	if textLength == 0 {
		return nil
	}

	tokens := make([]*sentences.Token, 0, 50)
	lastSpace := 0
	lineStart := false
	paragraphStart := false
	getNextWord := false

	for i, char := range text {
		// 新增：遇到换行符时强制分割句子
		if char == '\n' {
			cursor := i
			if lastSpace <= cursor && cursor < textLength {
				word := strings.TrimSpace(text[lastSpace:cursor])
				if word != "" {
					token := sentences.NewToken(word)
					token.Position = cursor
					token.LineStart = true // 标记为新行开始
					token.SentBreak = true
					tokens = append(tokens, token)
					lastSpace = cursor + 1 // 跳过换行符，下一句从新行开始
					continue               // 跳过后续常规处理
				}
			}
		}

		if !unicode.IsSpace(char) && !IsCjkPunct(char) && i != textLength-1 {
			continue
		}

		if IsCjkPunct(char) {
			i += len(string(char))
		}

		if char == '\n' {
			if lineStart {
				paragraphStart = true
			}
			lineStart = true
		}

		var cursor int
		if i == textLength-1 {
			cursor = textLength
		} else {
			cursor = i
		}

		word := strings.TrimSpace(text[lastSpace:cursor])

		if word == "" {
			continue
		}

		hasSentencePunct := p.PunctStrings.HasSentencePunct(word)
		if onlyPeriodContext && !hasSentencePunct && !getNextWord {
			lastSpace = cursor
			continue
		}

		token := sentences.NewToken(word)
		token.Position = cursor
		token.ParaStart = paragraphStart
		token.LineStart = lineStart
		tokens = append(tokens, token)

		lastSpace = cursor
		lineStart = false
		paragraphStart = false

		if hasSentencePunct {
			getNextWord = true
		} else {
			getNextWord = false
		}
	}
	if lastSpace < textLength {
		token := sentences.NewToken(text[lastSpace:])
		token.Position = textLength
		tokens = append(tokens, token)
	}

	if len(tokens) == 0 {
		token := sentences.NewToken(text)
		token.Position = textLength
		tokens = append(tokens, token)
	}

	return tokens
}

func NewSentenceTokenizer(s *sentences.Storage) (*sentences.DefaultSentenceTokenizer, error) {
	training := s

	if training == nil {
		b, err := data.Asset("data/english.json")
		if err != nil {
			return nil, err
		}

		training, err = sentences.LoadTraining(b)
		if err != nil {
			return nil, err
		}
	}

	// supervisor abbreviations
	abbrevs := []string{"sgt", "gov", "no"}
	for _, abbr := range abbrevs {
		training.AbbrevTypes.Add(abbr)
	}

	lang := NewPunctStrings()
	word := NewWordTokenizer(lang)
	annotations := sentences.NewAnnotations(training, lang, word)

	ortho := &sentences.OrthoContext{
		Storage:      training,
		PunctStrings: lang,
		TokenType:    word,
		TokenFirst:   word,
	}

	multiPunct := &english.MultiPunctWordAnnotation{
		Storage:      training,
		TokenParser:  word,
		TokenGrouper: &sentences.DefaultTokenGrouper{},
		Ortho:        ortho,
	}

	annotations = append(annotations, multiPunct)
	annotations = append(
		annotations,
		&DeeplangiPunctWordAnnotation{
			TokenGrouper: &sentences.DefaultTokenGrouper{},
		},
	)

	tokenizer := &sentences.DefaultSentenceTokenizer{
		Storage:       training,
		PunctStrings:  lang,
		WordTokenizer: word,
		Annotations:   annotations,
	}

	return tokenizer, nil
}

type DeeplangiPunctWordAnnotation struct {
	sentences.TokenGrouper
}

func (a *DeeplangiPunctWordAnnotation) Annotate(tokens []*sentences.Token) []*sentences.Token {
	for _, tokPair := range a.TokenGrouper.Group(tokens) {
		if len(tokPair) < 2 || tokPair[1] == nil {
			continue
		}

		a.tokenAnnotation(tokPair[0], tokPair[1])
	}
	a.tokenAnnotationTime(tokens)
	a.tokenAnnotationBrackets(tokens)

	return tokens
}

func (a *DeeplangiPunctWordAnnotation) tokenAnnotationTime(tokens []*sentences.Token) {
	newTokens := utils.Map(tokens, func(t *sentences.Token) *sentences.Token {
		newToken := &sentences.Token{
			Tok:       t.Tok + " ",
			Position:  t.Position,
			SentBreak: t.SentBreak,
			ParaStart: t.ParaStart,
			LineStart: t.LineStart,
			Abbr:      t.Abbr,
		}
		return newToken
	})
	joinText := strings.Join(utils.Map(newTokens, func(t *sentences.Token) string {
		return t.Tok
	}), "")

	// 计算每个原子文本在拼接字符串中的位置
	intervals := make([]Interval, len(newTokens))
	start := 0
	for i, token := range newTokens {
		end := start + len(token.Tok)
		intervals[i] = Interval{
			Start: start,
			End:   end - 1,
			ID:    i,
		}
		start = end // 更新起始位置
	}

	// 时间不断句
	for _, pattern := range consts.DATETIME_PATTERN {
		pattern = strings.ReplaceAll(pattern, ":", `\s*:\s*`)
		regex := regexp.MustCompile(pattern)
		if index := regex.FindStringIndex(joinText); index != nil {
			begin, end := index[0], index[1]
			beginIntervalIndex, success := FindInterval(begin, intervals)
			if !success {
				continue
			}
			endIntervalIndex, success := FindInterval(end-1, intervals)
			if !success {
				continue
			}
			for i := beginIntervalIndex; i < endIntervalIndex; i++ {
				tokens[i].SentBreak = false
			}
			//tokens[endIntervalIndex].SentBreak = true
			break
		}
	}
}

func (a *DeeplangiPunctWordAnnotation) tokenAnnotation(tokOne, tokTwo *sentences.Token) {
	// 链接不断句
	if (strings.HasSuffix(tokOne.Tok, "https:") || strings.HasSuffix(tokOne.Tok, "http:")) && strings.HasPrefix(tokTwo.Tok, "//") {
		tokOne.SentBreak = false
		return
	}

	if strings.HasSuffix(tokOne.Tok, ".") && tokTwo.Tok == "." {
		tokOne.SentBreak = false
		return
	}
	allPuncts := []string{}
	allPuncts = append(allPuncts, consts.CHINESE_SENTENCE_STOP_SIGN...)
	allPuncts = append(allPuncts, consts.ENGLISH_SENTENCE_STOP_SIGN...)

	for _, stop := range allPuncts {
		if strings.HasSuffix(tokOne.Tok, stop) {
			tokOne.SentBreak = true
			break
		}
	}
	for _, stop := range allPuncts {
		if strings.HasSuffix(tokTwo.Tok, stop) {
			tokTwo.SentBreak = true
			break
		}
	}
	extra := append(consts.SENTENCE_STOP_EXT, consts.CHINESE_SENTENCE_STOP_SIGN...)
	if tokOne.SentBreak && utils.Any(extra, func(stop string) bool {
		return strings.HasPrefix(tokTwo.Tok, stop)
	}) {
		tokOne.SentBreak = false
		tokTwo.SentBreak = true
	}
}

func (a *DeeplangiPunctWordAnnotation) tokenAnnotationBrackets(tokens []*sentences.Token) {
	pairs := map[string]map[string]int{
		"（": {
			"）": 1,
			")": 1,
		},
		"(": {
			")": 1,
			"）": 1,
		},
		"【": {"】": 1},
		"[": {"]": 1},
		"《": {"》": 1},
		"<": {">": 1},
		"「": {"」": 1},
		"『": {"』": 1},
		"{": {"}": 1},
		//"“": "”",
		//"‘": "’",
	}
	i := 0
	stack := utils.NewStack[string]()
	for i < len(tokens) {
		token := tokens[i]
		for _, char := range []rune(token.Tok) {
			charStr := string(char)
			if _, ok := pairs[charStr]; ok {
				stack.Push(charStr)
			} else {
				if !stack.IsEmpty() {
					top, _ := stack.Peek()
					if _, ok := pairs[top][charStr]; ok {
						stack.Pop()
					}
				}
			}
		}
		if !stack.IsEmpty() {
			if token.SentBreak {
				token.SentBreak = false
			}
		}

		i++
	}
}
