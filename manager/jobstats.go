package manager

import (
	"fmt"
)

//JobStats represents Job execution stats
type JobStats struct {
	MeanRss  float64
	MaxRss   float64
	RssCount float64
}

// func NewJobStats() *JobStats {
// 	return &JobStats{0, 0, 0}
// }

func (js *JobStats) AddRss(irss int64) {
	fmt.Printf("Adding rss %d", irss)
	js.RssCount++
	rss := float64(irss)
	if rss > js.MaxRss {
		js.MaxRss = rss
	}
	js.MeanRss = js.MeanRss + (rss-js.MeanRss)/js.RssCount
}
