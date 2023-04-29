package main

import (
	"strings"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

func NewTodoInput(id string) doc.InputElement {
	todosinput := doc.Input.WithID(id,"text")
	doc.SetAttribute(todosinput.AsElement(), "placeholder", "What needs to be done?")
	doc.SetAttribute(todosinput.AsElement(), "onfocus", "this.value=''")
	
	doc.Autofocus(todosinput.AsElement())

	todosinput.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
		v,ok:= evt.Value().(ui.Object).Get("value")
		if !ok{
			todosinput.SyncUISetData("value", ui.String(""))
			return false
		}
		s:= v.(ui.String)
		str := strings.TrimSpace(string(s)) // Trim value
		todosinput.SyncUISetData("value", ui.String(str))
		return false
	}))

	todosinput.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
		todosinput := doc.InputElement{evt.CurrentTarget()}
		
		v:= evt.(doc.KeyboardEvent).Key()

		if v == "Enter" {
			evt.PreventDefault()
			if todosinput.Value() != "" {
				todosinput.AsElement().TriggerEvent("newtodo", todosinput.Value())
			}
			todosinput.AsElement().SyncUISetData("value", ui.String(""))
			todosinput.Clear()
		}
		return false
	}))

	return todosinput
}
