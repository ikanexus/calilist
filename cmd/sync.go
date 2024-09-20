/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

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
		fmt.Println("syncing")
		db := database.NewDatabase()
		defer db.Close()
		al := anilist.NewAnilist()

		books := db.GetReadBooks()
		for _, book := range books {
			anilistId := book.AnilistId
			fmt.Printf("book: %v\n", book)
			seriesBooks := db.GetSeries(book)

			// fmt.Printf("books in series: %v\n", seriesBooks)
			latestVolume := seriesBooks[len(seriesBooks)-1].BookSeriesIndex
			chapterCount := 0
			for _, item := range seriesBooks {
				chapterCount += item.Chapters
			}

			anilistEntry := al.Get(anilistId)
			currentVolume := int(anilistEntry.ProgressVolumes)
			currentChapter := int(anilistEntry.Progress)

			newVolume := al.NormaliseVolumes(anilistEntry, latestVolume)
			newChapter := al.NormaliseChapters(anilistEntry, chapterCount)

			readingStatus := anilistEntry.Status
			if al.IsCompleted(anilistEntry, newVolume, newChapter) {
				readingStatus = "COMPLETED"
				fmt.Printf("Media is completed, changing status to %s\n", readingStatus)
			}

			fmt.Printf("Normalised volume: %d, chapter: %d => volume: %d, chapter: %d\n", latestVolume, chapterCount, newVolume, newChapter)

			if newVolume > currentVolume {
				fmt.Printf("Updating latest volume to %d from %d\n", newVolume, currentVolume)
				err := al.UpdateVolumes(anilistEntry, newVolume, readingStatus)
				cobra.CheckErr(err)
			} else {
				fmt.Printf("Current volume %d is greater than or equal to the new volume %d, skipping.\n", currentVolume, newVolume)
			}

			if newChapter > currentChapter {
				fmt.Printf("Updating chapter count to %d from %d\n", newChapter, currentChapter)
				err := al.UpdateChapters(anilistEntry, newChapter, readingStatus)
				cobra.CheckErr(err)
			} else {
				fmt.Printf("Current chapter count %d is greater than or equal to the new chapter count %d, skipping.\n", currentChapter, newChapter)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().String("appDb", "", "calibre-web app.db location")
	viper.BindPFlag("appDb", syncCmd.Flags().Lookup("appDb"))
	syncCmd.Flags().String("metadataDb", "", "calibre metadata.db location")
	viper.BindPFlag("metadataDb", syncCmd.Flags().Lookup("metadataDb"))
}
