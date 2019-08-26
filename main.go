package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yangyouwei/xiaoshuonew/bookinfo"
	"github.com/yangyouwei/xiaoshuonew/chapterinfo"
	"github.com/yangyouwei/xiaoshuonew/conflib"
	"io/ioutil"
	"sync"
)

var (
	//书籍存储路径
	BookStorePath string = conflib.Main_str.Filepath
	//并发配置
	Concurrent int = conflib.Main_str.Concurrent
	//book ch
	Books =make(chan string)
	//mysql
	MysqlString string = conflib.Mysql_conf_str.Username + ":" + conflib.Mysql_conf_str.Password + "@tcp(" + conflib.Mysql_conf_str.Ipaddress + ":" + conflib.Mysql_conf_str.Port + ")/" + conflib.Mysql_conf_str.DatabaseName
	//mysql 连接池
	Db *sql.DB
	//err
	err error
)

//book info
type Bookinfo struct {
	Bookname string
	BookId   int64
	Chapterrule string
}

//初始化mysql连接
func init()  {
	Db, err = sql.Open("mysql",MysqlString)
	check(err)
}

func main() {
	//获取小说的路径
	wg := sync.WaitGroup{}
	wg.Add(Concurrent+1)
	go func(wg *sync.WaitGroup) {
		GetBooksNames(BookStorePath,Books)
		close(Books)
		wg.Done()
	}(&wg)

	for i := 0; i < Concurrent; i++ {
		go ChapterWorker(Books,&wg)

	}
	wg.Wait()
}

func GetBooksNames(bookspath string,fn_ch chan string) {
	rd, err := ioutil.ReadDir(bookspath)
	if err != nil {
		fmt.Println("read dir fail:", err)
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := bookspath + "/" + fi.Name()
			GetBooksNames(fullDir, fn_ch)
			if err != nil {
				fmt.Println("read dir fail:", err)
			}
		} else {
			fullName := bookspath + "/" + fi.Name()
			fn_ch <- fullName
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func ChapterWorker(books chan string,wg *sync.WaitGroup)  {
	concurrent := conflib.Main_str.Concurrent
	for i:=0; i < concurrent; i++ {
		go func() {
			for {
				fp, isclose := <- books
				if !isclose {
					break
				}
				//获取bookinfo，章节规则，并存储
				b := bookinfo.BookInfo{}
				bookinfo.GetBookInfo(fp,&b)
				//fmt.Println(b.Bookname,":",b.ChapterRules)
				bookinfo.SaveBookInfo(&b,Db)
				//获取章节信息，并存储
				chapterinfo.GetChapterContent(b,Db)
				fmt.Println(b.Bookname,": ","is done")
			}
		}()
	}
	wg.Done()
}
