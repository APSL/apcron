package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/apsl/apcron/manager"
	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron"
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

	cmdSpecs, err := parse(file)
	if err != nil {
		log.Fatalf("Error parsing cron file: %v", err)
	}

	file.Close()

	mgr := manager.New()
	cron := cron.New()
	for _, cs := range cmdSpecs {
		job, err := mgr.CreateJob(cs.Cmd)
		if err != nil {
			log.Printf("Error creating job (%s) in manager: %v\n", cs.Cmd, err)
			return
		}
		cron.AddJob(cs.Spec, job)
		log.Printf("Scheduled job: id=%d. specs=%s, cmd=%s", job.GetID(), cs.Spec, cs.Cmd)
	}
	mgr.Start()
	cron.Start()

	// fmt.Println("Jobs Added:")
	// table := tablewriter.NewWriter(os.Stdout)
	// table.SetHeader([]string{"ID", "SPEC", "CMD"})
	// table.SetAutoWrapText(false)
	// for _, j := range c.Jobs {
	// 	data := []string{strconv.Itoa(j.ID), j.Spec, j.Cmd}
	// 	table.Append(data)
	// }
	// table.Render()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	s := <-c
	log.Printf("Got signal: %s. Exiting apcron.\n", s)
	cron.Stop()
	//mgr.Stop()
}
