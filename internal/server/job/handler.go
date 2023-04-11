package job

import (
	relationmapping "github.com/dlshle/wflow/internal/server/relation_mapping"
	"github.com/dlshle/wflow/proto"
)

type Handler interface {
	Get(id string) (*proto.JobReport, error)
	Put(jobReport *proto.JobReport) error
}

type jobHandler struct {
	store                  Store
	relationMappingHandler relationmapping.Handler
}

func (h *jobHandler) Get(id string) (*proto.JobReport, error) {
	return h.store.Get(id)
}

func (h *jobHandler) Put(jobReport *proto.JobReport) (err error) {
	_, err = h.store.Put(jobReport)
	return
}
