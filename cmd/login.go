/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Anilist and store in config",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Open the following URL in your browser and paste in the token")
		fmt.Println("https://anilist.co/api/v2/oauth/authorize?client_id=21288&response_type=token")
		fmt.Print("> ")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		token := input.Text()
		viper.Set("token", token)
		err := viper.SafeWriteConfig()
		cobra.CheckErr(err)
		fmt.Println("Saved")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
