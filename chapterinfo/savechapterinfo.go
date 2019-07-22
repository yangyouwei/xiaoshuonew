package chapterinfo

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/yangyouwei/xiaoshuonew/bookinfo"
	"github.com/yangyouwei/xiaoshuonew/chapterfilter"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
)

type ChapterInof struct {
	Bookid      int64    `db:"booksId"`
	Chapterid   int    `db:"chapterId"`
	Chaptername string `db:"chapterName"`
	Content     string `db:"content"`
	Size		int `db:"size"`
}

func GetChapterContent(b bookinfo.BookInfo,db *sql.DB)  {
	var Chapters = make(chan ChapterInof,100)
	var linesch = make(chan []byte)
	fi, err := os.Open(b.Sourcesfilename)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()
	isok := false
	br := bufio.NewReader(fi)
	for  {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			return
		}
		isok = chapterfilter.IfMatch(b.ChapterRules,a)
		isok = chapterfilter.IfMatch(b.ChapterRules,a)
		if isok {
			linesch <- a
			break
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	//去掉头部，不是章节和前言的部分，省下的全部放到ch中
	go func(wg *sync.WaitGroup) {
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				return
			}
			linesch <- a
		}
		close(linesch)
		wg.Done()
	}(&wg)

	//获取章节。将章节存入chan
	go func(wg *sync.WaitGroup) {
		for {
			chapter := 	getchapter(linesch,&b)
			Chapters <- *chapter
		}
		wg.Done()
	}(&wg)
	//save 章节到数据库
	go func(wg *sync.WaitGroup) {
		SaveChapter(Chapters,db)
		wg.Done()
	}(&wg)
	wg.Wait()
}

func getchapter(lines chan []byte,b *bookinfo.BookInfo) *ChapterInof {
	c := ChapterInof{}
	c.Bookid = b.Bookid
	n := 0
	for  {
		line,isclose := <- lines
		if !isclose {
			break
		}
		isok , err := regexp.Match(b.ChapterRules,line)
		if err != nil {
			fmt.Println(err)
		}
		if isok {
			n++
			c.Chaptername = string(line)
			c.Chapterid = n
			continue
		}
		//去空行
		isok3 , err := regexp.Match(`^\s*$`,line)
		if err != nil {
			fmt.Println(err)
		}
		if isok3 {
			continue
		}
		//去掉广告
		isok1 := strings.HasPrefix(string(line),"更多精彩，更多好书，尽在新奇书网—http://www.xqishu.com")
		if isok1 {
			isok , err := regexp.Match(`^(.*)(更多精彩，更多好书，尽在新奇书网—http://www.xqishu.com)$`,line)
			if err != nil {
				fmt.Println(err)
			}
			if isok {
				continue
			}
			reg := regexp.MustCompile(`^(.*)(更多精彩，更多好书，尽在新奇书网—http://www.xqishu.com)(.+$)`)
			result := reg.FindAllStringSubmatch(string(line),-1)
			s := result[0][3]
			line = []byte(s)
		}
		//加标签去行首空白字符
		c.Content = "&nbsp&nbsp&nbsp&nbsp" + fmtline(string(line)) + c.Content + "</br></br>"
	}
	c.Size = len(c.Content)/3
	return &c
}

func fmtline(s string) string {
	isok , err := regexp.Match(`^(.+)(\s+)(.+)$`,[]byte(s))
	if err != nil {
		fmt.Println(err)
	}
	if isok {
		reg := regexp.MustCompile(`^(.+)(\s+)(.+$)`)
		result := reg.FindAllStringSubmatch(s,-1)
		s = result[0][1]+result[0][3]
	}
	return s
}

func SaveChapter(c chan ChapterInof,db *sql.DB)  {
	chater := <- c
	n := chater.Bookid%int64(100)+int64(1)
	chapternum := "chapter_" + fmt.Sprint(n)

	chpatertable := fmt.Sprintf("INSERT %v ( booksId, chapterName, size,content,chapterId) VALUES (?,?,?,?,?)", chapternum)
	stmt, err := db.Prepare(chpatertable)
	check(err)

	res, err := stmt.Exec(chater.Bookid,chater.Chaptername,chater.Size,chater.Content,chater.Chapterid)
	check(err)

	_, err = res.LastInsertId() //必须是自增id的才可以正确返回。
	check(err)
	defer stmt.Close()
	stmt.Close()
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}


