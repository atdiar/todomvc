package main

import (
	"strings"
	"time"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type Todo = ui.Object

func NewTodo(title ui.String) Todo {
	NewID := ui.NewIDgenerator(time.Now().UnixNano())
	o := ui.NewObject()
	o.Set("id", ui.String(NewID()))
	o.Set("completed", ui.Bool(false))
	o.Set("title", title)
	return o
}

type TodoElement struct {
	ui.BasicElement
}

func FindTodoElement(t Todo) (TodoElement, bool) {
	todoid, ok := t.Get("id")
	if !ok {
		return TodoElement{}, false
	}
	todoidstr, ok := todoid.(ui.String)
	if !ok {
		return TodoElement{}, false
	}

	todo, ok := doc.Elements.ByID[string(todoidstr)]
	if ok {
		return TodoElement{ui.BasicElement{todo}}, true
	}
	return TodoElement{ui.BasicElement{todo}}, false
}

func NewTodoElement(t Todo) TodoElement {
	todoid, ok := t.Get("id")
	if !ok {
		return TodoElement{}
	}
	todoidstr, ok := todoid.(ui.String)
	if !ok {
		return TodoElement{}
	}

	newtodo := doc.Elements.NewConstructor("todo", func(name string, id string) *ui.Element {
		d := doc.NewDiv(name, id+"-view")
		doc.AddClass(d.AsElement(), "view")

		i := doc.NewInput("checkbox", name, id+"-completed")
		doc.AddClass(i.AsElement(), "toggle")

		edit := doc.NewInput("", name, id+"-edit")
		doc.AddClass(edit.AsElement(), "edit")

		l := doc.NewLabel(name, id+"-lbl")

		b := doc.NewButton(name, id+"-btn", "button")
		doc.AddClass(b.AsElement(), "destroy")

		d.SetChildren(i, l, b)
		li := doc.NewListItem(name, id).SetValue(d.AsElement())

		edit.AsElement().ShareLifetimeOf(li)

		li.AsElement().Watch("ui", "todo", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			t, ok := evt.NewValue().(ui.Object)
			if !ok {
				return true
			}

			_, ok = t.Get("id")
			if !ok {
				return true
			}

			todocomplete, ok := t.Get("completed")
			if !ok {
				return true
			}
			todocompletebool := todocomplete.(ui.Bool)

			if todocompletebool {
				doc.AddClass(li.AsElement(), "completed")
			} else {
				doc.RemoveClass(li.AsElement(), "completed")
			}

			todotitle, ok := t.Get("title")
			if !ok {
				return true
			}
			todotitlestr := todotitle.(ui.String)

			i.AsElement().SetUI("checked", todocompletebool)
			l.SetText(string(todotitlestr))
			edit.AsElement().SetDataSetUI("value", todotitlestr)

			return false
		}))

		li.AsElement().Watch("event", "toggle", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			res, ok := li.AsElement().GetData("todo")
			if !ok {
				panic("Cannot find corresponding todo element.")
			}
			todo := res.(ui.Object)

			b, ok := todo.Get("completed")
			if !ok {
				panic("wrong todo format. Should have completed property")
			}
			complete := !(b.(ui.Bool))

			newtodo := ui.Copy(todo).(ui.Object).Set("completed", ui.Bool(complete))

			li.AsElement().SetDataSetUI("todo", newtodo)
			return false
		}))

		li.AsElement().Watch("event", "edit", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			li.AsElement().SetDataSetUI("editmode", ui.Bool(true))
			return false
		}))

		li.AsElement().Watch("ui", "editmode", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			m := evt.NewValue().(ui.Bool)
			if m {
				doc.AddClass(li.AsElement(), "editing")
				li.AsElement().AppendChild(edit.AsElement())
				edit.Focus()
			} else {
				doc.RemoveClass(li.AsElement(), "editing")
				li.AsElement().RemoveChild(edit.AsElement())
			}
			return false
		}))

		i.AsElement().AddEventListener("click", ui.NewEventHandler(func(evt ui.Event) bool {
			//evt.PreventDefault()
			li.AsElement().Set("event", "toggle", ui.Bool(true))
			return false
		}))

		l.AsElement().AddEventListener("dblclick", ui.NewEventHandler(func(evt ui.Event) bool {
			li.AsElement().Set("event", "edit", ui.Bool(true))
			return false
		}))

		b.AsElement().AddEventListener("click", ui.NewEventHandler(func(evt ui.Event) bool {
			li.AsElement().Set("event", "delete", ui.Bool(true))
			return false
		}))

		edit.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
			v,ok:= evt.Value().(ui.Object).Get("value")
			if !ok{
				panic("framework error: unable to find change event value")
			}
			s:= v.(ui.String)
			str := strings.TrimSpace(string(s)) // Trim value
			edit.AsElement().SetDataSetUI("value", ui.String(str))
			return false
		}))

		edit.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
			val,ok:= evt.Value().(ui.Object).Get("key")
			if !ok{
				panic("framework error: unable to find event key")
			}
			if v:=val.(ui.String); v == "Escape" {
				evt.PreventDefault()
				edit.AsElement().Set("event", "canceledit", ui.Bool(true))
				return false
			}
			if v:=val.(ui.String);v == "Enter" {
				evt.PreventDefault()
				edit.Blur()
			}
			return false
		}))

		edit.AsElement().AddEventListener("blur", ui.NewEventHandler(func(evt ui.Event) bool {
			edit.AsElement().Set("event", "newtitle", edit.Value())
			return false
		}))

		li.AsElement().Watch("event", "edit", edit, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			li.AsElement().Set("ui", "editmode", evt.NewValue())
			return false
		}))

		li.AsElement().Watch("event", "canceledit", edit, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			res, ok := li.AsElement().GetData("todo")
			if !ok {
				return true
			}
			todo := res.(Todo)
			val, _ := todo.Get("title")
			edit.AsElement().SetDataSetUI("value", val.(ui.String))
			edit.Blur()
			return false
		}))

		li.AsElement().Watch("event", "newtitle", edit, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			res, ok := li.AsElement().GetData("todo")
			if !ok {
				// TODO maybe a debug statement although it should not happen.
				return true
			}
			todo := res.(Todo)

			todo.Set("title", evt.NewValue())
			li.AsElement().SetDataSetUI("todo", todo)
			edit.AsElement().Set("event", "edit", ui.Bool(false))
			return false
		}))

		return li.AsElement()

	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

	ntd := doc.LoadFromStorage(newtodo("todo", string(todoidstr))).Memoize()
	ntd.SetDataSetUI("todo", t)

	return TodoElement{ui.BasicElement{ntd}}
}
