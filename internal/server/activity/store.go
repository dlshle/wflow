package activity

import (
	"time"

	"github.com/dlshle/wflow/pkg/store"
	wutil "github.com/dlshle/wflow/pkg/utils"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type Store interface {
	Get(id string) (*proto.Activity, error)
	TxGet(tx store.SQLTransactional, id string) (*proto.Activity, error)
	Put(activity *proto.Activity) (*proto.Activity, error)
	TxPut(tx store.SQLTransactional, activity *proto.Activity) (*proto.Activity, error)
	TxDelete(tx store.SQLTransactional, id string) error
	GetAll() ([]*proto.Activity, error)
	TxGetAll(tx store.SQLTransactional) ([]*proto.Activity, error)
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

func (s *activityStore) Put(activity *proto.Activity) (*proto.Activity, error) {
	return s.TxPut(s.pbEntityStore.GetDB(), activity)
}

func (s *activityStore) TxPut(tx store.SQLTransactional, activity *proto.Activity) (*proto.Activity, error) {
	if activity.Id == "" {
		activity.Id = wutil.RandomUUID()
	}
	activityData, err := gproto.Marshal(activity)
	if err != nil {
		return nil, err
	}
	updatedPBEntity, err := s.pbEntityStore.TxPut(tx, &store.PBEntity{activity.Id, activityData, time.Now()})
	if err != nil {
		return nil, err
	}
	activity.Id = updatedPBEntity.ID
	return activity, nil
}

func (s *activityStore) TxDelete(tx store.SQLTransactional, id string) error {
	return s.pbEntityStore.TxDelete(tx, id)
}

func (s *activityStore) GetAll() ([]*proto.Activity, error) {
	return s.TxGetAll(s.pbEntityStore.GetDB())
}

func (s *activityStore) TxGetAll(tx store.SQLTransactional) ([]*proto.Activity, error) {
	entities, err := s.pbEntityStore.TxGetAll(tx)
	if err != nil {
		return nil, err
	}
	activities := make([]*proto.Activity, len(entities), len(entities))
	for i, entity := range entities {
		activity := &proto.Activity{}
		err = gproto.Unmarshal(entity.Payload, activity)
		if err != nil {
			return nil, err
		}
		activities[i] = activity
	}
	return activities, nil
}
