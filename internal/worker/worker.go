package worker

import (
	"github.com/google/uuid"
	"github.com/kanzihuang/temporal-shell/internal/shell"
	"github.com/kanzihuang/temporal-shell/pkg/common"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func Run(address string, namespace string, taskQueue string, activityMap map[string]string) error {
	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	hostTaskQueue := uuid.Must(uuid.NewV7()).String()

	hostWorker := worker.New(c, hostTaskQueue, worker.Options{})
	hostWorker.RegisterActivityWithOptions(shell.ReadFile, activity.RegisterOptions{Name: common.ReadFile})
	for name, command := range activityMap {
		hostWorker.RegisterActivityWithOptions(shell.BuildExecute(command), activity.RegisterOptions{Name: name})
	}
	if err := hostWorker.Start(); err != nil {
		return err
	}

	routeWorker := worker.New(c, taskQueue, worker.Options{})
	routeWorker.RegisterActivityWithOptions(shell.BuildGetHostTaskQueue(hostTaskQueue),
		activity.RegisterOptions{Name: common.GetHostTaskQueue})
	if err := routeWorker.Start(); err != nil {
		return err
	}

	<-worker.InterruptCh()
	routeWorker.Stop()
	hostWorker.Stop()
	return nil
}
