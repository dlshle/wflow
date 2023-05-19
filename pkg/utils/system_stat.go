package utils

import (
	"runtime"

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
