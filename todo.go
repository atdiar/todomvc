package main

import (
	"strings"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type Todo = ui.Object

func NewTodo(title ui.String) Todo {
	NewID := doc.Elements.NewID
	o := ui.NewObject()
	o.Set("id", ui.String(NewID()))
	o.Set("completed", ui.Bool(false))
	o.Set("title", title)
	return o
}

type TodoElement struct {
	*ui.Element
}


func FindTodoElement(d *doc.Document, t Todo) (TodoElement, bool) {
	todoid, ok := t.Get("id")
	if !ok {
		return TodoElement{}, false
	}
	todoidstr, ok := todoid.(ui.String)
	if !ok {
		return TodoElement{}, false
	}

	todo := d.GetElementById(string(todoidstr))
	if todo != nil {
		return TodoElement{todo}, true
	}
	return TodoElement{todo}, false
}

var newtodo = doc.Elements.NewConstructor("todo", func(id string) *ui.Element {
	d := doc.Div.WithID(id + "-view")
	doc.AddClass(d.AsElement(), "view")

	i := doc.Input.WithID(id+"-completed", "checkbox")
	doc.AddClass(i.AsElement(), "toggle")

	edit := doc.Input.WithID(id+"-edit", "")
	doc.AddClass(edit.AsElement(), "edit")

	l := doc.Label.WithID(id + "-lbl")

	b := doc.Button.WithID(id+"-btn", "button")
	doc.AddClass(b.AsElement(), "destroy")

	d.SetChildren(i, l, b)
	li := doc.Li.WithID(id).SetChildren(d.AsElement())

	edit.AsElement().BindDeletion(li.AsElement())


	li.Watch("ui","todo", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
	
	
		t := evt.NewValue().(Todo)

		title, _ := t.Get("title")
		titlestr := title.(ui.String)
		if len(titlestr) == 0 {
			evt.Origin().TriggerEvent("delete")
			return false
		}

		l.SetText(string(titlestr))
		edit.SetUI("value", titlestr)
		

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

		i.AsElement().SetUI("checked", todocompletebool)

		return false
	}).RunASAP())

	

	

	li.WatchEvent("toggle", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		res, ok := evt.Origin().Get("ui", "todo")
		if !ok {
			panic("Cannot find corresponding todo element.")
		}
		todo := res.(ui.Object)

		b, ok := todo.Get("completed")
		if !ok {
			panic("wrong todo format. Should have completed property")
		}
		complete := !(b.(ui.Bool))

		todo.Set("completed", ui.Bool(complete))

		evt.Origin().SetDataSetUI("todo", todo)

		return false
	}))

	li.AsElement().WatchEvent("edit", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		li.AsElement().SetUI("editmode", ui.Bool(true))
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
		li.TriggerEvent("toggle")
		return false
	}))

	l.AsElement().AddEventListener("dblclick", ui.NewEventHandler(func(evt ui.Event) bool {
		li.AsElement().TriggerEvent("edit", ui.Bool(true))
		return false
	}))

	b.AsElement().AddEventListener("click", ui.NewEventHandler(func(evt ui.Event) bool {
		li.AsElement().TriggerEvent("delete", ui.Bool(true))
		return false
	}))

	edit.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {

		v, ok := evt.Value().(ui.Object).Get("value")
		if !ok {
			edit.SetUI("value", ui.String(""))
			return false
		}
		s := v.(ui.String)
		str := strings.TrimSpace(string(s)) // Trim value
		edit.AsElement().SetUI("value", ui.String(str))
		return false
	}))

	edit.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
		val, ok := evt.Value().(ui.Object).Get("key")
		if !ok {
			panic("framework error: unable to find event key")
		}
		if v := val.(ui.String); v == "Escape" {
			evt.PreventDefault()
			edit.AsElement().TriggerEvent("canceledit", ui.Bool(true))
			return false
		}
		if v := val.(ui.String); v == "Enter" {
			evt.PreventDefault()
			edit.Blur()
		}
		return false
	}))

	edit.AsElement().AddEventListener("blur", ui.NewEventHandler(func(evt ui.Event) bool {
		edit.AsElement().TriggerEvent("newtitle", edit.Value())
		return false
	}))

	li.AsElement().WatchEvent("edit", edit, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		li.AsElement().Set("ui", "editmode", evt.NewValue())
		return false
	}))

	li.AsElement().WatchEvent("canceledit", edit, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		res, ok := li.AsElement().GetData("todo")
		if !ok {
			return true
		}
		todo := res.(Todo)
		val, _ := todo.Get("title")
		edit.AsElement().SetUI("value", val.(ui.String))
		edit.Blur()
		return false
	}))

	li.AsElement().WatchEvent("newtitle", edit, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		res, ok := li.AsElement().GetData("todo")
		if !ok {
			// TODO maybe a debug statement although it should not happen.
			return true
		}
		todo := res.(Todo)

		todo.Set("title", evt.NewValue())
		li.AsElement().SetDataSetUI("todo", todo)
		edit.AsElement().TriggerEvent("edit", ui.Bool(false))
		return false
	}))

	return li.AsElement()

}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

func NewTodoElement(t Todo) TodoElement {
	todoid, ok := t.Get("id")
	if !ok {
		panic("missing todo id")
	}
	todoidstr := todoid.(ui.String)

	ntd := doc.LoadFromStorage(newtodo(string(todoidstr)))
	ntd.SetDataSetUI("todo", t)

	return TodoElement{ntd}
}
