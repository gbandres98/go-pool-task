package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/gosuri/uilive"
	"log"
	"net/http"
	"os"
	"redsift-test/pool"
	"strings"
	"sync/atomic"
	"time"
)

type connectionResult struct {
	responseTime int64
	dataSize     int
}

func main() {
	workers, queueSize, input := parseArgs()

	workerPool := pool.InitPool(workers, context.Background())

	queueContext, cancelQueue := context.WithCancel(context.Background())
	workQueue := workerPool.NewQueue(queueSize, queueContext)

	var responseTimeSum atomic.Int64
	var dataSizeSum atomic.Int64

	uiWriter := uilive.New()
	uiWriter.Start()

	startTime := time.Now()

	go refreshStats(uiWriter, workQueue, workerPool, &responseTimeSum, &dataSizeSum)
	go createJobs(input, workQueue, cancelQueue)
	go handleResults(workQueue.ResultChannel(), workQueue.ErrorChannel(), &responseTimeSum, &dataSizeSum)

	<-queueContext.Done()

	printStats(uiWriter, workQueue, workerPool, &responseTimeSum, &dataSizeSum)
	uiWriter.Stop()
	log.Printf("Finished work in %s\n", time.Since(startTime).String())
}

func createJobs(fileName string, workQueue pool.Queue, cancelQueue context.CancelFunc) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = true
	transport.ResponseHeaderTimeout = 10 * time.Second

	client := http.DefaultClient
	client.Transport = transport
	client.Timeout = 10 * time.Second

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	totalJobs := 0

	for scanner.Scan() {
		line := scanner.Text()
		s := strings.Split(line, ",")

		if len(s) < 2 {
			continue
		}

		domain := s[1]

		workQueue.AddJob(func() (any, error) {
			t := time.Now()

			res, err := client.Get("http://" + domain)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()

			elapsed := time.Since(t).Milliseconds()

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(res.Body)
			if err != nil {
				return nil, err
			}

			size := buf.Len()

			return connectionResult{
				responseTime: elapsed,
				dataSize:     size,
			}, nil
		})

		totalJobs++
	}

	checkFinished(workQueue, totalJobs, cancelQueue)
}

func handleResults(resultChan <-chan any, errorChan <-chan error, responseTimeSum *atomic.Int64, dataSizeSum *atomic.Int64) {
	for {
		select {
		case r := <-resultChan:
			res, ok := r.(connectionResult)
			if !ok {
				continue
			}

			responseTimeSum.Add(res.responseTime)
			dataSizeSum.Add(int64(res.dataSize))
		case <-errorChan:
			continue
		}
	}
}

func checkFinished(workQueue pool.Queue, totalJobs int, cancelQueue context.CancelFunc) {
	for {
		if workQueue.FinishedJobCount() >= int32(totalJobs) {
			cancelQueue()
			return
		}
	}
}

func printStats(uiWriter *uilive.Writer, workQueue pool.Queue, workerPool pool.Pool, responseTimeSum *atomic.Int64, dataSizeSum *atomic.Int64) {
	var responseTimeAvg int64 = 0
	if workQueue.SuccessCount() > 0 {
		responseTimeAvg = responseTimeSum.Load() / int64(workQueue.SuccessCount())
	}

	var dataSizeAvg int64 = 0
	if workQueue.SuccessCount() > 0 {
		dataSizeAvg = dataSizeSum.Load() / int64(workQueue.SuccessCount())
	}

	fmt.Fprintf(uiWriter,
		"Workers: %d\n"+
			"Active workers: %d\n"+
			"=================\n"+
			"Created jobs: %d\n"+
			"Queued jobs: %d\n"+
			"Succeded jobs: %d\n"+
			"Errored jobs: %d\n"+
			"=================\n"+
			"Response time average: %dms\n"+
			"Data size average: %d bytes\n",
		workerPool.WorkerCount(),
		workerPool.ActiveWorkerCount(),
		workQueue.JobCount(),
		workQueue.QueuedJobCount(),
		workQueue.SuccessCount(),
		workQueue.ErrorCount(),
		responseTimeAvg,
		dataSizeAvg,
	)
}

func refreshStats(uiWriter *uilive.Writer, workQueue pool.Queue, workerPool pool.Pool, responseTimeSum *atomic.Int64, dataSizeSum *atomic.Int64) {
	ticker := time.NewTicker(500 * time.Millisecond)

	for range ticker.C {
		printStats(uiWriter, workQueue, workerPool, responseTimeSum, dataSizeSum)
	}
}

func parseArgs() (int, int, string) {
	workers := flag.Int("workers", 200, "number of workers to spawn")
	queueSize := flag.Int("queue-size", 20000, "size of job queue")
	input := flag.String("input", "input.csv", "input file name")

	flag.Parse()

	return *workers, *queueSize, *input
}
