package manager

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	//"github.com/shirou/gopsutil/process"
)

const MaxJobs int = 100

//Manager represent a process Manager
type Manager struct {
	sync.RWMutex
	run       chan int
	done      chan int
	Jobs      [MaxJobs]*Job
	jobsCount int
}

//New return a new process manager
func New() *Manager {
	m := &Manager{
		run:  make(chan int), //channel to signal running a job
		done: make(chan int),
	}
	return m
}

//RunAndWait runs  a Job on the manager and wait for process end

//CreateJob creates a new manager.Job from a command string
func (m *Manager) CreateJob(cmd, spec, shell string) (j *Job, err error) {
	m.Lock()
	defer m.Unlock()

	if m.jobsCount >= MaxJobs {
		err = fmt.Errorf("Manager.CreateJob: Error creating job: Max jobs reached")
		return
	}

	j = &Job{
		cmd:     cmd,
		Spec:    spec,
		shell:   shell,
		count:   0,
		process: nil,
		done:    m.done,
		run:     m.run,
		id:      m.jobsCount,
	}
	m.Jobs[m.jobsCount] = j
	m.jobsCount++
	return
}

//getJob return job from jobID
func (m *Manager) getJob(id int) (j *Job, ok bool) {
	m.RLock()
	defer m.RUnlock()
	if id >= m.jobsCount {
		log.Printf("Manager.getJob: Error getting jobID  %d >= jobCount=%d\n", id, m.jobsCount)
		return
	}
	j = m.Jobs[id]
	if j == nil {
		log.Printf("Manager.getJob: Error getting jobID")
		return
	}
	ok = true
	return
}

//Start manager service. Receives Jobs on channel Receiver
func (m *Manager) Start() {
	go func() {
		for {
			select {
			case jobID := <-m.run:
				log.Printf("Manager.main: received run signal for job %d\n", jobID)
				if job, ok := m.getJob(jobID); ok {
					go job.start()
				}
			case jobID := <-m.done:
				log.Printf("Manager.main: received done signal for job %d\n", jobID)
				if job, ok := m.getJob(jobID); ok {
					go job.finish()
				}
			}
		}
	}()
	go func() {
		for {
			time.Sleep(4 * 60 * time.Second)
			go m.PrintStats()
		}
	}()
}

func (m *Manager) PrintStats() {
	m.RLock()
	defer m.RUnlock()
	if m.jobsCount == 0 {
		return
	}
	fmt.Println("Current Jobs:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "STATUS", "PID", "COUNT", "MEANRSS", "MAXRSS", "CMD"})
	table.SetAutoWrapText(false)
	var id, status, pid, count, cmd string
	// for jobID, js := range m.Jobs {
	for jid := 0; jid < m.jobsCount; jid++ {
		id = strconv.Itoa(jid)
		job, _ := m.getJob(jid)
		cmd = job.cmd
		status = "stopped"
		count = strconv.Itoa(job.count)
		pid = "-"
		if job.process != nil {
			status = "running"
			pid = strconv.Itoa(job.process.GetPid())
		}
		meanrss := fmt.Sprintf("%.2f MB", job.Stats.MeanRss/1024)
		maxrss := fmt.Sprintf("%.2f MB", job.Stats.MaxRss/1024)
		data := []string{id, status, pid, count, meanrss, maxrss, cmd}
		table.Append(data)
	}
	table.Render()
}
