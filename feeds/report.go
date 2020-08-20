package feeds

import (
	"errors"
	"time"
)

//easyjson:json
type Report struct {
	Time     time.Time
	Source   string
	Channel  string
	Reporter string
	Info     string
}
type (
	ReportList struct {
		reports []Report
	}
)

func (rl *ReportList) AddReport(rep Report) (added bool, err error) {
	if rl == nil {
		return false, errors.New("report list not yet initialised. nil pointer")
	}

	// Check if we already have a report of this nature
	for _, r := range rl.reports {
		if r.Equals(rep) {
			return false, nil
		}
	}

	// report does not exist yet, add it
	rl.reports = append(rl.reports, rep)
	return true, nil
}

func (r1 Report) Equals(r2 Report) bool {
	equal := r1.Info == r2.Info && r1.Channel == r2.Channel && r1.Reporter == r2.Reporter
	equal = equal && (r1.Time.Sub(r2.Time) < 10*time.Second || r2.Time.Sub(r1.Time) < 10*time.Second)

	return equal
}
