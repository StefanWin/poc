package scaler

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/StefanWin/dcv/api"
	"github.com/docker/docker/client"
)

type Scaler struct {
	// <-chan means read-only
	RequestChannel <-chan api.ConversionRequest
	Config         *ScalerConfig
	DockerClient   *client.Client
	Workers        []*Worker
	isRunning      bool
	containerIndex int
}

func NewScaler(requestChannel <-chan api.ConversionRequest) (*Scaler, error) {
	config, err := NewScalerConfigFromEnv()
	if err != nil {
		return nil, err
	}
	// DOCKER_CERT_PATH
	// DOCKER_HOST
	// DOCKER_TLS_VERIFY \\ for insecure stuff
	// DOCKER_API_VERSION
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Scaler{
		Config:         config,
		RequestChannel: requestChannel,
		DockerClient:   client,
		containerIndex: 0,
		isRunning:      false,
	}, nil
}

func (scaler *Scaler) Initialize() error {
	log.Printf("[scaler]:: initializing %d workers\n", scaler.Config.MinWorkerContainers)
	namePrefix := scaler.Config.ContainerNamePrefix
	image := fmt.Sprintf("%s:%s", scaler.Config.ContainerImage, scaler.Config.ContainerTag)
	ch := make(chan *Worker, scaler.Config.MinWorkerContainers)
	var wg sync.WaitGroup
	for i := 1; i <= scaler.Config.MinWorkerContainers; i++ {
		wg.Add(1)
		// dispatch min amount of worker creations
		go StartWorkerAsync(&wg, ch, scaler.DockerClient, scaler.containerIndex, namePrefix, image)
		scaler.containerIndex++
	}
	wg.Wait()
	for i := 1; i <= scaler.Config.MinWorkerContainers; i++ {
		worker := <-ch
		if worker != nil {
			scaler.Workers = append(scaler.Workers, worker)
		}
	}
	if len(scaler.Workers) != scaler.Config.MinWorkerContainers {
		return fmt.Errorf("need %d containers, got %d", scaler.Config.MinWorkerContainers, len(scaler.Workers))
	}
	log.Printf("[scaler]:: successfully initialized")
	return nil
}

func (scaler *Scaler) Start() error {
	scaler.isRunning = true
	log.Println("[scaler]:: starting main loop")
	for scaler.isRunning {
		// start, remove := scaler.Evaluate()
		log.Println("=== Check =======================")
		for _, worker := range scaler.Workers {
			worker.Update(scaler.DockerClient)
			if !worker.IsHealthy() {
				log.Printf("[scaler]:: %s is unhealthy\n", worker.Name)
				// TODO : remove unhealthy containers
			}
		}
		log.Println("[scaler]:: updated worker container info")
		pendingRequests := len(scaler.RequestChannel)
		log.Printf("[scaler]:: %d requests in queue\n", pendingRequests)
		if pendingRequests > 0 {
			idleWorkers := make([]*Worker, 0)
			for _, worker := range scaler.Workers {
				if worker.GetRequestCount() == 0 {
					idleWorkers = append(idleWorkers, worker)
				}
			}
			idleWorkerCount := len(idleWorkers)
			log.Printf("[scaler]:: %d available idle workers\n", idleWorkerCount)
			requests := make([]api.ConversionRequest, 0)
			for i := 0; i < idleWorkerCount-pendingRequests; i++ {
				request := <-scaler.RequestChannel
				requests = append(requests, request)
			}
			// TODO : dispatch requests to workers
			for i, req := range requests {
				idleWorkers[i].DispatchRequest(req)
			}
			log.Printf("[scaler]:: dispatched %d requests to workers\n", len(requests))
		}

		// TODO : fetch status updates from workers
		for _, worker := range scaler.Workers {
			worker.UpdateStatus()
		}
		log.Println("[scaler]:: retrieved status updates from workers")

		// TODO : fetch files from workers
		log.Println("=================================")
		time.Sleep(time.Second * 10)
	}
	log.Println("[scaler]:: main loop exited")
	return nil
}

func (scaler *Scaler) Shutdown() {
	scaler.isRunning = false
	count := len(scaler.Workers)
	log.Printf("[scaler]:: removing %d workers\n", count)
	var wg sync.WaitGroup
	for _, worker := range scaler.Workers {
		wg.Add(1)
		go worker.RemoveAsync(&wg, scaler.DockerClient)
	}
	wg.Wait()
}

// There are no implicit type casts ...
func (scaler *Scaler) Evaluate() (int, int) {
	runningContainers := len(scaler.Workers)
	pendingRequests := len(scaler.RequestChannel)
	tasksPerContainer := scaler.Config.TasksPerContainer
	maxContainers := scaler.Config.MaxWorkerContainers
	minContainers := scaler.Config.MinWorkerContainers
	if pendingRequests == 0 {
		return 0, 0
	}
	if runningContainers == 0 {
		return int(math.Ceil(float64(pendingRequests) / float64(tasksPerContainer))), 0
	}
	start := 0
	remove := 0
	pendingsTasksPerContainer := int(math.Ceil(float64(pendingRequests) / float64(runningContainers)))
	if pendingsTasksPerContainer > tasksPerContainer {
		remainingTasks := pendingRequests - tasksPerContainer*runningContainers
		start = int(math.Ceil(float64(remainingTasks) / float64(tasksPerContainer)))
		if start+runningContainers > maxContainers {
			start = maxContainers - runningContainers
		}
	} else if pendingsTasksPerContainer < tasksPerContainer {
		requiredContainers := int(math.Max(
			float64(pendingRequests-tasksPerContainer*runningContainers),
			0,
		))
		remove = runningContainers - requiredContainers
		remove = int(math.Max(
			float64(runningContainers-remove),
			float64(runningContainers-requiredContainers-minContainers),
		))
	}
	return start, remove
}
