package bookinfo

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/yangyouwei/xiaoshuonew/chapterfilter"
	"io"
	"os"
	"path/filepath"
	"strings"
)
//var cr = `^\s*.*第(\d+|[一二三四五六七八九十百千万]+)章.*[^。]*$`

type BookInfo struct {
	Bookid          int64  `db:"id"`
	Bookname        string `db:"booksName"`
	Bookcahtpernum  int    `db:"chapters"`
	Sourcesfilename string `db:"sourcesfilename"`
	ChapterRules    string `db:"regexRules"`
}

func GetBookInfo(bookpath string,b *BookInfo)  {
	bn := strings.Split(filepath.Base(bookpath), ".")
	b.Bookname = bn[2]
	b.Sourcesfilename = bookpath
	GetBookRules(b)
}

func SaveBookInfo(b *BookInfo,db *sql.DB)  {
	fmt.Println(b)
	stmt, err := db.Prepare(`INSERT books ( booksName,chapters,sourcesfilename,regexRules) VALUES (?,?,?,?)`)
	check(err)

	res, err := stmt.Exec(b.Bookname,b.Bookcahtpernum,b.Sourcesfilename,b.ChapterRules)
	check(err)

	_, err = res.LastInsertId() //必须是自增id的才可以正确返回。
	check(err)
	defer stmt.Close()

	//idstr := fmt.Sprintf("%v", id)
	//fmt.Println(idstr)
	stmt.Close()
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetBookRules(b *BookInfo){
	//匹配章节规则
	//var isok bool
	rulesmap := chapterfilter.Makemap()
	fi, err := os.Open(b.Sourcesfilename)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	for i := 0; i < 500; i++ {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		for k, v := range rulesmap {
			isok := chapterfilter.IfMatch(k,a)
			if isok {
				rulesmap[k] = v+1
				fmt.Println(k,v)
			}

		}
	}
	b.ChapterRules = chapterfilter.RulesSort(rulesmap)
}

//func getrule(s string) (isok bool,r string) {
//	//匹配  xxx第xx章xxx  类似章节名称
//	isok1 , err := regexp.Match(cr,[]byte(s))
//	if err != nil {
//		fmt.Println(err)
//	}
//	//如果能匹配，则匹配更详细的规则作为该小说的规则
//	if isok1 {
//		//fmt.Println(s,":",cr)
//		rules := *conflib.Chapterrules1.Rules
//		for _,v := range rules {
//
//			isok2 , _ := regexp.Match(v,[]byte(s))
//			if isok2 {
//				//fmt.Println(v)
//				return true, v
//			}
//		}
//	}
//	//如果都匹配不上，使用规则2中的规则。
//	rules := *conflib.Chapterrules2.Rules
//	for _,v := range rules {
//		isok2 , _ := regexp.Match(v,[]byte(s))
//		if isok2 {
//			//fmt.Println(v)
//			return true, v
//		}
//	}
//	return false, ""
//}