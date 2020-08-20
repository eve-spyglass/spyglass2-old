package main

import (
	"github.com/Crypta-Eve/spyglass2/config"
	"github.com/Crypta-Eve/spyglass2/evemap"
	"github.com/Crypta-Eve/spyglass2/feeds"
	"github.com/Crypta-Eve/spyglass2/gui"
	"github.com/Crypta-Eve/spyglass2/logger"
	"github.com/Crypta-Eve/spyglass2/nats"
	"os"
	"path/filepath"
	"time"
)

var (
// config definition
)

func main() {

	// Want to be able to recover and show an error to the user
	//defer func() {
	//	if r := recover(); r != nil {
	//		fmt.Println("Recovered in f", r)
	//		os.Exit(-1)
	//	}
	//}()

	os.Setenv("FYNE_SCALE", "0.8")

	spg, _ := gui.NewGui()
	spg.ShowMainGui()
	done := spg.ShowSplashScreen()

	go func() {
		// Set up intial desktop logger
		logger.Configure("DEBUG")
		err := logger.ConfigureLogFile()
		if err != nil {
			logger.Log.Fatal(err)
		}

		// Set up the nats server
		natsPort := config.GetConfig().GetNatsPort()
		logger.Log.WithField("port", natsPort).Info("init internal nats server")

		msg, err := nats.New(natsPort)
		if err != nil {
			logger.Log.WithField("err", err).Panic("failed to init local nats")
		}
		go func() {
			logger.Log.Info("starting internal nats server")
			if err := msg.Run(); err != nil {
				logger.Log.WithField("err", err).Panic("failed to run local nats")
			}
		}()

		// Set up the eve log monitor
		evedir, err := config.FindEveDirectory()
		if err != nil {
			logger.Log.WithField("err", err).Panic("failed to find eve log dir")
		}
		logger.Log.WithField("dir", evedir).Info("evedir loaded")
		chatlogger, err := feeds.NewChatLogMonitor(filepath.Join(evedir, "Chatlogs"))
		if err != nil {
			logger.Log.WithField("err", err).Panic("failed to launch chatlogmon")
		}
		// TODO: If no chatrooms reported here then we need to prompt the user and say that they need to set the eve dir
		go chatlogger.MonitorLogDir()
		defer chatlogger.Close()

		newEden, err := evemap.CreateNewEden()
		if err != nil {
			logger.Log.WithError(err).Fatal("failed to create new eden")
		}

		_ = newEden.Maps

		// close the splash screen
		close(done)

		// Just for now until we have the wait loop of the gui
		time.Sleep(5 * time.Minute)

	}()

	go spg.SubscribeToChatLogs()
	spg.Run()

	logger.Log.Info("spyglass gracefully exiting")
}
