package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// workerCmd represents the base command when called without any subcommands
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "start worker with activities",
	Args:  cobra.NoArgs,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viper.GetStringMapString("activity"))
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	workerCmd.Flags().String("address", "127.0.0.1:7233", "The host and port (formatted as host:port) for the Temporal Frontend Service. [$TEMPORAL_ADDRESS]")
	viper.MustBindEnv("address", "TEMPORAL_ADDRESS")
	workerCmd.Flags().StringP("namespace", "n", "default", "Identifies a Namespace in the Temporal Workflow. [$TEMPORAL_NAMESPACE]")
	viper.MustBindEnv("namespace", "TEMPORAL_NAMESPACE")
	workerCmd.Flags().StringP("task-queue", "t", "", "Task Queue. [$TEMPORAL_TASK_QUEUE]")
	viper.MustBindEnv("task-queue", "TEMPORAL_TASK_QUEUE")
	workerCmd.Flags().StringToStringP("activity", "a", nil, "Mapping activity name to shell command.")

	if err := viper.BindPFlags(workerCmd.Flags()); err != nil {
		panic(fmt.Sprintf("error while binding pflags: %v", err))
	}
}
