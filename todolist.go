package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type TodosListElement struct {
	*ui.Element
}

func (t TodosListElement) GetList() ui.List {
	var tdl ui.List
	res, ok := t.AsElement().Get("ui", "todoslist")
	if !ok {
		tdl = ui.NewList().Commit()
	}
	tdl, ok = res.(ui.List)
	if !ok {
		tdl = ui.NewList().Commit()
	}
	return tdl
}

func (t TodosListElement) SetList(tdl ui.List) TodosListElement {
	t.SetDataSetUI("todoslist", tdl)
	return t
}

func(t TodosListElement) UpdateList(tdl ui.List) TodosListElement{
	t.SyncUISyncData("todoslist", tdl)
	t.signalUpdate()
	return t
}

func(t TodosListElement) signalUpdate() TodosListElement{
	t.TriggerEvent("updated")
	return t
}


func TodoListFromRef(ref *ui.Element) TodosListElement{
	return TodosListElement{ref}
}


func(t TodosListElement)AsViewElement() ui.ViewElement{
	return ui.ViewElement{t.AsElement()}
}

func displayWhen(filter string) func(ui.Value) bool{
	return func (v ui.Value)  bool{
		o := v.(Todo)
		cplte, _ := o.Get("completed")
		complete := cplte.(ui.Bool)

		if filter == "active" {
			if complete {
				return false
			}
			return true
		}

		if filter == "completed" {
			if !complete {
				return false
			}
			return true
		}
		return true	
	}
}



func newTodoListElement(document doc.Document, id string, options ...string) *ui.Element {
	t := document.Ul.WithID(id, options...)
	doc.AddClass(t.AsElement(), "todo-list")

	tview := ui.NewViewElement(t.AsElement(), ui.NewView("all"), ui.NewView("active"), ui.NewView("completed"))
	t.OnRouterMounted(func(r *ui.Router){
		names:= ui.NewList(ui.String("all"), ui.String("active"), ui.String("completed")).Commit()
		links:= ui.NewList(
			ui.String(r.NewLink("all").URI()),
			ui.String(r.NewLink("active").URI()),
			ui.String(r.NewLink("completed").URI()),
		).Commit()
		filterslist:=ui.NewObject()
		filterslist.Set("names",names)
		filterslist.Set("urls",links)

		t.AsElement().SetUI("filterslist",filterslist.Commit())
	})

	tview.AsElement().Watch("ui","filter", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		evt.Origin().TriggerEvent("renderlist")
		return false
	}))

	tview.AsElement().Watch("ui","todoslist", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		newlist:= evt.NewValue().(ui.List)

		for _, v := range newlist.UnsafelyUnwrap() {
			o:= v.(Todo)
			ntd, ok := FindTodoElement(doc.GetDocument(evt.Origin()),o)
			if !ok {
				ntd = TodosListElement{evt.Origin()}.NewTodo(o)
			}else{
				ntd.SetDataSetUI("todo",o) // TODO defer ?
			}	
		}

		TodosListElement{evt.Origin()}.signalUpdate()
		evt.Origin().TriggerEvent("renderlist")
		return false
	}))
	
	tview.OnActivated("all", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().SetUI("filter", ui.String("all"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-all")
		return false
	}))
	tview.OnActivated("active", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().SetUI("filter", ui.String("active"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-active")

		return false
	}))
	tview.OnActivated("completed", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().SetUI("filter", ui.String("completed"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-completed")
		return false
	}))

	t.WatchEvent("renderlist", t, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {

		t:= evt.Origin()
		
		// Retrieve current filter
		filterval,ok:= t.Get("ui","filter")
		var filter string
		if !ok{
			filter = "all"
		} else{
			filter = string(filterval.(ui.String))
		}

		// Retrieve current todos list
		todoslist,ok:= t.Get("ui","todoslist")
		var newlist ui.List
		if !ok{
			newlist = ui.NewList().Commit()
		} else{
			newlist = todoslist.(ui.List)
		}
		

		newChildren := make([]*ui.Element, 0, len(newlist.UnsafelyUnwrap()))

		for _, v := range newlist.UnsafelyUnwrap() {
			o:= v.(Todo)
			if displayWhen(filter)(o){
				ntd, ok := FindTodoElement(doc.GetDocument(evt.Origin()),o)
				if !ok {
					panic("todo not found for rendering...")
				}
				newChildren = append(newChildren, ntd.AsElement())
			}
		}
		t.SetChildren(newChildren...)
		
		return false
	}))

	return t.AsElement()
}

func NewTodoList(d doc.Document, id string, options ...string) TodosListElement {
	return TodosListElement{newTodoListElement(d,id, options...)}
}


func(t TodosListElement) NewTodo(o Todo) TodoElement{
	
	ntd := newTodoElement(doc.GetDocument(t.AsElement()),o)
	id, _ := o.Get("id")
	idstr := id.(ui.String)

	t.Watch("ui", "todo", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool { // escalates back to the todolist the data changes issued at the todo Element level
		var tdl ui.List
		res, ok := t.GetUI("todoslist")
		if !ok {
			tdl = ui.NewList().Commit()
		} else {
			tdl = res.(ui.List)
		}

		newval := evt.NewValue()

		rawlist:= tdl.UnsafelyUnwrap()

		for i, rawtodo := range rawlist {
			todo := rawtodo.(Todo)
			oldid, _ := todo.Get("id")

			if oldid == idstr {
				rawlist[i]=newval
				t.SetList(ui.NewListFrom(rawlist))			
				break
			}
		}
		return false
	}))

	t.WatchEvent( "delete", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		var tdl ui.List
		res, ok := t.AsElement().GetUI("todoslist")
		if !ok {
			tdl = ui.NewList().Commit()
		} else {
			tdl = res.(ui.List)
		}
		ntdl := ui.NewList()
		var i int
		
		for _, rawtodo := range tdl.UnsafelyUnwrap() {
			todo := rawtodo.(Todo)
			oldid, _ := todo.Get("id")
			if oldid == idstr {
				t,ok:= FindTodoElement(doc.GetDocument(evt.Origin()),rawtodo.(Todo))
				if ok{
					ui.Delete(t.AsElement())
				}
				continue
			}
			ntdl= ntdl.Append(rawtodo)
			i++
		}

		t.SetList(ntdl.Commit())
		return false
	}))

	return ntd
}