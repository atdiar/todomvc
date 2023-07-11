package main

import (
	"math/rand"
	"time"
	"strings"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
	. "github.com/atdiar/particleui/drivers/js/declarative"
)

func newIDgenerator(charlen int, seed int64) func() string {
	source := rand.NewSource(seed)
	r := rand.New(source)
	return func() string {
		var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
		b := make([]rune, charlen)
		for i := range b {
			b[i] = letter[r.Intn(len(letter))]
		}
		return string(b)
	}
}


var NewID = newIDgenerator(16,time.Now().UnixNano())

type Todo = ui.Object

func NewTodo(title ui.String) Todo {
	o := ui.NewObject()
	o.Set("id", ui.String(NewID()))
	o.Set("completed", ui.Bool(false))
	o.Set("title", title)
	return o.Commit()
}

type TodoElement struct {
	*ui.Element
}


func FindTodoElement(d doc.Document, t Todo) (TodoElement, bool) {
	todoid, ok := t.Get("id")
	if !ok {
		panic("wrong todo format! id is required")
	}
	todoidstr, ok := todoid.(ui.String)
	if !ok {
		panic("todo id should be a string!")
	}

	todo := d.GetElementById(string(todoidstr))
	if todo != nil {
		return TodoElement{todo}, true
	}
	return TodoElement{todo}, false
}

func newtodo(document doc.Document, id string, options ...string) *ui.Element {

	var li *ui.Element
	var i *ui.Element
	var l *ui.Element
	var b *ui.Element

	t:= E(document.Li.WithID(id,options...),
			Ref(&li),
			Children(
				E(document.Div.WithID(id+"-view"),
					CSS("view"),
					Children(
						E(document.Input.WithID(id+"-completed","checkbox"),
							Ref(&i),
							CSS("toggle"),
						),
						E(document.Label(),
							Ref(&l),
						),
						E(document.Button.WithID(id+"-btn","button"),
							Ref(&b),
							CSS("destroy"),
						),
					),
				),
			),
		)




	edit := document.Input.WithID(id+"-edit", "")
	doc.AddClass(edit.AsElement(), "edit")

	edit.AsElement().BindDeletion(li.AsElement())


	li.Watch("ui","todo", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
	
		t := evt.NewValue().(Todo)

		title, _ := t.Get("title")
		titlestr := title.(ui.String)
		if len(titlestr) == 0 {
			evt.Origin().TriggerEvent("delete")
			return false
		}

		doc.LabelElement{l}.SetText(string(titlestr))
		edit.SetUI("value", titlestr)
		

		todocomplete, ok := t.Get("completed")
		if !ok {
			panic("wrong todo format. Should have completed property")
		}
		todocompletebool := todocomplete.(ui.Bool)

		if todocompletebool {
			doc.AddClass(li.AsElement(), "completed")
		} else {
			doc.RemoveClass(li.AsElement(), "completed")
		}
	
		i.SetUI("checked", todocompletebool)
		

		return false
	}))

	

	

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

		todo = todo.MakeCopy().Set("completed", ui.Bool(complete)).Commit()

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
			edit.AsElement().TriggerEvent("canceledit")
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
			panic("todo should have a data prop")
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

		todo = todo.MakeCopy().Set("title", evt.NewValue()).Commit()
		li.AsElement().SetDataSetUI("todo", todo)
		edit.AsElement().TriggerEvent("edit", ui.Bool(false))
		return false
	}))

	return t

}

func newTodoElement(d doc.Document, t Todo) TodoElement {
	todoid, ok := t.Get("id")
	if !ok {
		panic("missing todo id")
	}
	todoidstr := todoid.(ui.String)

	ntd := newtodo(d, string(todoidstr))
	ntd.SetDataSetUI("todo", t)

	return TodoElement{ntd}
}
