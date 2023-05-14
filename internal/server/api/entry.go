package api

import (
	"fmt"

	"github.com/dlshle/aghs/server"
	"github.com/dlshle/gommon/utils"
	"github.com/dlshle/wflow/internal/protocol"
	"github.com/dlshle/wflow/internal/server/activity"
	"github.com/dlshle/wflow/internal/server/config"
	"github.com/dlshle/wflow/internal/server/job"
	relationmapping "github.com/dlshle/wflow/internal/server/relation_mapping"
	"github.com/dlshle/wflow/internal/server/service"
	"github.com/dlshle/wflow/internal/server/worker"
	"github.com/dlshle/wflow/pkg/store"
	"github.com/jmoiron/sqlx"
)

func Entry(serverID, configPath string) (err error) {
	var (
		cfg config.ServerConfig
		db  *sqlx.DB
		// jobStore               job.Store
		jobHandler job.Handler
		// workerStore            worker.Store
		workerManager          worker.Manager
		activityStore          activity.Store
		activityHandler        activity.Handler
		relationMappingHandler relationmapping.Handler
		adminService           service.AdminService
		httpServer             server.Server
		tcpServer              protocol.TCPServer
	)
	return utils.ProcessWithErrors(func() error {
		cfg, err = config.LoadConfig(configPath)
		return err
	}, func() error {
		db, err = setupDatabase(cfg)
		return err
	}, func() error {
		activityStore, activityHandler, err = setupActivityDependencies(cfg)
		return err
	}, func() error {
		relationMappingHandler, err = setupRelationMappingDependencies(db, activityStore)
		return err
	}, func() error {
		_, jobHandler, err = setupJobsDependencies(cfg, relationMappingHandler)
		return err
	}, func() error {
		_, workerManager, err = setupWorkerDependencies(cfg, relationMappingHandler, jobHandler, activityHandler)
		return err
	}, func() error {
		adminService, err = setupAdminDependencies(activityHandler, relationMappingHandler, jobHandler, workerManager)
		return err
	}, func() error {
		httpServer, err = NewHTTPServer(cfg.HTTPPort, adminService)
		return err
	}, func() error {
		tcpServer = NewTCPServer(serverID, "0.0.0.0", cfg.TCPPort, workerManager, jobHandler)
		return nil
	}, func() error {
		return runServers(httpServer, tcpServer)
	})
}

func runServers(httpServer server.Server, tcpServer protocol.TCPServer) (err error) {
	httpErrChan := make(chan error, 1)
	tcpErrChan := make(chan error, 1)
	go func() {
		httpErrChan <- httpServer.Start()
	}()
	go func() {
		tcpErrChan <- tcpServer.Start()
	}()
	defer close(httpErrChan)
	defer close(tcpErrChan)
	select {
	case err = <-httpErrChan:
		return err
	case err = <-tcpErrChan:
		return err
	}
}

func setupDatabase(cfg config.ServerConfig) (db *sqlx.DB, err error) {
	err = utils.ProcessWithErrors(func() error {
		db, err = sqlx.Open("postgres", getDatabaseConnectionString(cfg, "postgres"))
		return err
	}, func() error {
		return store.ExecMigration(db, "migrations")
	})
	return
}

func getDatabaseConnectionString(cfg config.ServerConfig, dbName string) string {
	fullDBUri := fmt.Sprintf("%s:%d", cfg.Database.Host, cfg.Database.Port)
	return fmt.Sprintf("postgresql://%s:%s@%s/defaultdb?database=%s", cfg.Database.Username, cfg.Database.Password,
		fullDBUri, dbName)
}

func setupWorkerDependencies(cfg config.ServerConfig, relationMappingHandler relationmapping.Handler, jobHandler job.Handler, activityHandler activity.Handler) (workerStore worker.Store, manager worker.Manager, err error) {
	workerStore, err = worker.NewStore(getDatabaseConnectionString(cfg, "workers"))
	manager = worker.NewManager(workerStore, relationMappingHandler, jobHandler, activityHandler)
	return
}

func setupActivityDependencies(cfg config.ServerConfig) (activityStore activity.Store, activityHandler activity.Handler, err error) {
	activityStore, err = activity.NewStore(getDatabaseConnectionString(cfg, "activities"))
	if err != nil {
		return
	}
	activityHandler = activity.NewHandler(activityStore)
	return
}

func setupJobsDependencies(cfg config.ServerConfig, relationMappingHandler relationmapping.Handler) (store job.Store, handler job.Handler, err error) {
	store, err = job.NewStore(getDatabaseConnectionString(cfg, "jobs"))
	if err != nil {
		return
	}
	handler = job.NewHandler(store, relationMappingHandler)
	return
}

func setupRelationMappingDependencies(db *sqlx.DB, activityStore activity.Store) (handler relationmapping.Handler, err error) {
	store := relationmapping.NewRelationMappingStore(db)
	handler = relationmapping.NewHandler(store, activityStore)
	return
}

func setupAdminDependencies(activityHandler activity.Handler, relationMappingHandler relationmapping.Handler, jobHandler job.Handler, workerManager worker.Manager) (adminService service.AdminService, err error) {
	adminService = service.NewAdminService(jobHandler, activityHandler, relationMappingHandler, workerManager)
	return
}
