package api

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/dlshle/aghs/contrib/middlewares"
	"github.com/dlshle/aghs/server"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/wflow/internal/server/service"
	"github.com/dlshle/wflow/proto"
	"github.com/golang/protobuf/jsonpb"
)

type httpServer struct {
	server.Server
}

type DispatchJobRequest struct {
	WorkerID   string `json:"workerId"`
	ActivityID string `json:"activityId"`
	Payload    string `json:"payload"`
}

func NewHTTPServer(port int, adminService service.AdminService) (*httpServer, error) {
	marshaller := jsonpb.Marshaler{OrigName: true}
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
				jsonified, err := marshaller.MarshalToString(workersResponse)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, jsonified)
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
				workerData, err := marshaller.MarshalToString(worker)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, workerData)
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
				jobData, err := marshaller.MarshalToString(job)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, jobData)
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
		WithRouteHandlers(server.PathHandlerBuilder("/jobs").
			Post(server.NewCHandlerBuilder[DispatchJobRequest]().UseDefaultUnmarshaller().OnRequest(func(c server.CHandle[DispatchJobRequest]) server.Response {
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "DispatchJob")
				ctx = logging.WrapCtx(ctx, "workerID", c.Data().WorkerID)
				ctx = logging.WrapCtx(ctx, "activityID", c.Data().ActivityID)
				data := c.Data()
				decoded, err := base64.StdEncoding.DecodeString(data.Payload)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				jobReport, err := adminService.DispatchJob(ctx, data.ActivityID, data.WorkerID, decoded)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				jobsResponseData, err := marshaller.MarshalToString(jobReport)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, jobsResponseData)
			}).MustBuild().HandleRequest).
			Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/jobs/:id/logs").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				jobID := c.PathParam("id")
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "jobID", jobID)
				jobLogs, err := adminService.GetLogsByJobID(ctx, jobID)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				wrappedLogs := &proto.WrappedLogs{Logs: jobLogs}
				logs, err := marshaller.MarshalToString(wrappedLogs)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, logs)
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
				activityData, err := marshaller.MarshalToString(activity)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, activityData)
			}).MustBuild().HandleRequest).Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/activities/:id/workers").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				activityID := c.PathParam("id")
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "GetActivityByID")
				ctx = logging.WrapCtx(ctx, "activityID", activityID)
				workers, err := adminService.GetWorkersByActivityID(ctx, activityID)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				activityData, err := marshaller.MarshalToString(&proto.AdminWorkersResponse{Workers: workers})
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, activityData)
			}).MustBuild().HandleRequest).Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/activeActivities").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "GetActiveActivities")
				activeActivities, err := adminService.ListAllActiveActivities(ctx)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				activeActivitiesMap := make(map[string]*proto.Activity)
				for _, activity := range activeActivities {
					activeActivitiesMap[activity.Id] = activity
				}
				activityData, err := marshaller.MarshalToString(&proto.AdminActivitiesResponse{Activities: activeActivitiesMap})
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, activityData)
			}).MustBuild().HandleRequest).Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/activities").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				ctx := context.Background()
				ctx = logging.WrapCtx(ctx, "adminService", "GetActivitiesWithJobIDs")
				activitiesWithJobIDs, err := adminService.ListAllActivitiesWithJobIDs(ctx)
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				activeActivitiesMap := make(map[string]*proto.ActivityWithJobIDs)
				for _, activity := range activitiesWithJobIDs {
					activeActivitiesMap[activity.Activity.Id] = activity
				}
				activityData, err := marshaller.MarshalToString(&proto.AdminActivitiesWithJobIDsResponse{Activities: activeActivitiesMap})
				if err != nil {
					return server.NewResponse(500, err.Error())
				}
				return server.NewResponse(200, activityData)
			}).MustBuild().HandleRequest).Build()).
		Build()
	if err != nil {
		return nil, err
	}
	actualHTTPServer, err := server.NewBuilder().WithMiddleware(middlewares.CORSAllowWildcardMiddleware).WithService(adminHTTPServerService).Address(fmt.Sprintf("0.0.0.0:%d", port)).Build()
	if err != nil {
		return nil, err
	}
	return &httpServer{actualHTTPServer}, err
}
