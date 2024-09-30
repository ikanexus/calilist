package internal

import (
	"github.com/charmbracelet/log"
	"github.com/kaiserbh/anilistgo"
	"github.com/spf13/viper"
)

type Anilist interface {
	Get(anilistId int) *anilistgo.MediaListEntry
	GetMediaInfo(anilistId int) *anilistgo.Media
	UpdateVolumes(entry *anilistgo.Media, volume int, status string) error
	UpdateChapters(entry *anilistgo.Media, chapters int, status string) error
	NormaliseChapters(entry *anilistgo.Media, chapters int) int
	NormaliseVolumes(entry *anilistgo.Media, volume int) int
	IsCompleted(entry *anilistgo.Media, volumes, chapters int) bool
}

type anilist struct {
	token string
}

func (a anilist) Get(anilistId int) *anilistgo.MediaListEntry {
	q := anilistgo.NewUserMediaListQuery()
	q.GetUserMediaList(anilistId, a.token)
	return q
}

func (a anilist) GetMediaInfo(anilistId int) *anilistgo.Media {
	q := anilistgo.NewMediaQuery()
	q.FilterMangaByID(anilistId)
	return q
}

func (a anilist) NormaliseVolumes(entry *anilistgo.Media, volume int) int {
	maxVolumes := entry.Volumes
	if maxVolumes != 0 && volume > maxVolumes {
		log.Debug("Normalised volume", "old", volume, "new", maxVolumes)
		volume = maxVolumes
	}
	return volume
}

func (a anilist) NormaliseChapters(entry *anilistgo.Media, chapters int) int {
	maxChapters := entry.Chapters
	if maxChapters != 0 && chapters > maxChapters {
		log.Debug("Normalised chapters", "old", chapters, "new", maxChapters)
		chapters = maxChapters
	}
	return chapters
}

func (a anilist) IsCompleted(entry *anilistgo.Media, volumes, chapters int) bool {
	maxVolumes := entry.Volumes
	maxChapters := entry.Chapters
	log.Debugf("max volumes: %d, max chapters: %d", maxVolumes, maxChapters)
	if maxVolumes == 0 || maxChapters == 0 {
		return false
	}
	if chapters >= maxChapters && volumes >= maxVolumes {
		return true
	}
	return false
}

func (a anilist) UpdateVolumes(entry *anilistgo.Media, volume int, status string) error {
	ok, err := anilistgo.SaveMediaListEntryProgressVolume(int(entry.ID), status, volume, a.token)
	if ok {
		return nil
	}
	return err
}

func (a anilist) UpdateChapters(entry *anilistgo.Media, chapters int, status string) error {
	ok, err := anilistgo.SaveMediaListEntry(int(entry.ID), status, chapters, a.token)
	if ok {
		return nil
	}
	return err
}

func NewAnilist() Anilist {
	token := viper.GetString("token")
	return &anilist{
		token: token,
	}
}
