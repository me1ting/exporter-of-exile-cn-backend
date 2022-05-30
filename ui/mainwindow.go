package ui

import (
	"log"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

const (
	appMainWindowWidth  = 400
	appMainWindowHeight = 300
)

type AppMainWindow struct {
	window      *walk.MainWindow
	tabs        *walk.TabWidget
	logPage     *LogPage
	settingPage *SettingPage
}

func NewAppMainWindow() (*AppMainWindow, error) {
	var err error

	mw := new(AppMainWindow)
	fixedSize := d.Size{Width: appMainWindowWidth, Height: appMainWindowHeight}

	if err = (d.MainWindow{
		AssignTo: &mw.window,
		Title:    "Exporter of Exile",
		Size:     fixedSize,
		MinSize:  fixedSize,
		Layout: d.VBox{
			Margins: d.Margins{5, 5, 5, 5},
		},
		Children: []d.Widget{
			d.TabWidget{
				AssignTo: &mw.tabs,
			},
		},
	}.Create()); err != nil {
		log.Fatal(err)
	}

	if mw.settingPage, err = NewSettingPage(); err != nil {
		return nil, err
	}
	mw.tabs.Pages().Add(mw.settingPage.TabPage)

	if mw.logPage, err = NewLogPage(); err != nil {
		return nil, err
	}
	mw.tabs.Pages().Add(mw.logPage.TabPage)

	return mw, nil
}

func (app *AppMainWindow) Run() int {
	return app.window.Run()
}
