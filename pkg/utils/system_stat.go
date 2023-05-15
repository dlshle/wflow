package utils

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/dlshle/wflow/proto"
)

type memory struct {
	MemTotal     int
	MemFree      int
	MemAvailable int
}

func readMemoryStats() memory {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memory{
		MemTotal:     int(memStats.TotalAlloc),
		MemFree:      int(memStats.HeapIdle),
		MemAvailable: int(memStats.HeapSys),
	}
}

func parseLine(raw string) (key string, value int) {
	fmt.Println(raw)
	text := strings.ReplaceAll(raw[:len(raw)-2], " ", "")
	keyValue := strings.Split(text, ":")
	return keyValue[0], toInt(keyValue[1])
}

func toInt(raw string) int {
	if raw == "" {
		return 0
	}
	res, err := strconv.Atoi(raw)
	if err != nil {
		panic(err)
	}
	return res
}

func GetSystemStat() *proto.SystemStat {
	cpuCount := runtime.NumCPU()
	memStat := readMemoryStats()
	totalMem := memStat.MemTotal
	availableMem := memStat.MemAvailable
	return &proto.SystemStat{
		CpuCount:               int32(cpuCount),
		TotalMemoryInBytes:     int32(totalMem),
		AvailableMemoryInBytes: int32(availableMem),
	}
}
