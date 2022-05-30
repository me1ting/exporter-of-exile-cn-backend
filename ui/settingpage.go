package ui

import (
	"log"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

type SettingPage struct {
	*walk.TabPage
}

func NewSettingPage() (*SettingPage, error) {
	var err error
	sp := new(SettingPage)

	if err = (d.TabPage{
		AssignTo: &sp.TabPage,
		Title:    "运行",
		MaxSize:  d.Size{200, 200},
		Layout:   d.VBox{},
		Children: []d.Widget{
			d.Composite{
				Layout: d.Grid{Columns: 3},
				Children: []d.Widget{
					d.Label{Text: "端口："},
					d.LineEdit{
						MaxLength: 5,
						MaxSize:   d.Size{100, 100},
					},
					d.PushButton{},
					d.Label{Text: "补丁："},
					d.LineEdit{},
					d.PushButton{},
				},
			},
			d.Composite{
				Layout: d.HBox{},
				Children: []d.Widget{
					d.PushButton{},
				},
			},
		},
	}.Create(d.NewBuilder(nil))); err != nil {
		log.Fatal(err)
	}

	return sp, nil
}
