/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/charmbracelet/log"
	"github.com/ikanexus/calilist/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync calibre database with anilist",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Starting Sync")
		dryrun, _ := cmd.Flags().GetBool("dryrun")
		sync := internal.NewSync(dryrun)
		defer sync.Close()
		updated, err := sync.Sync()
		if len(updated) == 0 {
			log.Info("REPORT: No updates required")
		} else {
			log.Infof("REPORT: Successfully Updated %v", updated)
		}
		return err
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
