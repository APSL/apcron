package manager

import "log"

//JobStats represents Job execution stats
type JobStats struct {
	MeanRss  float64
	MaxRss   float64
	RssCount float64
}

// func NewJobStats() *JobStats {
// 	return &JobStats{0, 0, 0}
// }
//AddRss calculates mean and max of the job, given Maxrss from a process
func (js *JobStats) AddRss(maxRss int64) {
	log.Printf("Adding rss  %d", maxRss)
	js.RssCount++
	rss := float64(maxRss)
	if rss > js.MaxRss {
		js.MaxRss = rss
	}
	js.MeanRss = js.MeanRss + (rss-js.MeanRss)/js.RssCount
}
