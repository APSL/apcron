package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"

	"github.com/fatih/color"
	raven "github.com/getsentry/raven-go"
)

type Process struct {
	sync.RWMutex
	Pid  int
	Done chan struct{}
	wg   sync.WaitGroup
	Job  *Job
	Cmd  *exec.Cmd
}

func NewProcess(job *Job) *Process {
	p := &Process{
		Done: make(chan struct{}),
		Pid:  0,
		Job:  job,
	}
	return p
}

func (p *Process) Wait() {
	p.wg.Wait()
}

func (p *Process) setPid(pid int) {
	p.Lock()
	defer p.Unlock()
	p.Pid = pid
}

func (p *Process) GetPid() int {
	p.RLock()
	defer p.RUnlock()
	return p.Pid
}

func (p *Process) String() string {
	if p.Pid != 0 {
		return fmt.Sprintf("[%d-%d]", p.Job.ID, p.GetPid())
	}
	return fmt.Sprintf("[Job%d]", p.Job.ID)
}

func (p *Process) Run(done chan int) int {
	log.Printf("Process: running %s\n", p.Job.Cmd)
	p.Cmd = exec.Command("bash", "-c", p.Job.Cmd)
	stdout, err := p.Cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error reading stout for %s: %v\n", p.Job.Cmd, err)
	}
	stderr, err := p.Cmd.StderrPipe()
	if err != nil {
		log.Printf("Error reading stderr for %s: %v\n", p.Job.Cmd, err)
	}
	red := color.New(color.FgRed).SprintFunc()
	p.wg.Add(2)
	go p.outPrinter(stdout, red("--out-->"))
	go p.outPrinter(stderr, "--err-->")
	p.Cmd.Start()
	if err != nil {
		log.Printf("Error starting  cmd %s: %v\n", p.Job.Cmd, err)
	}
	p.setPid(p.Cmd.Process.Pid)
	log.Printf("Process: %s Started\n", p)
	go func() {
		p.Cmd.Wait()
		// if err != nil {
		// 	log.Printf("Error executing command:%v out:%v err:%v", cmd, string(out), err)
		// 	sentryLog(cmd, out, err)
		// } else if cfg.Verbose {
		// 	log.Printf("cmd:%v out:%v err:%v", cmd, string(out), err)
		// }
		log.Printf("Process: %s Done. waiting for outpritner\n", p)
		p.wg.Wait()
		log.Printf("Process: %s Finished  \n", p)
		done <- p.Pid
		p.Done <- struct{}{}
	}()
	return p.Pid
}
func (p *Process) outPrinter(r io.Reader, prefix string) {
	defer p.wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("[%d] %s %s\n", p.Pid, prefix, scanner.Text())
	}
	// if err := scanner.Err(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "reading stdout for %s:  %v\n", cmd, err)
	// }
}

func sentryLog(cmd string, out []byte, err error) {
	if cfg.SentryDSN != "" {
		message := fmt.Sprintln("Error executing command", cmd)
		client, err := raven.NewWithTags(cfg.SentryDSN, map[string]string{"program": "go-cron", "error": err.Error(), "command": cmd})
		if err == nil {
			packet := &raven.Packet{Message: message, Extra: map[string]interface{}{"command": cmd, "output": string(out)}}
			client.Capture(packet, nil)
		}
	}
}
