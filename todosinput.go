package main

import (
	"strings"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

func NewTodoInput(name string, id string) doc.Input {
	todosinput := doc.NewInput("text", name, id)
	doc.SetAttribute(todosinput.AsElement(), "placeholder", "What needs to be done?")
	doc.SetAttribute(todosinput.AsElement(), "autofocus", "")
	doc.SetAttribute(todosinput.AsElement(), "onfocus", "this.value=''")

	todosinput.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
		s := evt.Value().(ui.String)
		str := strings.TrimSpace(string(s)) // Trim value
		todosinput.AsElement().SetDataSetUI("value", ui.String(str))
		return false
	}), doc.NativeEventBridge)

	todosinput.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
		if v:=evt.Value().(ui.String); v == "Enter" {
			evt.PreventDefault()
			if todosinput.Value() != "" {
				todosinput.AsElement().Set("event", "newtodo", todosinput.Value())
			}
			todosinput.AsElement().SetDataSetUI("value", ui.String(""))
			todosinput.Clear()
		}
		return false
	}), doc.NativeEventBridge)

	return todosinput
}
