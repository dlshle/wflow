package activity

import (
	"github.com/dlshle/wflow/pkg/store"
	"github.com/dlshle/wflow/proto"
)

type Handler interface {
	Get(id string) (*proto.Activity, error)
	Put(activity *proto.Activity) (*proto.Activity, error)
	TxPut(tx store.SQLTransactional, activity *proto.Activity) (*proto.Activity, error)
	GetAll() ([]*proto.Activity, error)
	TxGetAll(tx store.SQLTransactional) ([]*proto.Activity, error)
}

func NewHandler(store Store) Handler {
	return &handler{store}
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

func (h *handler) GetAll() ([]*proto.Activity, error) {
	return h.store.GetAll()
}

func (h *handler) TxGetAll(tx store.SQLTransactional) ([]*proto.Activity, error) {
	return h.store.TxGetAll(tx)
}
