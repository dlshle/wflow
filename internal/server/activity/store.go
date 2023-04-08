package activity

import (
	"github.com/dlshle/wflow/pkg/store"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type Store interface {
	Get(id string) (*proto.Activity, error)
	TxGet(tx store.SQLTransactional, id string) (*proto.Activity, error)
	TxPut(tx store.SQLTransactional, activity *proto.Activity) (*proto.Activity, error)
	TxDelete(tx store.SQLTransactional, id string) error
}

type activityStore struct {
	pbEntityStore store.PBEntityStore
}

func NewStore(connStr string) (Store, error) {
	pbEntityStore, err := store.Open(connStr, "activities")
	if err != nil {
		return nil, err
	}
	return &activityStore{
		pbEntityStore: pbEntityStore,
	}, nil
}

func (s *activityStore) Get(id string) (*proto.Activity, error) {
	return s.TxGet(s.pbEntityStore.GetDB(), id)
}

func (s *activityStore) TxGet(tx store.SQLTransactional, id string) (*proto.Activity, error) {
	pbEntity, err := s.pbEntityStore.TxGet(tx, id)
	if err != nil {
		return nil, err
	}
	activity := &proto.Activity{}
	err = gproto.Unmarshal(pbEntity.Payload, activity)
	return activity, err
}

func (s *activityStore) TxPut(tx store.SQLTransactional, activity *proto.Activity) (*proto.Activity, error) {
	activityData, err := gproto.Marshal(activity)
	if err != nil {
		return nil, err
	}
	updatedPBEntity, err := s.pbEntityStore.TxPut(tx, &store.PBEntity{activity.Id, activityData})
	if err != nil {
		return nil, err
	}
	activity.Id = updatedPBEntity.ID
	return activity, nil
}

func (s *activityStore) TxDelete(tx store.SQLTransactional, id string) error {
	return s.pbEntityStore.TxDelete(tx, id)
}
