package internal

import (
	"fmt"
	"math"
	"time"

	"github.com/charmbracelet/log"
	"github.com/kaiserbh/anilistgo"
)

type sync struct {
	db      Database
	anilist Anilist
	dryrun  bool
}

type Sync interface {
	Close() error
	Sync() ([]string, error)
}

func (s sync) Close() error {
	return s.db.Close()
}

func (s sync) filterInvalid(books []ReadBook) map[int]ReadBook {
	filtered := make(map[int]ReadBook)
	for _, book := range books {
		anilistId := book.AnilistId
		if anilistId == 0 {
			log.Debug("SKIPPING: Invalid anilist id for", "book", book.BookName)
			continue
		}
		filtered[anilistId] = book
	}
	return filtered
}

func (s sync) getChapterCount(book ReadBook, series []SeriesBook) int {
	var chapterCount int
	currentBook := series[len(series)-1]
	inProgress := book.ReadStatus == STATUS_IN_PROGRESS
	volumes := series
	if inProgress {
		// remove last volume to process later
		volumes = series[:len(series)-1]
	}

	// Count up all the chapters in the volumes
	for _, item := range volumes {
		chapterCount += item.Chapters
	}

	if inProgress {
		chapters := float64(currentBook.Chapters)
		progress := float64(book.ProgressPercent / 100.0)
		estimatedChapter := int(math.Round(chapters * progress))
		chapterCount += estimatedChapter
		log.Debug("IN PROGRESS: ", "estimated", chapterCount, "progress", fmt.Sprintf("%.2f%%", book.ProgressPercent))
	}
	return chapterCount
}

type anilistChange struct {
	anilist        Anilist
	dryrun         bool
	status         string
	currentVolume  int
	currentChapter int
	media          *anilistgo.Media
	volume         int
	chapter        int
}

func (s sync) getAnilistInfo(book ReadBook, anilistId, volume, chapter int) anilistChange {
	entry := s.anilist.Get(anilistId)
	media := s.anilist.GetMediaInfo(anilistId)

	volume = s.anilist.NormaliseVolumes(media, volume)
	chapter = s.anilist.NormaliseChapters(media, chapter)

	status := "CURRENT"
	if s.anilist.IsCompleted(media, volume, chapter) && book.ReadStatus == STATUS_FINISHED {
		status = "COMPLETED"
	}

	return anilistChange{
		dryrun:         s.dryrun,
		anilist:        s.anilist,
		status:         status,
		currentVolume:  int(entry.ProgressVolumes),
		currentChapter: int(entry.Progress),
		media:          media,
		volume:         volume,
		chapter:        chapter,
	}
}

func (a anilistChange) UpdateVolumes() (bool, error) {
	old := a.currentVolume
	_new := a.volume
	if a.volume > a.currentVolume {
		log.Info("UPDATING:", "type", "volume", "from", old, "to", _new)
		if a.dryrun {
			log.Warn("DRYRUN: would update volume with", "chapter", _new, "status", a.status)
		} else {
			err := a.anilist.UpdateVolumes(a.media, _new, a.status)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}
	log.Debug("SKIPPING: volume update", "current", old, "new", _new)
	return false, nil
}

func (a anilistChange) UpdateChapters() (bool, error) {
	old := a.currentChapter
	_new := a.chapter
	if _new > old {
		log.Info("UPDATING:", "type", "chapter", "from", old, "to", _new)
		if a.dryrun {
			log.Warn("DRYRUN: would update chapter with", "chapter", _new, "status", a.status)
		} else {
			err := a.anilist.UpdateChapters(a.media, _new, a.status)
			if err != nil {
				return false, err
			}
		}
		return true, nil
	}
	log.Debug("SKIPPING: chapter update", "current", a.currentChapter, "new", a.chapter)
	return false, nil
}

func (s sync) Sync() ([]string, error) {
	books := s.db.GetReadBooks()
	var updated []string
	filteredBooks := s.filterInvalid(books)

	for anilistId, book := range filteredBooks {
		log.Info("PROCESSING: ", "title", book.BookName, "book_id", book.BookId, "volume", book.BookSeriesIndex, "read_status", book.ReadStatus)

		seriesBooks := s.db.GetSeries(book)
		if len(seriesBooks) == 0 {
			log.Error("Couldn't find any books in series for", "book", book.BookId, "series_id", book.SeriesId)
			continue
		}

		currentBook := seriesBooks[len(seriesBooks)-1]
		latestVolume := currentBook.BookSeriesIndex
		chapterCount := s.getChapterCount(book, seriesBooks)

		change := s.getAnilistInfo(book, anilistId, latestVolume, chapterCount)

		volumeUpdated, err := change.UpdateVolumes()
		if err != nil {
			return updated, err
		}
		chapterUpdated, err := change.UpdateChapters()
		if err != nil {
			return updated, err
		}

		if volumeUpdated || chapterUpdated {
			updated = append(updated, book.BookName)
		}
		time.Sleep(1 * time.Second)
	}
	return updated, nil
}

func NewSync(dryrun bool) Sync {
	if dryrun {
		log.Warn("DRYRUN MODE")
	}
	return &sync{
		db:      NewDatabase(),
		anilist: NewAnilist(),
		dryrun:  dryrun,
	}
}
