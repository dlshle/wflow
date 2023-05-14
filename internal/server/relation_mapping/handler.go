package relationmapping

import (
	"context"

	"github.com/dlshle/gommon/errors"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/server/activity"
	"github.com/dlshle/wflow/pkg/store"
	"github.com/dlshle/wflow/proto"
)

type Handler interface {
	FindJobsByWorkerID(workerID string) ([]*proto.JobReport, error)
	FindWorkersByActivityID(activityID string) ([]*proto.Worker, error)
	ListAllActiveActivities() ([]*proto.Activity, error)
	TxUpdateWorkerJobs(ctx context.Context, tx store.SQLTransactional, worker *proto.Worker) error
	TxUpdateWorkerActivities(ctx context.Context, tx store.SQLTransactional, worker *proto.Worker) error
}

type handler struct {
	logger        logging.Logger
	store         *relationMappingStore // this maintains worker-activity relation to tell which worker can handle which activity at the moment
	activityStore activity.Store
}

func NewHandler(store *relationMappingStore, activityStore activity.Store) Handler {
	return &handler{
		logger:        logging.GlobalLogger.WithPrefix("[RelationMappingHandler]"),
		store:         store,
		activityStore: activityStore,
	}
}

func (h *handler) ListAllActiveActivities() ([]*proto.Activity, error) {
	return h.store.ListAllActiveActivities()
}

func (h *handler) FindJobsByWorkerID(workerID string) ([]*proto.JobReport, error) {
	return h.store.TxFindJobsByWorkerID(h.store.db, workerID)
}

func (h *handler) FindWorkersByActivityID(activityID string) ([]*proto.Worker, error) {
	return h.store.TxFindWorkersByActivityID(h.store.db, activityID)
}

func (h *handler) TxUpdateWorkerJobs(ctx context.Context, tx store.SQLTransactional, worker *proto.Worker) error {
	h.logger.Info(ctx, "update worker jobs for "+worker.Id)
	jobIDs, err := h.store.TxGetJobIDsByWorkerID(tx, worker.Id)
	if err != nil {
		return err
	}
	err = h.txPutActivitiesOnActivityStore(tx, worker.SupportedActivities)
	if err != nil {
		return errors.Error("failed to put activities for worker " + worker.Id + " due to " + err.Error())
	}
	set := transformStringSliceToMap(jobIDs)
	var toAdd []string
	for _, job := range worker.ActiveJobs {
		if set[job] {
			// to add
			toAdd = append(toAdd, job)
		}
	}
	h.logger.Infof(ctx, "[%s] to add jobs %v", worker.Id, toAdd)
	for _, jobID := range toAdd {
		exists, err := h.store.TxAddJobWorkerMapping(tx, jobID, worker.Id)
		if err != nil {
			h.logger.Errorf(ctx, "[%s] failed to add new job %s due to %s", worker.Id, jobID, err.Error())
			return err
		}
		if exists {
			h.logger.Warnf(ctx, "[%s] job %s already exists", worker.Id, jobID)
		}
	}
	return nil
}

func (h *handler) txPutActivitiesOnActivityStore(tx store.SQLTransactional, activities []*proto.Activity) error {
	for _, activity := range activities {
		_, err := h.activityStore.TxPut(tx, activity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *handler) TxUpdateWorkerActivities(ctx context.Context, tx store.SQLTransactional, worker *proto.Worker) error {
	h.logger.Info(ctx, "update worker activities for "+worker.Id)
	activitiesIDs, err := h.store.TxGetActivityIDsByWorkerID(tx, worker.Id)
	if err != nil {
		return err
	}
	toAdd, toDelete := computeActivitiesDiff(activitiesIDs, worker)
	h.logger.Infof(ctx, "[%s] to add activities %v, to delete activities %v", worker.Id, toAdd, toDelete)
	err = h.store.TxBulkDeleteActivityMappingsByWorkerID(tx, toDelete, worker.Id)
	if err != nil {
		h.logger.Errorf(ctx, "[%s] failed to delete activities due to %s", worker.Id, err.Error())
		return err
	}
	for _, newActivityID := range toAdd {
		exists, err := h.store.TxAddActivityWorkerMapping(tx, newActivityID, worker.Id)
		if err != nil {
			h.logger.Errorf(ctx, "[%s] failed to add new activity %s due to %s", worker.Id, newActivityID, err.Error())
			return err
		}
		if exists {
			h.logger.Warnf(ctx, "[%s] activity %s already exists", worker.Id, newActivityID)
		}
	}
	return nil
}

func computeActivitiesDiff(existingActivities []string, worker *proto.Worker) (added []string, deleted []string) {
	var toAdd []string
	var toDelete []string
	set := transformStringSliceToMap(existingActivities)
	for _, activity := range worker.SupportedActivities {
		if set[activity.Id] {
			// to add
			toAdd = append(toAdd, activity.Id)
		}
		delete(set, activity.Id)
	}
	for left := range set {
		toDelete = append(toDelete, left)
	}
	return toAdd, toDelete
}

func transformStringSliceToMap(slice []string) map[string]bool {
	m := make(map[string]bool)
	for _, ele := range slice {
		m[ele] = true
	}
	return m
}
