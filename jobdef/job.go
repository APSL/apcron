package jobdef

// Job is the cron spec to be parsed by a cron parser
// Spec is the crontab format string
// Cmd is the command to be executed
type Job struct {
	Spec  string
	Cmd   string
	Shell string
}
