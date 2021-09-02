package scaler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/StefanWin/dcv/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Worker struct {
	ID       string
	Name     string
	Info     types.ContainerJSON
	Requests []*api.ConversionRequest
}

func StartWorkerAsync(wg *sync.WaitGroup, ch chan<- *Worker, docker *client.Client, idx int, name string, image string) {
	defer wg.Done()
	worker, err := StartWorker(docker, idx, name, image)
	if err != nil {
		log.Printf("[worker]:: %v\n", err)
		ch <- nil
	}
	ch <- worker
}

func StartWorker(docker *client.Client, idx int, name string, image string) (*Worker, error) {
	// TODO : fix this
	data, _ := os.ReadFile(".service-env")
	envLines := strings.Split(string(data), "\n")
	ctx := context.Background()
	containerName := fmt.Sprintf("%s%d", name, idx)
	resp, err := docker.ContainerCreate(ctx, &container.Config{
		Image: image,
		Env:   envLines,
	}, nil, nil, nil, containerName)
	if err != nil {
		return nil, err
	}
	if err := docker.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	info, err := docker.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("[scaler]:: started container '%s' with %s\n", containerName, image)
	return &Worker{ID: resp.ID, Name: containerName, Info: info}, nil
}

func (worker *Worker) RemoveAsync(wg *sync.WaitGroup, docker *client.Client) error {
	defer log.Printf("[scaler]:: removed %s\n", worker.Name)
	defer wg.Done()
	return worker.Remove(docker)
}

func (worker *Worker) Remove(docker *client.Client) error {
	return docker.ContainerRemove(context.Background(), worker.ID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (worker *Worker) Update(docker *client.Client) error {
	var err error
	worker.Info, err = docker.ContainerInspect(context.Background(), worker.ID)
	return err
}

func (worker *Worker) IsHealthy() bool {
	if worker.Info.State != nil {
		if worker.Info.State.Health != nil {
			status := worker.Info.State.Health.Status
			if status == "healthy" || status == "starting" {
				return true
			}
		}
	}
	// TODO : investigate why the state can be nil
	return true
}

func (worker *Worker) DispatchRequest(request api.ConversionRequest) error {
	// TODO : multipart post request
	request.WorkerConversionID = "foobar"
	worker.Requests = append(worker.Requests, &request)
	log.Printf("[%s]:: received %s\n", worker.Name, request.ExternalConversionID)
	return nil
}

func (worker *Worker) UpdateStatus() {
	ip := worker.Info.NetworkSettings.IPAddress
	baseUrl := fmt.Sprintf("http://%s:%d", ip, 3001)
	for _, request := range worker.Requests {
		resp, err := http.Get(fmt.Sprintf("%s/conversion/%s?v2=true", baseUrl, request.WorkerConversionID))
		if err != nil {
			request.ConversionStatus = err.Error()
			continue
		}
		csr := api.ConversionStatusResponse{}

		if err := json.NewDecoder(resp.Body).Decode(&csr); err != nil {
			request.ConversionStatus = err.Error()
			continue
		}
		request.ConversionStatus = csr.Status
		fmt.Println(request.ConversionStatus)
		resp.Body.Close()
	}
	for _, req := range worker.Requests {
		log.Printf("[%s]:: %s => %s\n", worker.Name, req.ExternalConversionID, req.ConversionStatus)
	}
}

func (worker *Worker) GetRequestCount() int {
	return len(worker.Requests)
}
