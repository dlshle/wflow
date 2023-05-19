package logging

import (
	"context"
	"sync"
	"time"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type WFlowLogWriter struct {
	serverConn protocol.ServerConnection
	ctx        context.Context
	jobID      string
	logs       []*proto.JobLog
	logger     logging.Logger
	mu         *sync.Mutex
}

func NewWFlowLogWriter(ctx context.Context, jobID string, serverConn protocol.ServerConnection) logging.LogWriter {
	w := &WFlowLogWriter{
		ctx:        ctx,
		serverConn: serverConn,
		jobID:      jobID,
		logs:       make([]*proto.JobLog, 0),
		mu:         new(sync.Mutex),
		logger:     logging.GlobalLogger.WithPrefix("[WFlowLogWriter]"),
	}
	go w.logUploadJob()
	return w
}

func (w *WFlowLogWriter) logUploadJob() {
	ticker := *time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			w.safeWriteLogs()
		case <-w.ctx.Done():
			w.safeWriteLogs()
			return
		}
	}
}

func (w *WFlowLogWriter) withLock(cb func()) {
	w.mu.Lock()
	cb()
	w.mu.Unlock()
}

func (w *WFlowLogWriter) Write(entity *logging.LogEntity) {
	w.mu.Lock()
	w.logs = append(w.logs, &proto.JobLog{JobId: w.jobID, Message: entity.Message, Level: mapLogLevel(entity.Level), Timestamp: int32(entity.Timestamp.UnixMilli() / 1000), Contexts: entity.Context})
	w.mu.Unlock()
	if len(w.logs) >= 10 {
		w.safeWriteLogs()
	}
}

func (w *WFlowLogWriter) safeWriteLogs() {
	copied := make([]*proto.JobLog, len(w.logs), len(w.logs))
	w.withLock(func() {
		copy(copied, w.logs)
		w.logs = make([]*proto.JobLog, 0)
	})
	w.tryUploadLogs(w.logs)
}

func (w *WFlowLogWriter) tryUploadLogs(logs []*proto.JobLog) {
	uploadLogsRequest := &proto.WrappedLogs{
		Logs: logs,
	}
	logsData, err := gproto.Marshal(uploadLogsRequest)
	if err != nil {
		w.logger.Errorf(context.Background(), "failed to marshal logs: %s", err.Error())
		return
	}
	err = w.doUploadLogsWithRetry(&proto.Message{
		Id:      utils.RandomUUID(),
		Type:    proto.Type_UPLOAD_LOGS,
		Payload: logsData,
	})
	if err != nil {
		w.logger.Errorf(context.Background(), "failed to upload logs: %s", err.Error())
	} else {
		w.logger.Infof(context.Background(), "successfully uploaded %d logs", len(logs))
	}
}

func (w *WFlowLogWriter) doUploadLogsWithRetry(r *proto.Message) (err error) {
	var resp *proto.Message
	for i := 0; i < 3; i++ {
		resp, err = w.serverConn.Request(r)
		if err == nil {
			return
		}
		if resp.Status != proto.Status_OK {
			err = errors.Error(string(resp.Payload))
		}
	}
	return

}

func mapLogLevel(originalLevel int) proto.LogLevel {
	switch originalLevel {
	case logging.DEBUG:
		return logging.DEBUG
	case logging.INFO:
		return logging.INFO
	case logging.WARN:
		return logging.WARN
	case logging.ERROR:
		return logging.ERROR
	case logging.FATAL:
		return logging.FATAL
	default:
		return logging.INFO
	}
}
