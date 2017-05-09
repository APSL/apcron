package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/getsentry/raven-go"
	"github.com/robfig/cron"
)

//Runner represents a cron runner
type Runner struct {
	cron      *cron.Cron
	CmdPrefix string
}

// Job is the cron spec to be parsed by a cron parser
// Spec is the crontab format string
// Cmd is the command to be executed
type Job struct {
	Spec string
	Cmd  string
}

//NewRunner creates a new Runner
func NewRunner(cmdPrefix string) *Runner {
	r := &Runner{
		cron:      cron.New(),
		CmdPrefix: cmdPrefix,
	}
	return r
}

//Add adds a job to the runner
func (r *Runner) Add(job Job) error {
	cmd := job.Cmd
	if r.CmdPrefix != "" {
		cmd = fmt.Sprintf("%s %s", r.CmdPrefix, job.Cmd)
	}
	err := r.cron.AddFunc(job.Spec, r.cmdFunc(cmd))

	if cfg.Verbose {
		log.Printf("Add cron job spec:%v cmd:%v err:%v", job.Spec, cmd, err)
	}

	return err
}

func (r *Runner) AddJobs(jobs []Job) error {
	for _, j := range jobs {
		if err := r.Add(j); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) Len() int {
	return len(r.cron.Entries())
}

func (r *Runner) Start() {
	log.Println("Start runner")
	r.cron.Start()
}

func (r *Runner) Stop() {
	r.cron.Stop()
	log.Println("Stop runner")
}

func (r *Runner) cmdFunc(cmd string) func() {
	cmdFunc := func() {
		out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
		if err != nil {
			log.Printf("Error executing command:%v out:%v err:%v", cmd, string(out), err)
			if cfg.SentryDSN != "" {
				message := fmt.Sprintln("Error executing command", cmd)
				client, err := raven.NewWithTags(cfg.SentryDSN, map[string]string{"program": "go-cron", "error": err.Error(), "command": cmd})
				if err == nil {
					packet := &raven.Packet{Message: message, Extra: map[string]interface{}{"command": cmd, "output": string(out)}}
					client.Capture(packet, nil)
				}
			}
		} else if cfg.Verbose {
			log.Printf("cmd:%v out:%v err:%v", cmd, string(out), err)
		}
	}
	return cmdFunc
}
