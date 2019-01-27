package parser

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/apsl/apcron/jobdef"
)

const (
	// Specs must have 6 fields, the first is seconds
	lineRegexp = `^(\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.+)$`
)

// ParseCron parses crontab format and returns a Runner
func ParseCron(r io.Reader) (jobs []jobdef.Job, err error) {
	rp, _ := regexp.Compile(lineRegexp)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		if rp.MatchString(line) == true {
			m := rp.FindStringSubmatch(line)
			jobdef := jobdef.Job{
				Spec: m[1],
				Cmd:  m[2],
			}
			jobs = append(jobs, jobdef)
		}
	}

	if len(jobs) == 0 {
		err = errors.New("No valid crontab lines found")
	}
	return
}
