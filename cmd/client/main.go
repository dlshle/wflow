package main

import (
	"time"

	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/internal/client/api"
	"github.com/dlshle/wflow/internal/client/contrib"
	"github.com/dlshle/wflow/pkg/utils"
)

func main() {
	// load shell executor by default
	client, err := api.New(utils.GenerateUUIDOnString("local"), "localhost", 14514, []activity.WorkerActivity{contrib.NewShellActivity()})
	if err != nil {
		panic(err)
	}
	err = client.Start()
	if err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Minute)
}
