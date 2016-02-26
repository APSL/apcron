package main

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/robfig/cron"
	"log"
	"os/exec"
)

type Runner struct {
	cron *cron.Cron
}

func NewRunner() *Runner {
	r := &Runner{
		cron: cron.New(),
	}
	return r
}

func (r *Runner) Add(spec string, cmd string) error {
	err := r.cron.AddFunc(spec, r.cmdFunc(cmd))

	if configuration.Verbose {
		log.Printf("Add cron job spec:%v cmd:%v err:%v", spec, cmd, err)
	}

	return err
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
			if configuration.SentryDSN != "" {
				message := fmt.Sprintln("Error executing command", cmd)
				client, err := raven.NewWithTags(configuration.SentryDSN, map[string]string{"program": "go-cron", "error": err.Error(), "command": cmd})
				if err == nil {
					packet := &raven.Packet{Message: message, Extra: map[string]interface{}{"command": cmd, "output": string(out)}}
					client.Capture(packet, nil)
				}
			}
		} else if configuration.Verbose {
			log.Printf("cmd:%v out:%v err:%v", cmd, string(out), err)
		}
	}
	return cmdFunc
}
