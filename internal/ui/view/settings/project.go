package settings

import (
	"trkr"
	"trkr/internal/events"
	"trkr/internal/ui"
	"trkr/internal/ui/widget"
)

var uiElem *ui.Element

func CreateProjectDialog(parent *ui.Element) *ui.Element {
	core := ui.NewElementCoreInstance(showProject, hideProject, projectHandleInputs, projectDraw)
	uiElem = ui.NewElement(0, 0, 0, 0, core, parent)
	uiElem.TopPadding = 30
	uiElem.LeftPadding = 10
	uiElem.Visible = false
	_ = widget.NewInput("Name: ", 0, 0, 150, 26, func(a any) bool {
		trkr.Logf("Setting project filename to %s.\n", a)
		trkr.CurrentProject.Filename = a.(string) + ".json"
		return true
	}, uiElem)
	return uiElem
}

func showProject()                                                        {}
func hideProject()                                                        {}
func projectHandleInputs(input events.InputSnapshot, el *ui.Element) bool { return false }
func projectDraw(ctx events.EventContext, hasFocus bool) bool {
	uiElem.DrawContainer(ctx)
	return false
}
