package api

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	timerx "github.com/dlshle/gommon/timer"
	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/internal/client/job"
	"github.com/dlshle/wflow/internal/client/message_processors"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
)

const (
	DefaultMaxPoolSize            = 2048
	DefaultMaxAsyncPoolWorkerSize = 512
)

type WorkerClient interface {
	Start() error
	Close() error
}

type workerClient struct {
	ctx        context.Context
	workerID   string
	jobManager job.JobManager
	logger     logging.Logger
	tcpClient  protocol.TCPClient
	waitLock   *async.WaitLock
}

func New(address string, port int, workerActivities []activity.WorkerActivity) (*workerClient, error) {
	var err error
	var timer timerx.Timer
	// initialize activity uuids by names, activity name can't be empty
	workerActivities, err = initializeWorkerActivities(workerActivities)
	if err != nil {
		return nil, err
	}
	workerID, err := generateWorkerID()
	if err != nil {
		return nil, err
	}
	workerName, err := generateWorkerName()
	if err != nil {
		return nil, err
	}
	asyncPool := async.NewAsyncPool(workerID, DefaultMaxPoolSize, DefaultMaxAsyncPoolWorkerSize)
	jobManager := job.New(workerID, workerName, buildWorkerActivitiesMap(workerActivities))
	messageHandler := protocol.NewMessageHandler(map[proto.Type][]protocol.MessageProcessor{
		proto.Type_CANCEL_JOB:          {message_processors.CreateCancelJobProcessor(jobManager)},
		proto.Type_DISPATCH_JOB:        {message_processors.CreateDispatchJobProcessor(jobManager)},
		proto.Type_QUERY_JOB:           {message_processors.CreateQueryJobProcessor(jobManager)},
		proto.Type_QUERY_WORKER_STATUS: {message_processors.CreateQueryWorkerStatusProcessor(workerID, jobManager)},
	})
	tcpClient := protocol.NewTCPClient(workerID,
		address,
		port,
		asyncPool,
		messageHandler,
		jobManager.SupportedActivities(),
		func(c protocol.ServerConnection, s *proto.Server) {
			if timer != nil {
				timer.Stop()
			}
			err := jobManager.InitReportingServer(c, s)
			if err != nil {
				panic(err)
			}
			timer = timerx.New(time.Minute, func() {
				err := jobManager.InitReportingServer(c, s)
				if err != nil {
					logging.GlobalLogger.Warnf(context.Background(), "failed to init reporting server: %v", err)
				}
			})
			timer.Repeat()
		})
	return &workerClient{
		ctx:        context.Background(),
		workerID:   workerID,
		jobManager: jobManager,
		tcpClient:  tcpClient,
		waitLock:   async.NewWaitLock(),
		logger:     logging.GlobalLogger.WithPrefix("[Worker]"),
	}, nil
}

func generateWorkerID() (workerID string, err error) {
	var (
		macAddr  string
		hostName string
	)
	macAddr, err = utils.GetMACAddress()
	if err != nil {
		return
	}
	hostName, err = os.Hostname()
	if err != nil {
		return
	}
	workerID = utils.GenerateUUIDOnString(macAddr + hostName + strconv.Itoa(os.Getpid()))
	return
}

func generateWorkerName() (workerName string, err error) {
	var (
		macAddr  string
		hostName string
	)
	macAddr, err = utils.GetMACAddress()
	if err != nil {
		return
	}
	hostName, err = os.Hostname()
	if err != nil {
		return
	}
	workerName = hostName + ":" + macAddr + ":" + strconv.Itoa(os.Getpid())
	return
}

func initializeWorkerActivities(activities []activity.WorkerActivity) ([]activity.WorkerActivity, error) {
	for i := range activities {
		rawActivity := activities[i].Activity()
		if rawActivity.Name == "" {
			return nil, errors.Error("activity name is empty")
		}
		if rawActivity.Id == "" {
			rawActivity.Id = utils.GenerateUUIDOnString(rawActivity.Name)
		}
		activities[i] = activity.NewWorkerActivity(rawActivity, activities[i].Handler())
	}
	return activities, nil
}

func (c *workerClient) Start() error {
	serverConn, err := c.tcpClient.Connect()
	if err != nil {
		return err
	}
	c.logger.Infof(c.ctx, "server %s connected", serverConn.Address())
	return c.jobManager.InitReportingServer(serverConn, c.tcpClient.ServerInfo())
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
