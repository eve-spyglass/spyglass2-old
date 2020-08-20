package gui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/Crypta-Eve/spyglass2/config"
	"github.com/Crypta-Eve/spyglass2/feeds"
	"github.com/nats-io/nats.go"
	"strconv"
	"strings"
)

type (
	SpyglassGUI struct {
		app fyne.App
		intelFeed *widget.Group
	}
)

func NewGui() (*SpyglassGUI, error) {

	spg := &SpyglassGUI{}
	w := app.New()

	spg.app = w

	return spg, nil

}

func (ap *SpyglassGUI) ShowMainGui() {

	w := ap.app.NewWindow("Spyglass 2")
	grid := fyne.NewContainerWithLayout(layout.NewGridLayout(2))
	mapGroup := widget.NewGroup("Map")
	grid.AddObject(mapGroup)

	//logGroup := widget.NewGroup("Intel Logs")
	//grid.AddObject(logGroup)
	feed := widget.NewGroupWithScroller("Intel Logs")
	grid.AddObject(feed)
	ap.intelFeed = feed

	w.SetContent(grid)
	w.Show()

}

func (ap *SpyglassGUI) Run() {
	ap.app.Run()
}

func (ap *SpyglassGUI) SubscribeToChatLogs() error {

	nc, err := nats.Connect(strings.Join([]string{config.GetConfig().NatsHost, strconv.Itoa(config.GetConfig().GetNatsPort())}, ":"))
	if err != nil {
		return err
	}

	nc.Subscribe("spyglass.reports.evelog", func(msg *nats.Msg) {
		rep := &feeds.Report{}
		rep.UnmarshalJSON(msg.Data)
		ap.CreateChatLogEntry(*rep)
	})

	return nil

}

func (ap *SpyglassGUI) ShowSplashScreen() chan interface{} {

	done := make(chan interface{})

	var splash fyne.Window

	if drv, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		splash = drv.CreateSplashWindow()
		//gd := fyne.NewContainerWithLayout(layout.NewCenterLayout())
		//gd.AddObject(canvas.NewImageFromResource(resourceLogoPng))
		splash.SetContent(widget.NewIcon(resourceSplashPng))
		splash.Show()
	} else {
		return nil
	}

	go func() {
		// wait for either a send or closing of the channel then remove the splash screen
		_, _ = <-done
		splash.Close()
	}()

	return done
}
