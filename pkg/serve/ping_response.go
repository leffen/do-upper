package serve

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type pingResponse struct {
	pingNum     int
	url         string
	time        time.Time
	dur         time.Duration
	totalDur    time.Duration
	dnsDuration time.Duration
	tlsDuration time.Duration
	connectTime time.Duration
}

type pingResponseNotifier interface {
	OnStart() error
	OnNewMeasurement(p pingResponse) error
	OnEnd() error
}

func (p *pingResponse) clearTimers() {
	p.dur = 0
	p.dnsDuration = 0
	p.tlsDuration = 0
	p.connectTime = 0
	p.time = time.Now()
}

func (p *pingResponse) toJSON() (string, error) {
	resp := struct {
		PingNum       int    `json:"ping_num"`
		Time          string `json:"time"`
		URL           string `json:"url"`
		Duration      int64  `json:"duration_ns"`
		TotalDuration int64  `json:"total_dutation_ns"`
		DNSDuration   int64  `json:"dns_duration_ns"`
		TLSDuration   int64  `json:"tls_duartion_ns"`
		ConnectTime   int64  `json:"connect_time_ns"`
	}{
		p.pingNum, p.time.Format(time.RFC3339), p.url, p.dur.Nanoseconds(), p.totalDur.Nanoseconds(), p.dnsDuration.Nanoseconds(), p.tlsDuration.Nanoseconds(), p.connectTime.Nanoseconds(),
	}

	b, err := json.Marshal(resp)
	return string(b), err
}

func (p *pingResponse) debugOut() {
	logrus.Debug(p.asString())
}

func (p *pingResponse) asString() string {
	return fmt.Sprintf("%06d %s %-30.30s Total: %v First byte after: %v  DNS: %v TLS Handshake:%v Connect: %v \n", p.pingNum, p.time.Format(time.RFC3339), p.url, p.totalDur, p.dur, p.dnsDuration, p.tlsDuration, p.connectTime)
}
