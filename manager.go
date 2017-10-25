package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

//Manager represent a process Manager
type Manager struct {
	sync.RWMutex
	ProcList map[int]*Process
	Receiver chan *Process
	Done     chan int
}

//NewManager return a new process manager
func NewManager() *Manager {
	m := &Manager{
		Receiver: make(chan *Process),
		Done:     make(chan int),
		ProcList: make(map[int]*Process),
	}
	return m
}

//RunAndWait runs  a Job on the manager and wait for process end
func (m *Manager) RunAndWait(job *Job) {
	p := m.Run(job)
	<-p.Done
	log.Printf("Manager.RunAndWait: finished Process %s\n", p)
}

//Run runs  a Job on the manager
func (m *Manager) Run(job *Job) *Process {
	log.Println("Manager.Run: sending process to manager")
	p := NewProcess(job)
	m.Receiver <- p
	return p
}

//Start manager service. Receives Jobs on channel Receiver
func (m *Manager) Start() {
	go func() {
		for {
			select {
			case <-time.After(2 * time.Second):
				log.Printf("Manager.main: waiting for Jobs. Current: %d\n", len(m.ProcList))
				go m.PrintStats()
			case p := <-m.Receiver:
				log.Printf("Manager.main: received process %s\n", p)
				go m.add(p)
			case pid := <-m.Done:
				log.Printf("Manager.main: received done signal for [%d]\n", pid)
				go m.finish(pid)
			}
		}
	}()
}

func (m *Manager) add(p *Process) {
	pid := p.Run(m.Done)
	m.Lock()
	m.ProcList[pid] = p
	m.Unlock()
	log.Printf("Manager.add: Appended and launched process [%d]\n", p.Pid)
}

func (m *Manager) finish(pid int) {
	m.Lock()
	delete(m.ProcList, pid)
	m.Unlock()
	log.Printf("Manager.finish: deleted [%d]\n", pid)
}

func (m *Manager) PrintStats() {
	m.RLock()
	for pid, p := range m.ProcList {
		fmt.Printf("[%d] - %v\n", pid, p)
	}
	m.RUnlock()
}
