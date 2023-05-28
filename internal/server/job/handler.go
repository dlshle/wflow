package job

import (
	relationmapping "github.com/dlshle/wflow/internal/server/relation_mapping"
	"github.com/dlshle/wflow/proto"
)

type Handler interface {
	ListJobIDsByActivityID(activityID string) ([]string, error)
	Get(id string) (*proto.JobReport, error)
	Put(jobReport *proto.JobReport) (*proto.JobReport, error)
}

type jobHandler struct {
	store                  Store
	relationMappingHandler relationmapping.Handler
}

func NewHandler(store Store, relationMappingHandler relationmapping.Handler) Handler {
	return &jobHandler{store, relationMappingHandler}
}

func (h *jobHandler) Get(id string) (*proto.JobReport, error) {
	return h.store.Get(id)
}

func (h *jobHandler) Put(jobReport *proto.JobReport) (*proto.JobReport, error) {
	return h.store.Put(jobReport)
}

func (h *jobHandler) ListJobIDsByActivityID(activityID string) ([]string, error) {
	return h.store.ListJobIDsByActivityID(activityID)
}
