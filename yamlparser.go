package main

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/apsl/apcron/job"
	yaml "gopkg.in/yaml.v2"
)

type CronJob struct {
	Second   string `yaml:"second"`
	Minute   string `yaml:"minute"`
	Hour     string `yaml:"hour"`
	MonthDay string `yaml:"monthday"`
	Month    string `yaml:"month"`
	WeekDay  string `yaml:"weekday"`
	Command  string `yaml:"command"`
}
type CronJobs map[string]CronJob

func val(s string, d string) string {
	if s == "" {
		return d
	}
	return s
}

//Spec returns crontab spec (without command)
func (c *CronJob) Spec() string {
	return fmt.Sprintf("%s %s %s %s %s %s",
		val(c.Second, "0"), val(c.Minute, "*"), val(c.Hour, "*"),
		val(c.MonthDay, "*"), val(c.Month, "*"), val(c.WeekDay, "*"))
}

//ParseYaml parses a yaml and returns a Job slice
func ParseYaml(r io.Reader) (jobs []job.Job, err error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	var cronJobs CronJobs
	err = yaml.Unmarshal(data, &cronJobs)
	if err != nil {
		return
	}
	for _, cronJob := range cronJobs {
		job := job.Job{
			Spec: cronJob.Spec(),
			Cmd:  cronJob.Command,
		}
		jobs = append(jobs, job)
	}
	return
}
