package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Database interface {
	GetSeries(book ReadBook) []SeriesBook
	GetReadBooks() []ReadBook
	Close() error
}

type ReadBook struct {
	BookId          int    `db:"book_id"`
	BookName        string `db:"book_name"`
	BookSeriesIndex int    `db:"book_series_index"`
	SeriesId        int    `db:"series_id"`
	AnilistId       int    `db:"anilist_id"`
}

type SeriesBook struct {
	BookId          int    `db:"book_id"`
	BookName        string `db:"book_name"`
	BookSeriesIndex int    `db:"book_series_index"`
	AnilistId       int    `db:"anilist_id"`
	Chapters        int    `db:"chapters"`
}

type database struct {
	db *sqlx.DB
}

func dbConnection(appDb string, calibreDb string) *sqlx.DB {
	fileFmt := "file:%s?mode=ro"
	calibreDb = fmt.Sprintf(fileFmt, calibreDb)
	appDb = fmt.Sprintf(fileFmt, appDb)
	db := sqlx.MustConnect("sqlite3", appDb)
	_ = db.MustExec(fmt.Sprintf("attach database '%s' as calibre;", calibreDb))
	return db
}

func (d database) GetReadBooks() []ReadBook {
	readBooks := []ReadBook{}
	err := d.db.Select(&readBooks, `
WITH ranked_books AS (
	SELECT
		b.id as book_id,
		b.title AS book_name,
		s.id AS series_id,
		b.series_index as book_series_index,
		i.val as anilist_id,
        ROW_NUMBER() OVER (PARTITION BY s.id ORDER BY b.series_index DESC) as rn
	FROM
		book_read_link brl
	LEFT JOIN
		calibre.books b ON b.id = brl.book_id
	LEFT JOIN
		calibre.books_series_link bsl ON bsl.book = b.id
	LEFT JOIN
		calibre.series s ON bsl.series = s.id
	LEFT JOIN
		calibre.identifiers i ON i.book = b.id AND i.type = 'anilist'
	WHERE
		brl.last_modified > datetime('now', '-30 day')
	AND
		brl.read_status = 1
) SELECT
	book_id,
	book_name,
	series_id,
	book_series_index,
	anilist_id
FROM
	ranked_books
WHERE
	rn = 1 OR series_id IS NULL
ORDER BY
	series_id, book_series_index DESC;
`)

	cobra.CheckErr(err)

	return readBooks
}

func (d database) GetSeries(book ReadBook) []SeriesBook {
	seriesBooks := []SeriesBook{}
	err := d.db.Select(&seriesBooks, `
SELECT 
    b.id AS book_id,
    b.title AS book_name,
    b.series_index AS book_series_index,
    i.val AS anilist_id,
    c.value AS chapters
FROM 
    books b
INNER JOIN 
    books_series_link bsl ON b.id = bsl.book
LEFT JOIN 
    identifiers i ON b.id = i.book AND i.type = 'anilist'
LEFT JOIN 
    custom_column_15 c ON b.id = c.book
WHERE 
    bsl.series = ?
	AND b.series_index <= ?
ORDER BY 
    b.series_index;
`, book.SeriesId, book.BookSeriesIndex)

	cobra.CheckErr(err)

	return seriesBooks

}

func (d *database) Close() error {
	d.db.Exec("detach database calibre;")
	return d.db.Close()
}

func NewDatabase() Database {
	appDb := viper.GetString("appDb")
	metadataDb := viper.GetString("metadataDb")
	return &database{
		db: dbConnection(appDb, metadataDb),
	}
}
