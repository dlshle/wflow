package activity

import "github.com/dlshle/wflow/proto"

type WorkerActivity interface {
	Activity() *proto.Activity
	Handler() ActivityHandler
}

type workerActivity struct {
	activity *proto.Activity
	handler  ActivityHandler
}

func NewWorkerActivity(activity *proto.Activity, handler ActivityHandler) WorkerActivity {
	return &workerActivity{
		activity: activity,
		handler:  handler,
	}
}

func (a *workerActivity) Activity() *proto.Activity {
	return a.activity
}

func (a *workerActivity) Handler() ActivityHandler {
	return a.handler
}
