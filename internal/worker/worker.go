package worker

import (
	"github.com/google/uuid"
	"github.com/kanzihuang/temporal-shell/internal/shell"
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

	hostTaskQueue := taskQueue + "-" + uuid.Must(uuid.NewV7()).String()
	activities := shell.NewActivities(hostTaskQueue)

	hostWorker := worker.New(c, hostTaskQueue, worker.Options{DisableWorkflowWorker: true})
	hostWorker.RegisterActivity(activities)
	for name, command := range activityMap {
		hostWorker.RegisterActivityWithOptions(shell.BuildBash(command), activity.RegisterOptions{Name: name})
	}
	if err := hostWorker.Start(); err != nil {
		return err
	}

	routeWorker := worker.New(c, taskQueue, worker.Options{DisableWorkflowWorker: true})
	routeWorker.RegisterActivity(activities.Begin)
	if err := routeWorker.Start(); err != nil {
		return err
	}

	<-worker.InterruptCh()
	routeWorker.Stop()
	hostWorker.Stop()
	return nil
}
