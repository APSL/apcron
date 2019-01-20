package process

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/fatih/color"
)

type Process struct {
	sync.RWMutex
	Pid   int
	wg    sync.WaitGroup
	ID    int
	JobID int
	Cmd   string
	cmd   *exec.Cmd
}

func New(cmd string, jobId, id int) *Process {
	// strconv.Itoa(job.ID), job.Cmd
	p := &Process{
		// Done: make(chan struct{}),
		ID:    id,
		JobID: jobId,
		Cmd:   cmd,
	}
	return p
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
		return fmt.Sprintf("[ID%d/R%d/%d]", p.JobID, p.ID, p.GetPid())
	}
	return fmt.Sprintf("[ID%d/R%d/-]", p.JobID, p.ID)
}

//Run exec the process and starts printing its sdout and stderr in parallel.
func (p *Process) Run(done chan int) (err error) {
	log.Printf("Process: starting %s: %s\n", p, p.Cmd)
	p.cmd = exec.Command("xbash", "-c", p.Cmd)

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error reading stout for %s: %v\n", p.Cmd, err)
		return
	}
	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		log.Printf("Error reading stderr for %s: %v\n", p.Cmd, err)
		return
	}

	err = p.cmd.Start()
	if err != nil {
		return
	}

	p.setPid(p.cmd.Process.Pid)
	log.Printf("Process: %s Started\n", p)
	p.wg.Add(2)
	go p.outPrinter(stdout, "[II]", color.FgBlue)
	go p.outPrinter(stderr, "[EE]", color.FgRed)

	go func() {
		p.wg.Wait()        //Waits for outPrinter to exit
		err = p.cmd.Wait() //Waits for os.Exec Cmd to exit. Will close pipes.
		if err != nil {
			log.Printf("process: Exec Error: %s\n", err)
		}
		log.Printf("Process: %s Finished  \n", p)
		done <- p.JobID
	}()
	return nil
}
func (p *Process) outPrinter(r io.Reader, prefix string, c color.Attribute) {
	color := color.New(c).SprintFunc()
	prefix = color(prefix)
	defer p.wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		log.Printf("%s %s %s\n", p, prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "process.outPrinter: Error scanning out for %d:  %v\n", p.ID, err)
	}
}

// func sentryLog(cmd string, out []byte, err error) {
// 	if cfg.SentryDSN != "" {
// 		message := fmt.Sprintln("Error executing command", cmd)
// 		client, err := raven.NewWithTags(cfg.SentryDSN, map[string]string{"program": "go-cron", "error": err.Error(), "command": cmd})
// 		if err == nil {
// 			packet := &raven.Packet{Message: message, Extra: map[string]interface{}{"command": cmd, "output": string(out)}}
// 			client.Capture(packet, nil)
// 		}
// 	}
// }
