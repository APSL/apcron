package manager

import (
	"log"
	"runtime"
	"sync"

	"github.com/apsl/apcron/process"
)

type jobID int

type Job struct {
	sync.RWMutex
	id      int
	cmd     string
	Spec    string //not used in manager, only for showing info
	process *process.Process
	count   int
	done    chan int //Chan to receive
	run     chan int //Chan to signal manager job run request
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
	log.Println("Job.Run: sending job to manager")
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
	p := process.New(j.cmd, j.id, processID)
	err := p.Run(j.done)
	if err != nil {
		log.Printf("Run.start: Error launching process %s: %s", p, err)
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
	j.process = nil
	log.Printf("Job.finish ended [%d]. pid %d\n", j.id, pid)
	runtime.GC() //Force Garbage Collection of process.
}
