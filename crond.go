package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/robfig/cron"
)

//Crond represents a cron runner
type Crond struct {
	cron      *cron.Cron
	CmdPrefix string
	Manager   *Manager
	Jobs      []Job
}

//NewCrond creates a new Runner
func NewCrond(cmdPrefix string, m *Manager) *Crond {
	r := &Crond{
		cron:      cron.New(),
		CmdPrefix: cmdPrefix,
		Manager:   m,
		Jobs:      []Job{},
	}
	return r
}

//Add adds a job to the crond
func (c *Crond) Add(job Job) error {
	if c.CmdPrefix != "" {
		job.Cmd = fmt.Sprintf("%s %s", c.CmdPrefix, job.Cmd)
	}
	err := c.cron.AddFunc(job.Spec, c.jobFunc(job))
	if err != nil {
		log.Printf("Error adding cron job: %s: %s\n", job.Spec, err)
	}

	log.Printf("Crond: Added cron job with ID: [%d] spec: [%s] cmd: [%s]\n", job.ID, job.Spec, job.Cmd)
	c.Jobs = append(c.Jobs, job)

	return err
}

func (c *Crond) jobFunc(j Job) func() {
	jobFunc := func() {
		c.Manager.RunAndWait(&j)
	}
	return jobFunc
}

func (c *Crond) AddJobs(jobs []Job) error {
	for i, j := range jobs {
		j.ID = i
		if err := c.Add(j); err != nil {
			return err
		}
	}
	c.PrintTable()
	return nil
}

func (c *Crond) PrintTable() {
	fmt.Println("Jobs Added:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "SPEC", "CMD"})
	table.SetAutoWrapText(false)
	for _, j := range c.Jobs {
		data := []string{strconv.Itoa(j.ID), j.Spec, j.Cmd}
		table.Append(data)
	}
	table.Render()
}

func (c *Crond) Len() int {
	return len(c.cron.Entries())
}

func (c *Crond) Start() {
	log.Println("Start crond")
	c.cron.Start()
}

func (c *Crond) Stop() {
	c.cron.Stop()
	log.Println("Stop crond")
}
