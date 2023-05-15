package main

import (
	"github.com/dlshle/wflow/internal/client/activity"
	"github.com/dlshle/wflow/internal/client/api"
	"github.com/dlshle/wflow/internal/client/contrib"
)

func main() {
	// load shell executor by default
	client, err := api.New("local", "localhost", 14514, []activity.WorkerActivity{contrib.NewShellActivity()})
	if err != nil {
		panic(err)
	}
	err = client.Start()
	if err != nil {
		panic(err)
	}
}
