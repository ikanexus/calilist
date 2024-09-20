package anilist

import (
	"github.com/kaiserbh/anilistgo"
	"github.com/spf13/viper"
)

const url = "https://graphql.anilist.co"

type Anilist interface {
	Get(anilistId int) *anilistgo.MediaListEntry
	UpdateVolumes(entry *anilistgo.MediaListEntry, volume int, status string) error
	UpdateChapters(entry *anilistgo.MediaListEntry, chapters int, status string) error
	NormaliseChapters(entry *anilistgo.MediaListEntry, chapters int) int
	NormaliseVolumes(entry *anilistgo.MediaListEntry, volume int) int
	IsCompleted(entry *anilistgo.MediaListEntry, volumes, chapters int) bool
}

type anilist struct {
	token string
}

func (a anilist) Get(anilistId int) *anilistgo.MediaListEntry {
	q := anilistgo.NewUserMediaListQuery()
	q.GetUserMediaList(anilistId, a.token)
	return q
}

func (a anilist) NormaliseVolumes(entry *anilistgo.MediaListEntry, volume int) int {
	maxVolumes := entry.Media.Volumes
	if maxVolumes != 0 && volume >= maxVolumes {
		volume = maxVolumes
	}
	return volume
}

func (a anilist) NormaliseChapters(entry *anilistgo.MediaListEntry, chapters int) int {
	maxChapters := entry.Media.Chapters
	if maxChapters != 0 && chapters >= maxChapters {
		chapters = maxChapters
	}
	return chapters
}

func (a anilist) IsCompleted(entry *anilistgo.MediaListEntry, volumes, chapters int) bool {
	maxVolumes := entry.Media.Volumes
	maxChapters := entry.Media.Chapters
	if maxVolumes == 0 || maxChapters == 0 {
		return false
	}
	if chapters >= maxChapters && volumes >= maxVolumes {
		return true
	}
	return false
}

func (a anilist) UpdateVolumes(entry *anilistgo.MediaListEntry, volume int, status string) error {
	ok, err := anilistgo.SaveMediaListEntryProgressVolume(int(entry.MediaID), status, volume, a.token)
	if ok {
		return nil
	}
	return err
}

func (a anilist) UpdateChapters(entry *anilistgo.MediaListEntry, chapters int, status string) error {
	ok, err := anilistgo.SaveMediaListEntry(int(entry.MediaID), status, chapters, a.token)
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
