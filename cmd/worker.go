package cmd

import (
	"errors"
	"fmt"
	"github.com/kanzihuang/temporal-bash/internal/worker"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		taskQueue := viper.GetString("task-queue")
		if len(taskQueue) == 0 {
			return errors.New("task-queue is required")
		}
		activityMap := viper.GetStringMapString("activity")
		if len(activityMap) == 0 {
			return errors.New("activity is required")
		}
		return worker.Run(
			viper.GetString("address"),
			viper.GetString("namespace"),
			viper.GetBool("tls"),
			taskQueue,
			activityMap)
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
	workerCmd.Flags().Bool("tls", false, "Enable base TLS encryption. Does not have additional options like mTLS or client certs.. [$TEMPORAL_TLS]")
	viper.MustBindEnv("tls", "TEMPORAL_TLS")
	workerCmd.Flags().StringP("task-queue", "t", "", "Task Queue. [$TEMPORAL_TASK_QUEUE]")
	viper.MustBindEnv("task-queue", "TEMPORAL_TASK_QUEUE")
	workerCmd.Flags().StringToStringP("activity", "a", nil, "Mapping activity name to bash command.")

	if err := viper.BindPFlags(workerCmd.Flags()); err != nil {
		panic(fmt.Sprintf("error while binding pflags: %v", err))
	}
}
