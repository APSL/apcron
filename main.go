package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	CrontabPath string `envconfig:"cron_file"`
	SentryDSN   string `envconfig:"sentry_dsn"`
	Verbose     bool   `envconfig:"cron_verbose"`
	CmdPrefix   string `envconfig:"cron_cmd_prefix"`
}

var (
	cfg Configuration
)

func init() {
	cfg.CrontabPath = "crontab"
	envconfig.Process("", &cfg)
	flag.StringVar(&cfg.CrontabPath, "file", cfg.CrontabPath, "Crontab file path, env: CRON_FILE")
	flag.BoolVar(&cfg.Verbose, "v", cfg.Verbose, "Show/log messages, env: CRON_VERBOSE")
	flag.StringVar(&cfg.SentryDSN, "sentry-dsn", cfg.SentryDSN, "Sentry DSN, env: SENTRY_DSN")
	flag.StringVar(&cfg.CmdPrefix, "cmd-prefix", cfg.CmdPrefix, "Preffix to append to commands (ex: python manage.py). env: CRON_CMD_PREFIX")
}

func main() {
	flag.Parse()

	file, err := os.Open(cfg.CrontabPath)
	if err != nil {
		log.Fatalf("crontab path:%v err:%v", cfg.CrontabPath, err)
	}

	// Parse crontab or yaml
	parse := ParseCron
	ext := path.Ext(cfg.CrontabPath)
	if ext == ".yaml" || ext == ".yml" {
		parse = ParseYaml
	}

	jobs, err := parse(file)
	if err != nil {
		log.Fatalf("Error parsing cron file: %v", err)
	}

	file.Close()

	runner := NewRunner(cfg.CmdPrefix)
	runner.AddJobs(jobs)

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
