package scaler

import (
	"fmt"
	"os"
	"strconv"
)

type ScalerConfig struct {
	TasksPerContainer   int
	MaxWorkerContainers int
	MinWorkerContainers int
	ContainerNamePrefix string
	ContainerImage      string
	ContainerTag        string
}

// TODO : fix this miss
// probably best to search for library with struct marshaling
func NewScalerConfigFromEnv() (*ScalerConfig, error) {
	cfg := ScalerConfig{}
	val, ok := os.LookupEnv("TASKS_PER_CONTAINER")
	if !ok {
		return nil, fmt.Errorf("TASKS_PER_CONTAINER is not set")
	}
	tasksPerContainer, err := strconv.Atoi(val)
	if err != nil {
		return nil, fmt.Errorf("TASKS_PER_CONTAINER is not a number")
	}
	cfg.TasksPerContainer = tasksPerContainer
	val, ok = os.LookupEnv("MAX_WORKER_CONTAINERS")
	if !ok {
		return nil, fmt.Errorf("MAX_WORKER_CONTAINERS is not set")
	}
	maxWorkerContainers, err := strconv.Atoi(val)
	if err != nil {
		return nil, fmt.Errorf("MAX_WORKER_CONTAINERS is not a number")
	}
	cfg.MaxWorkerContainers = maxWorkerContainers
	val, ok = os.LookupEnv("MIN_WORKER_CONTAINERS")
	if !ok {
		return nil, fmt.Errorf("MIN_WORKER_CONTAINERS is not set")
	}
	minWorkerContainers, err := strconv.Atoi(val)
	if err != nil {
		return nil, fmt.Errorf("MIN_WORKER_CONTAINERS is not a number")
	}
	cfg.MinWorkerContainers = minWorkerContainers
	cfg.ContainerNamePrefix = "go-dev-con_"
	cfg.ContainerImage = "teamparallax/conversion-service"
	cfg.ContainerTag = "latest"
	return &cfg, nil
}
