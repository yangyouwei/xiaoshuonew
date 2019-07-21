package chapterinfo

type ChapterInof struct {
	Bookid      int    `db:"booksId"`
	Chapterid   int    `db:"chapterId"`
	Chaptername string `db:"chapterName"`
	Chapter     string `db:"content"`
}


