package api

import (
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/internal/client/job"
	"github.com/dlshle/wflow/internal/client/message_processors"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
)

type WorkerClient interface {
	Start() error
	Close() error
}

type workerClient struct {
	workerID   string
	jobManager job.JobManager
	tcpClient  protocol.TCPClient
}

func New(workerID, address string, port int, workerActivities []activity.WorkerActivity) (*workerClient, error) {
	var err error
	// initialize activity uuids by names, activity name can't be empty
	workerActivities, err = initializeWorkerActivities(workerActivities)
	if err != nil {
		return nil, err
	}
	jobManager := job.New(workerID, buildWorkerActivitiesMap(workerActivities))
	messageHandler := protocol.NewMessageHandler(map[proto.Type][]protocol.MessageProcessor{
		proto.Type_CANCEL_JOB:          {message_processors.CreateCancelJobProcessor(jobManager)},
		proto.Type_DISPATCH_JOB:        {message_processors.CreateDispatchJobProcessor(jobManager)},
		proto.Type_QUERY_JOB:           {message_processors.CreateQueryJobProcessor(jobManager)},
		proto.Type_QUERY_WORKER_STATUS: {message_processors.CreateQueryWorkerStatusProcessor(workerID, jobManager)},
	})
	tcpClient := protocol.NewTCPClient(workerID, address, port, messageHandler, jobManager.SupportedActivities(), func(c protocol.ServerConnection) {
		jobManager.InitReportingServer(c)
	})
	return &workerClient{
		workerID:   workerID,
		jobManager: jobManager,
		tcpClient:  tcpClient,
	}, nil
}

func initializeWorkerActivities(activities []activity.WorkerActivity) ([]activity.WorkerActivity, error) {
	for i := range activities {
		rawActivity := activities[i].Activity()
		if rawActivity.Name == "" {
			return nil, errors.Error("activity name is empty")
		}
		rawActivity.Id = utils.GenerateUUIDOnString(rawActivity.Name)
		activities[i] = activity.NewWorkerActivity(rawActivity, activities[i].Handler())
	}
	return activities, nil
}

func (c *workerClient) Start() error {
	serverConn, err := c.tcpClient.Connect()
	if err != nil {
		return err
	}
	return c.jobManager.InitReportingServer(serverConn)
}

func (c *workerClient) Close() error {
	return c.tcpClient.Close()
}

func buildWorkerActivitiesMap(workerActivities []activity.WorkerActivity) map[string]activity.WorkerActivity {
	activityMap := make(map[string]activity.WorkerActivity)
	for _, wa := range workerActivities {
		activityMap[wa.Activity().Id] = wa
	}
	return activityMap
}
