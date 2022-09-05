package main

import (
	"strings"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

func NewTodoInput(id string) doc.InputElement {
	todosinput := doc.Input("text", id)
	doc.SetAttribute(todosinput.AsElement(), "placeholder", "What needs to be done?")
	//doc.SetAttribute(todosinput.AsElement(), "autofocus", "")
	doc.SetAttribute(todosinput.AsElement(), "onfocus", "this.value=''")
	
	doc.Autofocus(todosinput.AsElement())

	todosinput.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
		v,ok:= evt.Value().(ui.Object).Get("value")
		if !ok{
			panic("framework error: unable to find change event value")
		}
		s:= v.(ui.String)
		str := strings.TrimSpace(string(s)) // Trim value
		todosinput.AsElement().SetDataSetUI("value", ui.String(str))
		return false
	}))

	todosinput.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
		todosinput := doc.InputElement{ui.BasicElement{evt.CurrentTarget()}}
		val,ok:= evt.Value().(ui.Object).Get("key")
		if !ok{
			panic("framework error: unable to find event key")
		}

		if v:=val.(ui.String); v == "Enter" {
			evt.PreventDefault()
			if todosinput.Value() != "" {
				todosinput.AsElement().Set("event", "newtodo", todosinput.Value())
			}
			todosinput.AsElement().SetDataSetUI("value", ui.String(""))
			todosinput.Clear()
		}
		return false
	}))

	return todosinput
}
