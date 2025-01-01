package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "temporal-shell",
	Short: "register shell activities for temporal",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viper.GetStringMapString("activity"))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().String("address", "127.0.0.1:7233", "The host and port (formatted as host:port) for the Temporal Frontend Service. [$TEMPORAL_ADDRESS]")
	viper.MustBindEnv("address", "TEMPORAL_ADDRESS")
	rootCmd.PersistentFlags().StringP("namespace", "n", "default", "Task Queue. [$TEMPORAL_NAMESPACE]")
	viper.MustBindEnv("address", "TEMPORAL_NAMESPACE")
	rootCmd.PersistentFlags().StringP("task-queue", "t", "", "Task Queue. [$TEMPORAL_TASK_QUEUE]")
	viper.MustBindEnv("address", "TEMPORAL_TASK_QUEUE")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(fmt.Sprintf("error while binding pflags: %v", err))
	}

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringToStringP("activity", "a", nil, "Activity")

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		panic(fmt.Sprintf("error while binding pflags: %v", err))
	}
}
