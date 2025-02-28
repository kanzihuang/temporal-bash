package worker

import (
	"context"
	"crypto/tls"
	"github.com/google/uuid"
	"github.com/kanzihuang/temporal-bash/internal/bash"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	ch := make(chan interface{}, 1)
	g, ctx := errgroup.WithContext(context.Background())

	hostWorker := worker.New(c, hostTaskQueue, worker.Options{DisableWorkflowWorker: true})
	hostWorker.RegisterActivity(activities)
	for name, command := range activityMap {
		hostWorker.RegisterActivityWithOptions(bash.BuildBash(command), activity.RegisterOptions{Name: name})
	}
	g.Go(func() error {
		return hostWorker.Run(ch)
	})

	routeWorker := worker.New(c, taskQueue, worker.Options{DisableWorkflowWorker: true})
	routeWorker.RegisterActivity(activities.Begin)
	g.Go(func() error {
		return routeWorker.Run(ch)
	})

	go func() {
		select {
		case <-sig:
			close(ch)
		case <-ctx.Done():
			close(ch)
		}
	}()
	return g.Wait()
}
