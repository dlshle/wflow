package utils

import (
	"bufio"
	"fmt"
	"os"
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
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	bufio.NewScanner(file)
	scanner := bufio.NewScanner(file)
	res := memory{}
	for scanner.Scan() {
		key, value := parseLine(scanner.Text())
		switch key {
		case "MemTotal":
			res.MemTotal = value
		case "MemFree":
			res.MemFree = value
		case "MemAvailable":
			res.MemAvailable = value
		}
	}
	return res
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
