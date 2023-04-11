package activity

import (
	"github.com/dlshle/wflow/pkg/store"
	"github.com/dlshle/wflow/proto"
)

type Handler interface {
	Get(id string) (*proto.Activity, error)
	Put(activity *proto.Activity) (*proto.Activity, error)
	TxPut(tx store.SQLTransactional, activity *proto.Activity) (*proto.Activity, error)
}

type handler struct {
	store Store
}

func (h *handler) Get(id string) (*proto.Activity, error) {
	return h.store.Get(id)
}

func (h *handler) Put(activity *proto.Activity) (*proto.Activity, error) {
	return h.store.Put(activity)
}

func (h *handler) TxPut(tx store.SQLTransactional, activity *proto.Activity) (*proto.Activity, error) {
	return h.store.TxPut(tx, activity)
}