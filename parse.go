package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// parseHtmlToMd 将字典HTML内容转换为Markdown格式
func parseHtmlToMd(htmlContent string) (string, string, error) {
	word := ""
	// 创建goquery文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", "", fmt.Errorf("解析HTML失败: %v", err)
	}

	var md strings.Builder

	// 解析词汇条目
	doc.Find(".trans.ty_pinyin").Remove()
	doc.Find(".entry").Each(func(i int, entry *goquery.Selection) {
		// 解析词汇标题
		title := entry.AttrOr("d:title", "")
		if title != "" {
			md.WriteString(fmt.Sprintf("# %s\n\n", title))
		}
		word = title

		// 解析发音信息
		parsePronunciation(entry, &md)

		// 解析词性和释义
		parseGrammarBlocks(entry, &md)

		// 解析短语动词
		parsePhrasalVerbs(entry, &md)
	})

	return word, md.String(), nil
}

var phDialectMap = map[string]string{
	"BrE": "英",
	"AmE": "美",
}

// parsePronunciation 解析发音信息
func parsePronunciation(entry *goquery.Selection, md *strings.Builder) {
	hwgNode := entry.Find(".hwg")
	if hwgNode.Length() <= 0 {
		return
	}
	hwgNode.Find(".prx").Each(func(i int, prx *goquery.Selection) {
		dialect := prx.AttrOr("dialect", "")
		pronunciation := prx.Find(".ph").Text()

		if dialect != "" && pronunciation != "" {
			md.WriteString(fmt.Sprintf("`%s/%s/` ", phDialectMap[dialect], pronunciation))
		}
	})

	// 如果有发音信息，添加换行
	if hwgNode.Find(".prx").Length() > 0 {
		md.WriteString("\n\n")
		return
	}

	hwgNode.Find(".pr").Each(func(i int, prx *goquery.Selection) {
		pronunciation := prx.ChildrenFiltered(".ph").Text()
		md.WriteString(fmt.Sprintf("`%s`\n\n", strings.TrimSpace(pronunciation)))
	})
}

// parseGrammarBlocks 解析语法块（词性分类）
func parseGrammarBlocks(entry *goquery.Selection, md *strings.Builder) {
	entry.ChildrenFiltered(".gramb").Each(func(i int, gramb *goquery.Selection) {
		// 获取词性标签（如 A., B., C.）
		pos := gramb.Children().First().Text()

		if pos != "" {
			md.WriteString(fmt.Sprintf("##### %s\n\n", pos))
		}

		// 解析语义块
		parseSemanticBlocks(gramb, md)
	})
}

// parseSemanticBlocks 解析语义块（不同释义）
func parseSemanticBlocks(gramb *goquery.Selection, md *strings.Builder) {
	gramb.Find(".semb").Each(func(i int, semb *goquery.Selection) {
		// 获取释义编号
		senseNum := semb.ChildrenFiltered(".tg_semb").Text()

		// 获取翻译
		translation := semb.ChildrenFiltered(".trg").Text()

		if translation == "" {
			translation = semb.ChildrenFiltered(".trgg").Text()
		}

		// 写入释义
		if senseNum != "" || translation != "" {
			md.WriteString(fmt.Sprintf("###### %s %s", senseNum, translation))
			md.WriteString("\n\n")
		}

		// 解析例句
		parseExamples(semb, md)
	})
}

// parseExamples 解析例句
func parseExamples(parent *goquery.Selection, md *strings.Builder) {
	parent.Find(".exg").Each(func(i int, exg *goquery.Selection) {
		example := exg.Find(".ex").Text()
		translation := exg.Find(".trans").Text()

		if example != "" {
			md.WriteString(fmt.Sprintf("- *%s*\n", strings.TrimSpace(example)))
			if translation != "" {
				md.WriteString(fmt.Sprintf("  - *%s*\n", strings.TrimSpace(translation)))
			}
		}
	})

	if parent.Find(".exg").Length() > 0 {
		md.WriteString("\n")
	}
}

// parsePhrasalVerbs 解析短语动词
func parsePhrasalVerbs(entry *goquery.Selection, md *strings.Builder) {
	phrasalSection := entry.Find(".pvb")
	if phrasalSection.Length() == 0 {
		return
	}

	md.WriteString(fmt.Sprintf("## %s\n\n", phrasalSection.Children().First().Text()))

	phrasalSection.Find(".pvsec").Each(func(i int, pvsec *goquery.Selection) {
		// 获取短语动词
		phrasal := pvsec.Find(".pv").Text()
		if phrasal != "" {
			md.WriteString(fmt.Sprintf("**%s**\n\n", strings.TrimSpace(phrasal)))
		}

		// 获取词性
		pos := pvsec.Find(".ps").Text()
		if pos != "" {
			md.WriteString(fmt.Sprintf("%s\n\n", pos))
		}

		pvsec.Find(".semb").Each(func(i int, semb *goquery.Selection) {
			// 获取释义
			md.WriteString(fmt.Sprintf("%s\n\n", semb.First().Text()))
		})

		// 解析例句
		parseExamples(pvsec, md)
	})
}
