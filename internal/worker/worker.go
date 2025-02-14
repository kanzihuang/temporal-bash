package worker

import (
	"crypto/tls"
	"github.com/google/uuid"
	"github.com/kanzihuang/temporal-bash/internal/bash"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func Run(address string, namespace string, useTls bool, taskQueue string, activityMap map[string]string) error {
	opts := client.Options{
		HostPort:  address,
		Namespace: namespace,
	}
	if useTls {
		opts.ConnectionOptions.TLS = &tls.Config{}
	}
	c, err := client.Dial(opts)
	if err != nil {
		return err
	}
	defer c.Close()

	hostTaskQueue := taskQueue + "-" + uuid.Must(uuid.NewV7()).String()
	activities := bash.NewActivities(hostTaskQueue)

	hostWorker := worker.New(c, hostTaskQueue, worker.Options{DisableWorkflowWorker: true})
	hostWorker.RegisterActivity(activities)
	for name, command := range activityMap {
		hostWorker.RegisterActivityWithOptions(bash.BuildBash(command), activity.RegisterOptions{Name: name})
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
