package feeds

import (
	"fmt"
	"github.com/Crypta-Eve/spyglass2/config"
	"github.com/Crypta-Eve/spyglass2/logger"
	"github.com/TomOnTime/utfutil"
	"github.com/mailru/easyjson"
	"github.com/nats-io/nats.go"
	"github.com/rainycape/unidecode"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type (
	ChatLogMon struct {
		chatLogDir string
		logs       map[string]*LogFile

		natsConn *nats.Conn
	}

	LogFile struct {
		LogPath     string
		Active      bool
		ChannelName string
		Listener    string
		StartTime   time.Time
		LastLine    uint64

		info os.FileInfo
	}
)

const (
	encodingHint = utfutil.UTF16LE
	dateformat   = "2006.01.02 15:04:05"
)

func NewChatLogMonitor(chatLogDir string) (*ChatLogMon, error) {
	clm := &ChatLogMon{
		chatLogDir: chatLogDir,
	}

	logs := make(map[string]*LogFile)
	clm.logs = logs

	clm.LoadInitialLogs()

	natsURL := "nats://localhost:" + strconv.Itoa(config.GetConfig().GetNatsPort())
	nts, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}
	clm.natsConn = nts

	return clm, nil
}

func NewLogFile(path string) (*LogFile, error) {

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	data, err := utfutil.ReadFile(path, encodingHint)
	if err != nil {
		return nil, err
	}
	dataDecoded := unidecode.Unidecode(string(data))

	lines := strings.Split(dataDecoded, "\n")

	channel := strings.TrimSpace(strings.Split(lines[7], ":")[1])
	listener := strings.TrimSpace(strings.Split(lines[8], ":")[1])
	start := strings.TrimSpace(strings.SplitN(lines[9], ":", 2)[1])
	startTime, err := time.Parse(dateformat, start)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"err":        err,
			"start":      start,
			"dateformat": dateformat,
			"logfile":    path,
		}).Warn("failed to parse log start time")
		startTime = time.Now()
	}

	lf := &LogFile{
		LogPath:     path,
		Active:      false,
		ChannelName: channel,
		Listener:    listener,
		StartTime:   startTime,
		LastLine:    13, // This is the blank line after the header block, but before the actual intel
		info:        info,
	}

	return lf, nil

}

func (cl *ChatLogMon) LoadInitialLogs() {

	if cl.chatLogDir == "" {
		return
	}

	files, err := ioutil.ReadDir(cl.chatLogDir)
	if err != nil {
		logger.Log.WithField("err", err)
		return
	}

	for _, file := range files {
		lgf, err := NewLogFile(filepath.Join(cl.chatLogDir, file.Name()))
		if err == nil {
			if time.Since(lgf.StartTime) < 25*time.Hour {
				cl.logs[file.Name()] = lgf
				lgf.OnUpdated(file)
				logger.Log.WithField("file", lgf.LogPath).Debug("added log file to list")
			}
		}
	}
}

func (cl *ChatLogMon) pollKnownLogs() {
	for _, log := range cl.logs {
		//logger.Log.WithField("file", name).Debug("checking log")
		fl, err := os.Stat(log.LogPath)
		if err != nil {
			logger.Log.WithField("err", err).Warn("error updating log")
		}

		if fl.ModTime().After(log.info.ModTime()) || fl.Size() > log.info.Size() {
			// This log has been updated, exciting!
			reps := log.OnUpdated(fl)
			cl.sendReportsToNats(reps)
		}
	}
}

func (cl *ChatLogMon) pollForNewLogs() {
	if cl.chatLogDir == "" {
		return
	}

	files, err := ioutil.ReadDir(cl.chatLogDir)
	if err != nil {
		logger.Log.WithField("err", err)
		return
	}

	for _, file := range files {
		if _, ok := cl.logs[file.Name()]; !ok {
			// The file isnt recorded, it must be new
			lgf, err := NewLogFile(filepath.Join(cl.chatLogDir, file.Name()))
			if err == nil {
				if time.Since(lgf.StartTime) < 25*time.Hour {
					cl.logs[file.Name()] = lgf
					reps := lgf.OnUpdated(file)
					cl.sendReportsToNats(reps)
					logger.Log.WithField("file", lgf.LogPath).Debug("added log file to list")
				}
			}
		}
	}
}

func (cl *ChatLogMon) GetChannelNameList() (rooms []string) {
	rooms = make([]string, 0)
	for _, l := range cl.logs {
		exists := false
		for _, cn := range rooms {
			if cn == l.ChannelName {
				exists = true
				break
			}
		}
		if !exists {
			rooms = append(rooms, l.ChannelName)
		}
	}

	return rooms
}

func (cl *ChatLogMon) MonitorLogDir() {

	counter := 0
	const resetCounter = 4
	const sleepTimer = 250 // milliseconds

	for {
		switch counter {
		case 0:
			cl.pollForNewLogs()
		case resetCounter:
			counter = 0
			fallthrough
		default:
			cl.pollKnownLogs()
		}

		counter++
		time.Sleep(sleepTimer * time.Microsecond)
	}

}

func (cl *ChatLogMon) sendReportsToNats(reports []Report) {

	for _, rep := range reports {
		bt, err := easyjson.Marshal(rep)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"err": err,
				"rep": rep,
			}).Warn("failed to marshal report")
		}
		// TODO Make the subject a constant somewhere
		err = cl.natsConn.Publish("spyglass.reports.evelog", bt)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"err": err,
				"rep": rep,
			}).Warn("failed to publish report to nats")
		}
	}

}

func (cl *ChatLogMon) Close() {

}

func (lf *LogFile) OnUpdated(newStat os.FileInfo) ([]Report){

	// TODO: Implement log active to bounce log checking here

	logger.Log.WithField("log", lf.ChannelName).Debug("updated log")
	lf.info = newStat

	// first we need to read in the file
	data, err := utfutil.ReadFile(lf.LogPath, encodingHint)
	if err != nil {
		logger.Log.WithError(err).Warn("failed to read logfile onupdate")
	}
	dataDecoded := unidecode.Unidecode(string(data))

	lines := strings.Split(dataDecoded, "\n")

	logger.Log.WithFields(logrus.Fields{
		"channel":  lf.ChannelName,
		"listener": lf.Listener,
		"lines":    lines[lf.LastLine-1:],
	}).Info("updated log")

	reports := make([]Report, 0)

	for _, line := range lines[lf.LastLine - 1:] {
		rep := parseLogLineToReport(lf, line)
		if rep.Source == "" {
			// Bad report
			continue
		}
		reports = append(reports, rep)
	}

	lf.LastLine = uint64(len(lines))

	return reports
}

func parseLogLineToReport(lf *LogFile, line string) Report {
	if len(line) < 23 {
		return Report{}
	}
	t, err := time.Parse(dateformat, line[2:21])
	if err != nil {
		logger.Log.WithError(err).Warn("invalid log line, couldnt parse date")
		return Report{}
	}

	info := strings.Split(line[24:], ">")


	return Report{
		Time:     t,
		Source:   lf.Listener,
		Channel:  lf.ChannelName,
		Reporter: info[0][:len(info[0])-1],
		Info:     strings.TrimSpace(info[1]),
	}
}