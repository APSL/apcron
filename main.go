package main

import (
	"flag"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

type Configuration struct {
	CrontabPath string
	NumCPU      int
	SentryDSN   string
	Verbose     bool
}

var (
	configuration Configuration
)

func init() {
	configuration.CrontabPath = "crontab"
	configuration.NumCPU = runtime.NumCPU()
	envconfig.Process("CRON", &configuration)
	flag.StringVar(&configuration.CrontabPath, "file", configuration.CrontabPath, "Crontab file path, env: CRON_CRONTABPATH")
	flag.IntVar(&configuration.NumCPU, "cpu", configuration.NumCPU, "Maximum number of CPUs, env: CRON_NUMCPU")
	flag.StringVar(&configuration.SentryDSN, "sentry", configuration.SentryDSN, "Sentry DSN to log unsuccesful commands, env: CRON_SENTRYDSN")
	flag.BoolVar(&configuration.Verbose, "v", configuration.Verbose, "Show/log messages (CRON_VERBOSE)")
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(configuration.NumCPU)

	file, err := os.Open(configuration.CrontabPath)
	if err != nil {
		log.Fatalf("crontab path:%v err:%v", configuration.CrontabPath, err)
	}

	parser, err := NewParser(file)
	if err != nil {
		log.Fatalf("Parser read err:%v", err)
	}

	runner, err := parser.Parse()
	if err != nil {
		log.Fatalf("Parser parse err:%v", err)
	}

	file.Close()

	var wg sync.WaitGroup
	shutdown(runner, &wg)

	runner.Start()
	wg.Add(1)

	wg.Wait()
	log.Println("End cron")
}

func shutdown(runner *Runner, wg *sync.WaitGroup) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		log.Println("Got signal: ", s)
		runner.Stop()
		wg.Done()
	}()
}
