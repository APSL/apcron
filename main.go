package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/apsl/apcron/manager"
	"github.com/apsl/apcron/parser"
	"github.com/kelseyhightower/envconfig"
	"github.com/olekukonko/tablewriter"
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
	cfg.CrontabPath = "crontab.yaml"
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

	jobDefs, err := parser.ParseYaml(file)
	if err != nil {
		log.Fatalf("Error parsing cron file: %v", err)
	}

	file.Close()

	mgr := manager.New()
	cron := cron.New()
	for _, jd := range jobDefs {
		if cfg.CmdPrefix != "" {
			jd.Cmd = cfg.CmdPrefix + " " + jd.Cmd
		}
		job, err := mgr.CreateJob(jd.Cmd, jd.Spec, jd.Shell)
		if err != nil {
			log.Printf("Error creating job (%s) in manager: %v\n", jd.Cmd, err)
			return
		}
		cron.AddJob(jd.Spec, job)
		log.Printf("Scheduled job: id=%d. specs=%s, cmd=%s, shell=%s", job.GetID(), jd.Spec, jd.Cmd, jd.Shell)
	}
	mgr.Start()
	cron.Start()

	printJobsTable(cron.Entries())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	s := <-c
	log.Printf("Got signal: %s. Exiting apcron.\n", s)
	cron.Stop()
	//mgr.Stop()
}

func printJobsTable(entries []*cron.Entry) {
	fmt.Println("Jobs Added:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "SPEC", "NEXT", "CMD"})
	table.SetAutoWrapText(false)
	for _, e := range entries {
		job := e.Job.(*manager.Job)
		dif := e.Next.Sub(time.Now())
		next := fmt.Sprintf("%s (%s)", dif, e.Next.String())
		data := []string{strconv.Itoa(job.GetID()), job.Spec, next, job.GetCmd()}
		table.Append(data)
	}
	table.Render()
}
