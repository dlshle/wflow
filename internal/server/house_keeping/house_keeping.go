package housekeeping

import (
	"context"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gommon/utils"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron"
)

type houseKeeping struct {
	ctx          context.Context
	logger       logging.Logger
	executor     async.Executor
	schedule     cron.Schedule
	cronExpr     string
	db           *sqlx.DB
	keepInterval time.Duration
}

func StartHouseKeeping(ctx context.Context, executor async.Executor, cronExpr string, db *sqlx.DB, keepInterval time.Duration) error {
	hk := &houseKeeping{
		ctx:          ctx,
		logger:       logging.GlobalLogger.WithPrefix("[HouseKeeping]"),
		executor:     executor,
		cronExpr:     cronExpr,
		db:           db,
		keepInterval: keepInterval,
	}
	return hk.init()
}

func (hk *houseKeeping) init() error {
	schedule, err := cron.Parse(hk.cronExpr)
	if err != nil {
		return err
	}
	hk.schedule = schedule
	hk.executor.Execute(hk.runHouseKeeping)
	return nil
}

func (hk *houseKeeping) runHouseKeeping() {
	nextTime := hk.schedule.Next(time.Now())
	hk.logger.Infof(hk.ctx, "house keeping job will run at %s", nextTime.String())
	hk.executor.Execute(func() {
		remainingTime := nextTime.Sub(time.Now())
		time.Sleep(remainingTime)
		hk.execute()
		hk.executor.Execute(hk.runHouseKeeping)
	})
}

func (hk *houseKeeping) execute() {
	hk.logger.Infof(hk.ctx, "start to clean jobs")
	err := utils.ProcessWithErrors(func() error {
		hk.logger.Infof(hk.ctx, "cleaning jobs from %v ago", hk.keepInterval.String())
		_, err := hk.db.Exec("DELETE FROM jobs WHERE created_at < now() - interval '" + hk.keepInterval.String() + "'")
		return err
	}, func() error {
		hk.logger.Infof(hk.ctx, "cleaning job logs from %v ago", hk.keepInterval.String())
		_, err := hk.db.Exec("DELETE FROM job_logs WHERE created_at < now() - interval '" + hk.keepInterval.String() + "'")
		return err
	}, func() error {
		hk.logger.Infof(hk.ctx, "cleaning workers from %v ago", hk.keepInterval.String())
		_, err := hk.db.Exec("DELETE FROM workers WHERE last_heartbeat < now() - interval '" + hk.keepInterval.String() + "'")
		return err
	}, func() error {
		hk.logger.Infof(hk.ctx, "cleaning activities from %v ago", hk.keepInterval.String())
		_, err := hk.db.Exec("DELETE FROM activities WHERE created_at < now() - interval '" + hk.keepInterval.String() + "'")
		return err
	}, func() error {
		hk.logger.Infof(hk.ctx, "cleaning relation_mappings from %v ago", hk.keepInterval.String())
		_, err := hk.db.Exec("DELETE FROM relation_mappings WHERE created_at < now() - interval '" + hk.keepInterval.String() + "'")
		return err
	})
	if err != nil {
		hk.logger.Errorf(hk.ctx, "failed to clean jobs due to %v, will retry in 30 minutes", err)
		time.Sleep(30 * time.Minute)
		hk.execute()
		return
	}
}
