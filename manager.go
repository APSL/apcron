package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

//Manager represent a process Manager
type Manager struct {
	sync.RWMutex
	ProcList map[int]*Process
	Receiver chan *Job
	Done     chan int
	Jobs     map[int]*JobStat
}

type JobStat struct {
	Job     *Job
	Process *Process
	Count   int
	Done    chan struct{}
}

//NewManager return a new process manager
func NewManager() *Manager {
	m := &Manager{
		Receiver: make(chan *Job),
		Done:     make(chan int),
		ProcList: make(map[int]*Process),
		Jobs:     make(map[int]*JobStat),
	}
	return m
}

//RunAndWait runs  a Job on the manager and wait for process end
func (m *Manager) RunAndWait(job *Job) {
	js := m.Run(job)
	log.Printf("Manager.RunAndWait: Waiting for jo %d...\n", job.ID)
	<-js.Done
	log.Printf("Manager.RunAndWait: finished Process %d\n", job.ID)
}

//Run runs  a Job on the manager
func (m *Manager) Run(job *Job) *JobStat {
	log.Println("Manager.Run: sending job to manager")
	m.Receiver <- job
	return m.getJobStat(job)
}

func (m *Manager) getJobStat(job *Job) *JobStat {
	m.Lock()
	defer m.Unlock()
	js, ok := m.Jobs[job.ID]
	if !ok {
		js = &JobStat{
			Job:     job,
			Count:   0,
			Process: nil,
			Done:    make(chan struct{}),
		}
		m.Jobs[job.ID] = js
	}
	return js
}

//Start manager service. Receives Jobs on channel Receiver
func (m *Manager) Start() {
	go func() {
		for {
			select {
			case job := <-m.Receiver:
				// log.Printf("Manager.main: received job %d\n", job.ID)
				go m.add(job)
			case id := <-m.Done:
				// log.Printf("Manager.main: received done signal for job [%d]\n", id)
				go m.finish(id)
			}
		}
	}()
	go func() {
		for {
			time.Sleep(3600 * time.Second)
			//log.Printf("Manager.main: waiting for Jobs. Current: %d\n", len(m.ProcList))
			go m.PrintStats()
			// select {
			// case <-time.After(8 * time.Second):
			// 	log.Printf("Manager.main: waiting for Jobs. Current: %d\n", len(m.ProcList))
			// 	go m.PrintStats()
			// }
		}
	}()
}

func (m *Manager) add(job *Job) {
	js := m.getJobStat(job)
	m.RLock()
	p := js.Process
	m.RUnlock()
	if p != nil {
		log.Printf("Manager.add: refusing new job: %s is running", js.Process)
		js.Done <- struct{}{}
		return
	}
	p = NewProcess(job, js.Count)
	p.Run(m.Done)
	m.Lock()
	js.Count++
	p.JobID = job.ID
	js.Process = p
	m.ProcList[p.Pid] = p
	m.Unlock()
	log.Printf("Manager.add: Appended and launched process [%d]\n", p.Pid)
	m.PrintStats()
}

func (m *Manager) finish(jobID int) {
	m.Lock()
	defer m.Unlock()
	js, ok := m.Jobs[jobID]
	if !ok {
		log.Printf("Manager.finish: Error receiving job stop for %d: job does not exist", jobID)
		return
	}
	if js.Process == nil {
		log.Printf("Manager.finish: Error getting process for job %d: Job has no process registered", jobID)
		return
	}
	pid := js.Process.GetPid()
	delete(m.ProcList, pid)
	js.Process = nil
	js.Done <- struct{}{}
	log.Printf("Manager.finish ended [%d]\n", js.Job.ID)
}

func (m *Manager) PrintStats() {
	m.RLock()
	defer m.RUnlock()
	// for pid, p := range m.ProcList {
	// 	fmt.Printf("[%d] - %v\n", pid, p)
	// }
	// m.RUnlock()
	if len(m.Jobs) == 0 {
		return
	}
	fmt.Println("Current Jobs:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "STATUS", "PID", "COUNT", "CMD"})
	table.SetAutoWrapText(false)
	var id, status, pid, count, cmd string
	for jobID, js := range m.Jobs {
		id = strconv.Itoa(jobID)
		cmd = js.Job.Cmd
		status = "stopped"
		count = strconv.Itoa(js.Count)
		pid = "-"
		if js.Process != nil {
			status = "running"
			pid = strconv.Itoa(js.Process.GetPid())
		}
		data := []string{id, status, pid, count, cmd}
		table.Append(data)
	}
	table.Render()
}
