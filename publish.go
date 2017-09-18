package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	TMPL_WIKI_HAED = `
---
title: @TITLE
comments: false
categories: @CATEGORIES
tags:
@TAGS
date: @DATE
---

<!-- toc -->`

	MORE_TAG = `<!-- more -->`
)

var title string
var categories string
var tags string
var formatTags string
var lenNum int
var srcPath string
var dstPath string

func main() {

	flag.StringVar(&title, "t", "", "wiki titile name")
	flag.StringVar(&categories, "c", "", "set wiki document categories")
	flag.StringVar(&tags, "g", "", "set wiki document set tags")
	flag.IntVar(&lenNum, "l", 10, "set wiki document set tags")

	flag.StringVar(&srcPath, "src", "", "set src Markdown file path")
	flag.StringVar(&dstPath, "dst", "", "set dst Markdown file path")

	flag.Parse()

	if len(title) < 1 && len(categories) < 1 && len(tags) < 1 {
		flag.Usage()

		os.Exit(1)
	}

	//Args check
	checkAndSetTitle()
	checkAndSetCategories()
	checkAndSetTags()

	if len(dstPath) < 1 {
		dstPath = "/tmp/" + title + ".md"
	}

	//Make formated-wiki doc
	NewWikiMKDocument()

}

func NewWikiMKDocument() {

	wikiHead := genWikiHead()
	tagFile, err := setMoreTagFromFile(srcPath, lenNum)
	if err != nil {
		fmt.Println("set More tag to file and get context failed.Err:", err)
		os.Exit(1)
	}
	imgPathTrans := replaceImgPath(tagFile)
	wikiDoc := wikiHead + "\n" + imgPathTrans

	ioutil.WriteFile(dstPath, []byte(wikiDoc), 0644)
	//	fmt.Println(wikiDoc)
}

func getTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func getMDfileContext(filename string) string {

	context, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("get MarkDown file failed .ERR:", err)
		return ""
	}
	return string(context)
}
func genWikiHead() string {
	var wikiHead = TMPL_WIKI_HAED

	s := strings.Replace(wikiHead, "@TITLE", title, 1)
	s = strings.Replace(s, "@CATEGORIES", categories, 1)
	s = strings.Replace(s, "@TAGS", formatTags, 1)
	s = strings.Replace(s, "@DATE", getTime(), 1)

	return s
}

func replaceImgPath(data string) string {

	var imgPaths []string
	re := regexp.MustCompile(`\!\[.*\]\(\.\S+\)`)

	matchs := re.FindAllString(data, -1)

	for _, v := range matchs {
		if strings.Contains(v, "(./") {
			paths := strings.SplitN(v, "(.", 2)
			paths[1] = paths[1][:len(paths[1])-1]
			imgPaths = append(imgPaths, paths[1])
		}
		continue
	}

	for _, v := range imgPaths {

		sub := strings.SplitAfter(v, "/")
		filename := sub[len(sub)-1]

		rePath := "/" + title + "/" + filename
		data = strings.Replace(data, v, rePath, -1)
	}

	return data
}

func checkAndSetTitle() {
	if len(title) < 1 {
		fmt.Println("wiki titile name is not set.")
		os.Exit(2)
	}

}

func checkAndSetCategories() {
	if len(categories) < 1 {
		fmt.Println("wiki categories is not set.")
		os.Exit(3)
	}

}

func checkAndSetTags() {
	if len(tags) < 1 {
		fmt.Println("wiki tags  is not set.")
		os.Exit(4)
	}

	tagSlice := strings.Split(tags, ",")

	for i, v := range tagSlice {
		if i == len(tagSlice)-1 {
			v = "   - " + v
		} else {
			v = "   - " + v + "\n"
		}
		formatTags = formatTags + v
	}
}

func insertMoreTag(s []string, line int) {

	up := s[:line]
	bottom := s[line:]

	up = append(up, "\n", MORE_TAG, "\n")

	copy(up, bottom)

}

func readLines(data string) []string {

	var lines []string
	rd := bufio.NewReader(strings.NewReader(data))

	for {
		str, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		lines = append(lines, str)
	}
	return lines
}

func setMoreTagFromFile(fileName string, line int) (string, error) {

	var index int
	var offset int

	f, err := os.OpenFile(fileName, os.O_RDWR, 0644)

	if err != nil {
		fmt.Println("file open failed.err: ", err.Error())
	}

	rd := bufio.NewReader(f)
	for {
		if index == line {
			break
		}
		str, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		offset = offset + len(str)
		index++
	}

	//分割读取前Ｎ行
	fileDataUp := make([]byte, offset)
	_, err = f.ReadAt(fileDataUp, 0)

	//读取后N行
	n, _ := f.Seek(0, os.SEEK_END)
	fileDataBotton := make([]byte, n-int64(offset))
	_, err = f.ReadAt(fileDataBotton, int64(offset))

	defer f.Close()

	newData := string(fileDataUp) + "\n" + MORE_TAG + "\n" + string(fileDataBotton)

	return newData, nil
}
