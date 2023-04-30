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
		tdl = ui.NewList()
	}
	tdl, ok = res.(ui.List)
	if !ok {
		tdl = ui.NewList()
	}
	return tdl
}

func (t TodosListElement) SetList(tdl ui.List) TodosListElement {
	t.SetDataSetUI("todoslist", tdl)
	return t
}

func(t TodosListElement) UpdateList(tdl ui.List) TodosListElement{
	t.SyncUISetData("todoslist", tdl)
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



var newTodolistElement = doc.Elements.NewConstructor("todoslist", func(id string) *ui.Element {
	t := doc.Ul.WithID(id)
	doc.AddClass(t.AsElement(), "todo-list")

	tview := ui.NewViewElement(t.AsElement(), ui.NewView("all"), ui.NewView("active"), ui.NewView("completed"))
	t.OnRouterInit(func(r *ui.Router){
		names:= ui.NewList(ui.String("all"), ui.String("active"), ui.String("completed"))
		links:= ui.NewList(
			ui.String(r.NewLink("all").URI()),
			ui.String(r.NewLink("active").URI()),
			ui.String(r.NewLink("completed").URI()),
		)
		filterslist:=ui.NewObject()
		filterslist.Set("names",names)
		filterslist.Set("urls",links)

		t.AsElement().SetUI("filterslist",filterslist)
	})

	tview.AsElement().Watch("ui","filter", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		evt.Origin().TriggerEvent("renderlist")
		return false
	}))

	tview.AsElement().Watch("ui","todoslist", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
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
		ui.DEBUG("rendering...")
		t:= evt.Origin()

		t.RemoveChildren()
		
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
			newlist = ui.NewList()
		} else{
			newlist = todoslist.(ui.List)
		}

		newChildren := make([]*ui.Element, 0, len(newlist))

		for _, v := range newlist {
			o:= v.(Todo)
			if displayWhen(filter)(o){
				ntd, ok := FindTodoElement(doc.GetDocument(evt.Origin()),o)
				if !ok {
					ntd = TodosListElement{t}.NewTodo(o)
				}else{
					ui.DEBUG("todo existed already")
					ntd.SetDataSetUI("todo",o) // TODO use SyncUI
				}	
				newChildren = append(newChildren, ntd.AsElement())
			}
		}

		t.SetChildrenElements(newChildren...)
		
		return false
	}))

	return t.AsElement()
}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

func NewTodosListElement(id string, options ...string) TodosListElement {
	return TodosListElement{doc.LoadFromStorage(newTodolistElement(id, options...))}
}


func(t TodosListElement) NewTodo(o Todo) TodoElement{
	
	ntd := NewTodoElement(o)
	id, _ := o.Get("id")
	idstr := id.(ui.String)

	t.Watch("ui", "todo", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool { // escalate back to the todolist the data changes issued at the todo Element level
		var tdl ui.List
		res, ok := t.GetUI("todoslist")
		if !ok {
			tdl = ui.NewList()
		} else {
			tdl = res.(ui.List)
		}

		newval := evt.NewValue()

		filter,_:= t.GetUI("filter")

		for i, rawtodo := range tdl {
			todo := rawtodo.(Todo)
			oldid, _ := todo.Get("id")

			if oldid == idstr {
				if !ui.Equal(rawtodo, newval){
					tdl[i] = newval
					t.UpdateList(tdl)

					if !displayWhen(string(filter.(ui.String)))(newval.(Todo)){
						evt.Origin().Parent.RemoveChild(evt.Origin())
					}
				}				
				break
			}
		}
		return false
	}))

	t.WatchEvent( "delete", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		var tdl ui.List
		res, ok := t.AsElement().GetUI("todoslist")
		if !ok {
			tdl = ui.NewList()
		} else {
			tdl = res.(ui.List)
		}
		ntdl := tdl[:0]
		var i int
		
		for _, rawtodo := range tdl {
			todo := rawtodo.(Todo)
			oldid, _ := todo.Get("id")
			if oldid == idstr {
				t,ok:= FindTodoElement(doc.GetDocument(evt.Origin()),rawtodo.(Todo))
				if ok{
					ui.Delete(t.AsElement())
				}
				continue
			}
			ntdl= append(ntdl,rawtodo)
			i++
		}
		ntdl = ntdl[:i]
		t.UpdateList(ntdl)
		return false
	}))

	return ntd
}