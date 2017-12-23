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
	Pid int
	// Done chan struct{}
	wg    sync.WaitGroup
	ID    int
	JobID int
	Cmd   string
	cmd   *exec.Cmd
}

func NewProcess(job *Job, id int) *Process {
	// strconv.Itoa(job.ID), job.Cmd
	p := &Process{
		// Done: make(chan struct{}),
		ID:    id,
		JobID: job.ID,
		Cmd:   job.Cmd,
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
		return fmt.Sprintf("[%d/%d/%d]", p.JobID, p.ID, p.GetPid())
	}
	return fmt.Sprintf("[%d/%d/-]", p.JobID, p.ID)
}

func (p *Process) Run(done chan int) int {
	log.Printf("Process: starting %s: %s\n", p, p.Cmd)
	p.cmd = exec.Command("bash", "-c", p.Cmd)
	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error reading stout for %s: %v\n", p.Cmd, err)
	}
	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		log.Printf("Error reading stderr for %s: %v\n", p.Cmd, err)
	}
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	p.wg.Add(2)
	go p.outPrinter(stdout, blue("[II]"))
	go p.outPrinter(stderr, red("[EE]"))
	p.cmd.Start()
	if err != nil {
		log.Printf("Error starting  cmd %s: %v\n", p.Cmd, err)
	}
	p.setPid(p.cmd.Process.Pid)
	log.Printf("Process: %s Started\n", p)
	go func() {
		p.cmd.Wait()
		// if err != nil {
		// 	log.Printf("Error executing command:%v out:%v err:%v", cmd, string(out), err)
		// 	sentryLog(cmd, out, err)
		// } else if cfg.Verbose {
		// 	log.Printf("cmd:%v out:%v err:%v", cmd, string(out), err)
		// }
		log.Printf("Process: %s Done. waiting for stdout and stderr printers to end...\n", p)
		p.wg.Wait()
		log.Printf("Process: %s Finished  \n", p)
		done <- p.JobID
		//p.Done <- struct{}{}
	}()
	return p.Pid
}
func (p *Process) outPrinter(r io.Reader, prefix string) {
	defer p.wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("%s %s %s\n", p, prefix, scanner.Text())
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
