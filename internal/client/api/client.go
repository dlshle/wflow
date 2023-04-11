package api

import (
	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/internal/client/job"
	"github.com/dlshle/wflow/internal/client/message_processors"
	"github.com/dlshle/wflow/internal/protocol"
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
	jobManager := job.New(workerID, buildWorkerActivitiesMap(workerActivities))
	messageHandler := protocol.NewMessageHandler(map[proto.Type][]protocol.MessageProcessor{
		proto.Type_CANCEL_JOB:          {message_processors.CreateCancelJobProcessor(jobManager)},
		proto.Type_DISPATCH_JOB:        {message_processors.CreateDispatchJobProcessor(jobManager)},
		proto.Type_QUERY_JOB:           {message_processors.CreateQueryJobProcessor(jobManager)},
		proto.Type_QUERY_WORKER_STATUS: {message_processors.CreateQueryWorkerStatusProcessor(workerID, jobManager)},
	})
	tcpClient := protocol.NewTCPClient(workerID, address, port, messageHandler, jobManager.SupportedActivities())
	return &workerClient{
		workerID:   workerID,
		jobManager: jobManager,
		tcpClient:  tcpClient,
	}, nil
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
