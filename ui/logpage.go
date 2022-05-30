package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/me1ting/exporter-of-exile-cn-backend/log"
)

const (
	maxLogLinesDisplayed = 10000
)

type LogPage struct {
	*walk.TabPage
	logView *walk.TableView
	model   *logModel
}

func NewLogPage() (*LogPage, error) {
	lp := &LogPage{}

	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	if lp.TabPage, err = walk.NewTabPage(); err != nil {
		return nil, err
	}
	disposables.Add(lp)

	lp.Disposing().Attach(func() {
		lp.model.quit <- true
	})

	lp.SetTitle("日志")
	lp.SetLayout(walk.NewVBoxLayout())

	if lp.logView, err = walk.NewTableView(lp); err != nil {
		return nil, err
	}
	lp.logView.SetAlternatingRowBG(true)
	lp.logView.SetLastColumnStretched(true)
	lp.logView.SetGridlines(true)

	contextMenu, err := walk.NewMenu()
	if err != nil {
		return nil, err
	}
	lp.logView.AddDisposable(contextMenu)
	copyAction := walk.NewAction()
	copyAction.SetText("复制")
	copyAction.SetShortcut(walk.Shortcut{walk.ModControl, walk.KeyC})
	copyAction.Triggered().Attach(lp.onCopy)
	contextMenu.Actions().Add(copyAction)
	lp.ShortcutActions().Add(copyAction)
	selectAllAction := walk.NewAction()
	selectAllAction.SetText("全选")
	selectAllAction.SetShortcut(walk.Shortcut{walk.ModControl, walk.KeyA})
	selectAllAction.Triggered().Attach(lp.onSelectAll)
	contextMenu.Actions().Add(selectAllAction)
	lp.ShortcutActions().Add(selectAllAction)
	lp.logView.SetContextMenu(contextMenu)
	setSelectionStatus := func() {
		copyAction.SetEnabled(len(lp.logView.SelectedIndexes()) > 0)
		selectAllAction.SetEnabled(len(lp.logView.SelectedIndexes()) < len(lp.model.items))
	}
	lp.logView.SelectedIndexesChanged().Attach(setSelectionStatus)

	stampCol := walk.NewTableViewColumn()
	stampCol.SetName("Time")
	stampCol.SetTitle("时间")
	stampCol.SetFormat("2006-01-02 15:04:05.000")
	stampCol.SetWidth(140)
	lp.logView.Columns().Add(stampCol)

	msgCol := walk.NewTableViewColumn()
	msgCol.SetName("Message")
	msgCol.SetTitle("日志消息")
	lp.logView.Columns().Add(msgCol)

	lp.model = newLogModel(lp)
	lp.model.RowsReset().Attach(setSelectionStatus)
	lp.logView.SetModel(lp.model)
	setSelectionStatus()

	disposables.Spare()

	return lp, nil
}

func (lp *LogPage) isAtBottom() bool {
	return len(lp.model.items) == 0 || lp.logView.ItemVisible(len(lp.model.items)-1)
}

func (lp *LogPage) scrollToBottom() {
	lp.logView.EnsureItemVisible(len(lp.model.items) - 1)
}

func (lp *LogPage) onCopy() {
	var logLines strings.Builder
	selectedItemIndexes := lp.logView.SelectedIndexes()
	if len(selectedItemIndexes) == 0 {
		return
	}
	for i := 0; i < len(selectedItemIndexes); i++ {
		logItem := lp.model.items[selectedItemIndexes[i]]
		logLines.WriteString(fmt.Sprintf("%s: %s\r\n", logItem.Time.Format("2006-01-02 15:04:05.000"), logItem.Message))
	}
	walk.Clipboard().SetText(logLines.String())
}

func (lp *LogPage) onSelectAll() {
	lp.logView.SetSelectedIndexes([]int{-1})
}

type logModel struct {
	walk.ReflectTableModelBase
	lp    *LogPage
	quit  chan bool
	items []*log.LogEntry
}

func newLogModel(lp *LogPage) *logModel {
	mdl := &logModel{lp: lp, quit: make(chan bool)}
	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				var items []*log.LogEntry
				items = log.Global.EmptyBuffer()
				if len(items) == 0 {
					continue
				}
				mdl.lp.Synchronize(func() {
					isAtBottom := mdl.lp.isAtBottom() && len(lp.logView.SelectedIndexes()) <= 1

					mdl.items = append(mdl.items, items...)
					if len(mdl.items) > maxLogLinesDisplayed {
						mdl.items = mdl.items[len(mdl.items)-maxLogLinesDisplayed:]
					}
					mdl.PublishRowsReset()

					if isAtBottom {
						mdl.lp.scrollToBottom()
					}
				})

			case <-mdl.quit:
				ticker.Stop()
				break
			}
		}
	}()

	return mdl
}

func (mdl *logModel) Items() interface{} {
	return mdl.items
}
