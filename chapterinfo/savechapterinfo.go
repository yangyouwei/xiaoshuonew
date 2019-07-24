package chapterinfo

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/yangyouwei/xiaoshuonew/bookinfo"
	"github.com/yangyouwei/xiaoshuonew/chapterfilter"
	"github.com/yangyouwei/xiaoshuonew/conflib"
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
	var linesch = make(chan string,100)
	fi, err := os.Open(b.Sourcesfilename)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	for  {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			return
		}
		if chapterfilter.IfMatch(conflib.HR.Hr,a)||chapterfilter.IfMatch(b.ChapterRules,a) {
			linesch <- string(a)
			break
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	//save 章节到数据库
	go func(wg *sync.WaitGroup) {
		SaveChapter(Chapters,db)
		wg.Done()
	}(&wg)

	//获取章节。将章节存入chan
	go func(wg *sync.WaitGroup) {
		getchapter(linesch,&b,Chapters)
		wg.Done()
	}(&wg)

	//去掉头部，不是章节和前言的部分，省下的全部放到ch中
	go func(wg *sync.WaitGroup) {
		for {
			a, _, c := br.ReadLine()
			if c == io.EOF {
				break
			}
			line := deletead(string(a))
			if line == "" {
				continue
			}
			linesch <- line
		}
		close(linesch)
		wg.Done()
	}(&wg)
	wg.Wait()
}

func getchapter(lines chan string,b *bookinfo.BookInfo,chapterch chan ChapterInof) {
	tempchapter := ""
	loop:
	for {
		c := ChapterInof{}
		c.Bookid = b.Bookid
		for  {
			//读取行
			line,isclose := <- lines
			if !isclose {
				c.Size = len(c.Content)/3
				chapterch <- c
				close(chapterch)
				fmt.Println("finished")
				return
			}

			isok , err := regexp.Match(conflib.HR.Hr,[]byte(line))
			if err != nil {
				fmt.Println(err)
			}

			if isok {
				if c.Content == "" {
					c.Chaptername = line
					continue
				}else {
					tempchapter = line
					break
				}
			}
			//判断章节
			isok , err = regexp.Match(b.ChapterRules,[]byte(line))
			if err != nil {
				fmt.Println(err)
			}
			if isok {
				//fmt.Println(string(line))
				if c.Content == "" {
					c.Chaptername = line
					continue
				}else {
					tempchapter = line
					break
				}
			}
			
			if c.Chaptername == "" {
				c.Chaptername = tempchapter
			}
			//fmt.Println(c.Chaptername)
			//去空行
			isok3 , err := regexp.Match(`^\s*$`,[]byte(line))
			if err != nil {
				fmt.Println(err)
			}
			if isok3 {
				continue
			}
			//去掉广告
			//isok1 := strings.HasPrefix(string(line),conflib.Adrule_str.Adstring)
			//if isok1 {
			//	isok , err := regexp.Match(conflib.Adrule_str.Rules1,line)  //删除广告行
			//	if err != nil {
			//		fmt.Println(err)
			//	}
			//	if isok {
			//		continue
			//	}
			//	reg := regexp.MustCompile(conflib.Adrule_str.Rules2)
			//	result := reg.FindAllStringSubmatch(string(line),-1)
			//	s := result[0][1] + result[0][3]    //内容中有广告，去广告。保留内容
			//	line = []byte(s)
			//}
			//加标签去行首空白字符
			if c.Content == "" {
				c.Content = "&nbsp&nbsp&nbsp&nbsp" + fmtline(line) + "</br></br>"
			}else {
				c.Content = c.Content + fmtline(line) + "</br></br>"
			}
			//章节字数小于100 忽略。去重

		}
		c.Size = len(c.Content)/3
		if c.Size < 100 {
			c.Content = ""
			goto loop
		}
		chapterch <- c
		goto loop
	}
}

func deletead(s string) string {
	isok1 := strings.HasPrefix(s,conflib.Adrule_str.Adstring)
	if isok1 {
		isok , err := regexp.Match(conflib.Adrule_str.Rules2,[]byte(s))
		if err != nil {
			fmt.Println(err)
		}
		if isok {
			reg := regexp.MustCompile(conflib.Adrule_str.Rules2)
			result := reg.FindAllStringSubmatch(s,-1)
			return result[0][2]
		}

		isok2 , err := regexp.Match(conflib.Adrule_str.Rules1,[]byte(s))
		if err != nil {
			fmt.Println(err)
		}
		if isok2 {
			return ""
		}
	}
	return s
}

func fmtline(s string) string {
	isok , err := regexp.Match(`^(\s+)(.+)$`,[]byte(s))
	if err != nil {
		fmt.Println(err)
	}
	if isok {
		reg := regexp.MustCompile(`^(\s+)(.+$)`)
		result := reg.FindAllStringSubmatch(s,-1)
		s = result[0][2]
	}
	return s
}

func SaveChapter(c chan ChapterInof,db *sql.DB)  {
	cn := 0
	for  {
		chater, isok := <- c
		if !isok  {
			break
		}
		cn = cn + 1
		chater.Chapterid = cn
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

}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}



