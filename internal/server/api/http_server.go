package api

import (
	"context"
	"fmt"

	"github.com/dlshle/aghs/server"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/server/service"
	"github.com/dlshle/wflow/proto"
	gproto "google.golang.org/protobuf/proto"
)

type httpServer struct {
	server.Server
}

func NewHTTPServer(port int, adminService service.AdminService) (*httpServer, error) {
	adminHTTPServerService, err := server.NewServiceBuilder().
		Id("adminService").
		WithRouteHandlers(server.PathHandlerBuilder("/activeWorkers").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "ListAllActiveWorkerConnections")
				workerProtos, err := adminService.ListAllActiveWorkers(ctx)
				workersMap := make(map[string]*proto.Worker)
				for _, worker := range workerProtos {
					workersMap[worker.Id] = worker
				}
				workersResponse := &proto.AdminWorkersResponse{Workers: workersMap}
				workersResponseData, err := gproto.Marshal(workersResponse)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, string(workersResponseData))
			}).MustBuild().HandleRequest)).
		WithRouteHandlers(server.PathHandlerBuilder("/worker/:id").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				workerID := c.PathParam("id")
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "GetWorkerByID")
				ctx = logging.WrapCtx(ctx, "workerID", workerID)
				worker, err := adminService.GetWorker(ctx, workerID)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				workerData, err := gproto.Marshal(worker)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, string(workerData))
			}).MustBuild().HandleRequest).
			Delete(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				workerID := c.PathParam("id")
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "DisconnectWorkerByID")
				ctx = logging.WrapCtx(ctx, "workerID", workerID)
				err := adminService.DisconnectWorker(ctx, workerID)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, nil)
			}).MustBuild().HandleRequest).
			Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/jobs/:id").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				jobID := c.PathParam("id")
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "GetJobByID")
				ctx = logging.WrapCtx(ctx, "jobID", jobID)
				job, err := adminService.GetJobByID(ctx, jobID)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				jobData, err := gproto.Marshal(job)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, string(jobData))
			}).MustBuild().HandleRequest).
			Delete(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				jobID := c.PathParam("id")
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "CancelJobByID")
				ctx = logging.WrapCtx(ctx, "jobID", jobID)
				err := adminService.CancelJob(ctx, jobID)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, nil)
			}).MustBuild().HandleRequest).
			Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/activities/:id").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				activityID := c.PathParam("id")
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "GetActivityByID")
				ctx = logging.WrapCtx(ctx, "activityID", activityID)
				activity, err := adminService.GetActivity(ctx, activityID)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				activityData, err := gproto.Marshal(activity)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, string(activityData))
			}).MustBuild().HandleRequest).Build()).
		Build()
	if err != nil {
		return nil, err
	}
	actualHTTPServer, err := server.NewBuilder().WithService(adminHTTPServerService).Address(fmt.Sprintf("0.0.0.0:%d", port)).Build()
	if err != nil {
		return nil, err
	}
	return &httpServer{actualHTTPServer}, err
}
