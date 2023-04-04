package grpc

import (
	"fmt"

	"github.com/dlshle/gommon/errors"
	"google.golang.org/grpc"
)

func mian() {
	grpc.NewServer()
	err := errors.Error("no")
	fmt.Println(err)
}
