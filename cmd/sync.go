/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"math"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ikanexus/calilist/internal/anilist"
	"github.com/ikanexus/calilist/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync calibre database with anilist",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting Sync")
		dryrun, _ := cmd.Flags().GetBool("dryrun")
		if dryrun {
			log.Warn("DRYRUN MODE")
		}
		db := database.NewDatabase()
		defer db.Close()
		al := anilist.NewAnilist()

		books := db.GetReadBooks()
		updated := 0
		for _, book := range books {
			anilistId := book.AnilistId
			if anilistId == 0 {
				log.Error("Invalid anilist id for", "book", book.BookName)
			}
			log.Info("Processing read book", "title", book.BookName, "book_id", book.BookId, "series_id", book.SeriesId, "volume", book.BookSeriesIndex, "read_status", book.ReadStatus)
			seriesBooks := db.GetSeries(book)

			// fmt.Printf("books in series: %v\n", seriesBooks)
			if len(seriesBooks) == 0 {
				log.Error("Couldn't find any books in series for", "book", book.BookId, "series_id", book.SeriesId)
				continue
			}
			currentBook := seriesBooks[len(seriesBooks)-1]
			latestVolume := currentBook.BookSeriesIndex
			chapterCount := 0
			for _, item := range seriesBooks {
				chapterCount += item.Chapters
			}
			if book.ReadStatus == database.STATUS_IN_PROGRESS {
				currentBookChapters := currentBook.Chapters
				totalChapters := chapterCount
				chapterCount -= currentBookChapters
				progressPercent := float64(book.ProgressPercent / 100.0)
				estimatedChapterProgress := int(math.Round(float64(currentBookChapters) * progressPercent))
				chapterCount += estimatedChapterProgress
				log.Info("Current chapter", "estimated", chapterCount, "total", totalChapters, "progress", progressPercent)
			}

			anilistEntry := al.Get(anilistId)
			currentVolume := int(anilistEntry.ProgressVolumes)
			currentChapter := int(anilistEntry.Progress)

			newVolume := al.NormaliseVolumes(anilistEntry, latestVolume)
			newChapter := al.NormaliseChapters(anilistEntry, chapterCount)

			readingStatus := "CURRENT"
			if al.IsCompleted(anilistEntry, newVolume, newChapter) && book.ReadStatus == database.STATUS_FINISHED {
				readingStatus = "COMPLETED"
				log.Debug("Media is completed, changing status", "status", readingStatus)
			}

			entryTitle := book.BookName

			if newVolume > currentVolume {
				log.Info("Updating latest volume for", "book", entryTitle, "from", currentVolume, "to", newVolume)
				if dryrun {
					log.Warn("DRYRUN: would update volume with", "volume", newVolume, "status", readingStatus)
				} else {
					err := al.UpdateVolumes(anilistEntry, newVolume, readingStatus)
					cobra.CheckErr(err)
				}
				updated += 1
			} else {
				log.Debug("Skipping volume update - current >= new", "current", currentVolume, "new", newVolume)
			}

			if newChapter > currentChapter {
				log.Info("Updating chapter count for", "book", entryTitle, "from", currentChapter, "to", newChapter)
				if dryrun {
					log.Warn("DRYRUN: would update chapters with", "chapter", newChapter, "status", readingStatus)
				} else {
					err := al.UpdateChapters(anilistEntry, newChapter, readingStatus)
					cobra.CheckErr(err)
				}
				updated += 1
			} else {
				log.Debug("Skipping chapter update - current >= new", "current", currentChapter, "new", newChapter)
			}
			time.Sleep(1 * time.Second)
		}
		if updated == 0 {
			log.Info("No updates required")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().String("appDb", "", "calibre-web app.db location")
	viper.BindPFlag("appDb", syncCmd.Flags().Lookup("appDb"))
	syncCmd.Flags().String("metadataDb", "", "calibre metadata.db location")
	viper.BindPFlag("metadataDb", syncCmd.Flags().Lookup("metadataDb"))
	syncCmd.Flags().Bool("dryrun", false, "Perform a dryrun")
}
