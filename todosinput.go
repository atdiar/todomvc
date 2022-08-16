package main

import (
	"strings"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

func NewTodoInput(name string, id string) doc.Input {
	todosinput := doc.NewInput("text", name, id)
	doc.SetAttribute(todosinput.AsElement(), "placeholder", "What needs to be done?")
	//doc.SetAttribute(todosinput.AsElement(), "autofocus", "")
	doc.SetAttribute(todosinput.AsElement(), "onfocus", "this.value=''")
	
	/*todosinput.AsElement().OnFirstTimeMounted(ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		evt.Origin().Watch("event","navigationend",evt.Origin().Root(),ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
			doc.FocusAndScrollOnlyIfNecessary(evt.Origin())
			return false
		}))
		
		return false
	}))*/
	doc.Autofocus(todosinput.AsElement())

	todosinput.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
		v,ok:= evt.Value().(ui.Object).Get("value")
		ui.DEBUG(evt.Value().(ui.Object))
		if !ok{
			panic("framework error: unable to find change event value")
		}
		s:= v.(ui.String)
		str := strings.TrimSpace(string(s)) // Trim value
		todosinput.AsElement().SetDataSetUI("value", ui.String(str))
		return false
	}))

	todosinput.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
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
