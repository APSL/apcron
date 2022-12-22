package manager

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/apsl/apcron/process"
	"github.com/getsentry/sentry-go"
)

type jobID int

type Job struct {
	sync.RWMutex
	id      int
	cmd     string
	shell   string
	Spec    string //not used in manager, only for showing info
	process *process.Process
	count   int
	done    chan int //Chan to receive
	run     chan int //Chan to signal manager job run request
	Stats   JobStats
}

//GetID returns job ID
func (j *Job) GetID() int {
	return j.id
}

//GetCmd returns job cmd
func (j *Job) GetCmd() string {
	return j.cmd
}

//Run runs  a Job on the manager and returns not waiting for process fininsh
//We could wait, but seems that external scheduler do not need this by now.
func (j *Job) Run() {
	log.Printf("Job.Run: sending job %d to manager", j.GetID())
	// go j.start()
	j.run <- j.id
}

func (j *Job) start() {
	j.Lock()
	defer j.Unlock()
	if j.process != nil {
		log.Printf("Manager.add: refusing job execution: %s is running", j.process)
		return
	}
	processID := j.count
	p := process.New(j.cmd, j.shell, j.id, processID)
	err := p.Run(j.done)
	if err != nil {
		log.Printf("Run.start: Error launching process %s: %s", p, err)
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("apcron-version", "v")
			scope.SetExtra("jobCmd", j.cmd)
			scope.SetExtra("jobSpec", j.Spec)
			scope.SetExtra("jobExecutions", j.count)
			scope.SetExtra("jobShell", j.shell)
			scope.SetExtra("jobMaxRSS", j.Stats.MaxRss)
			scope.SetExtra("jobMeanRSS", j.Stats.MeanRss)
			scope.SetLevel(sentry.LevelError)
			sentry.CaptureMessage(fmt.Sprintf("apcron: Error running cronjob: %s", err.Error()))
			//sentry.CaptureException(err)
		})
		return
	}
	j.process = p
	j.count++
	log.Printf("Run.start: launched process %s\n", p)
}

func (j *Job) finish() {
	j.Lock()
	defer j.Unlock()
	if j.process == nil {
		log.Printf("Manager.finish: Error getting process for job %d: Job has no process registered", j.id)
		return
	}
	pid := j.process.GetPid()
	j.addStats()
	j.process = nil
	log.Printf("Job.finish ended [%d]. pid %d\n", j.id, pid)
	runtime.GC() //Force Garbage Collection of process.
}

func (j *Job) addStats() {
	j.Stats.AddRss(j.process.GetMaxRss())
}
