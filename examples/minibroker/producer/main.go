package main

import (
	"app/examples/minibroker/producer/cmd"
	"app/pkg/db"
	"app/pkg/mq"
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	nWorkers int
)

func init() {
	numcpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numcpu) // Try to use all available CPUs.
	nWorkers = numcpu
	flag.IntVar(&nWorkers, "workers", nWorkers, "Get numbers of workers")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	rdb := db.GetRedisDbClient(context.Background())

	quitChannel := make(chan os.Signal, 1)
	msgBroker := mq.NewMsgBroker(1024)
	mq.InitMQTTClient(msgBroker)

	// Start Staging Channel -> Redis Workers
	for i := 0; i < nWorkers; i++ {
		go cmd.SendData(ctx, msgBroker.StagingC, rdb)
	}

	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}
