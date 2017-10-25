package main

// Job is the cron spec to be parsed by a cron parser
// Spec is the crontab format string
// Cmd is the command to be executed
type Job struct {
	Spec string
	Cmd  string
	ID   int
}

// func (j *Job) Run() {
// 	if j.Manager != nil {
// 		j.Manager.RunAndWait(j)
// 	}
// 	log.Printf("Job: finished")
// }
