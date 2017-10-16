package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestYamlParser(t *testing.T) {

	s := `
foo:
    second: '3'
    minute: '*/1'
    command: 'foo_command'
bar:
    second: '0'
    minute: '10,30,50'
    command: 'bar_command'
baz:
    minute: '10'
    hour: '3'
    command: 'baz_command'
`
	expectedJobs := []Job{
		Job{Spec: "3 */1 * * * *", Cmd: "foo_command"},
		Job{Spec: "0 10,30,50 * * * *", Cmd: "bar_command"},
		Job{Spec: "0 10 3 * * *", Cmd: "baz_command"},
	}
	strReader := strings.NewReader(s)
	jobs, err := ParseYaml(strReader)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(jobs, expectedJobs) {
		t.Fatalf("Failed parsing yaml jobs:\n--> Expected: %s\n--> Received: %s", expectedJobs, jobs)
	}
	t.Logf("jobs: %#v", jobs)
}
