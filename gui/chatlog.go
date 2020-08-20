package gui

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/Crypta-Eve/spyglass2/feeds"
	"image"
	"image/color"
	"time"
)

func (ap *SpyglassGUI) CreateChatLogEntry(report feeds.Report) {

	rep := widget.NewLabel(report.Reporter)
	placeholderAvatar := image.NewRGBA(image.Rect(0, 0, 64, 64))
	cyan := color.RGBA{100, 200, 200, 0xff}
	for x := 0; x < 64; x++ {
		for y := 0; y < 64; y++ {
			switch {
			case x < 32 && y < 32:
				placeholderAvatar.Set(x, y, cyan)
			case x >= 32 && y >= 32:
				placeholderAvatar.Set(x, y, color.White)
			}
		}
	}
	avatar := canvas.NewImageFromImage(placeholderAvatar)
	avatar.SetMinSize(fyne.Size{
		Width:  32,
		Height: 32,
	})
	avatar.FillMode = canvas.ImageFillContain
	tim := widget.NewLabelWithStyle(report.Time.Format(time.Kitchen), fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	channel := widget.NewLabel(report.Channel)
	message := widget.NewLabel(report.Info)

	vbox := fyne.NewContainerWithLayout(layout.NewHBoxLayout())

	header := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	header.AddObject(rep)
	header.AddObject(tim)

	body := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	body.AddObject(avatar)
	//body.AddObject(layout.NewSpacer())
	stack := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	stack.AddObject(message)
	stack.AddObject(channel)
	body.AddObject(stack)

	vbox.AddObject(header)
	vbox.AddObject(body)

	ap.intelFeed.Append(vbox)
}
