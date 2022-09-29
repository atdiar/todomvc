package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type TodosListElement struct {
	ui.BasicElement
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

func TodoListFromRef(ref *ui.Element) TodosListElement{
	return TodosListElement{ui.BasicElement{ref}}
}

func (t TodosListElement) SetList(tdl ui.List) TodosListElement {
	t.AsElement().SetDataSetUI("todoslist", tdl)
	return t
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
	t := doc.Ul(id)
	doc.AddClass(t.AsElement(), "todo-list")

	tview := ui.NewViewElement(t.AsElement(), ui.NewView("all"), ui.NewView("active"), ui.NewView("completed"))
	ui.UseRouter(t.AsElement(),func(r *ui.Router){
		names:= ui.NewList(ui.String("all"), ui.String("active"), ui.String("completed"))
		links:= ui.NewList(
			ui.String(r.NewLink("all").URI()),
			ui.String(r.NewLink("active").URI()),
			ui.String(r.NewLink("completed").URI()),
		)
		filterslist:=ui.NewObject()
		filterslist.Set("names",names)
		filterslist.Set("urls",links)

		t.AsElement().SetDataSetUI("filterslist",filterslist)
	})

	tview.AsElement().Watch("event","filter", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		evt.Origin().SetDataSetUI("filter",evt.NewValue())

		o:= ui.NewObject()
		o.Set("filter", evt.NewValue())
		tdl,ok:= evt.Origin().Get("data","todoslist")
		if !ok{
			o.Set("todoslist", ui.NewList())
		} else{
			o.Set("todoslist",tdl.(ui.List))
		}
		evt.Origin().Set("event","renderlist",o)
		return false
	}))

	tview.AsElement().Watch("ui","todoslist", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		o:= ui.NewObject()
		var filter = "all"
		f,ok:= evt.Origin().Get("ui","filter")
		if !ok{
			o.Set("filter", ui.String(filter))
		} else{
			filter = string(f.(ui.String))
			o.Set("filter",f)
		}

		o.Set("todoslist", evt.NewValue())
		evt.Origin().Set("event", "renderlist",o)
		return false
	}))
	
	tview.OnActivated("all", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().Set("event","filter", ui.String("all"))
		doc.GetWindow().SetTitle("TODOMVC-all")
		return false
	}))
	tview.OnActivated("active", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().Set("event","filter", ui.String("active"))
		doc.GetWindow().SetTitle("TODOMVC-active")

		return false
	}))
	tview.OnActivated("completed", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().Set("event","filter", ui.String("completed"))
		doc.GetWindow().SetTitle("TODOMVC-completed")
		return false
	}))

	t.AsElement().Watch("event", "renderlist", t, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		t:= evt.Origin()
		t.RemoveChildren()
		// We retrieve old list so that the elements that were removed can be definitively deleted
		var oldlist ui.List
		oo,ok:= evt.OldValue().(ui.Object)
		if ok{
			oldlist = oo.MustGetList("todoslist")
		}


		o:= evt.NewValue().(ui.Object)
		
		filter:= string(o.MustGetString("filter"))
		newlist:= o.MustGetList("todoslist")
		list:= newlist.Filter(displayWhen(filter))


		newChildren := make([]*ui.Element, 0, len(list))
		childrenSet := make(map[string]struct{},len(newlist))

		for _, v := range list {
			// Let's get each todo
			o := v.(Todo)
			id, _ := o.Get("id")
			idstr := id.(ui.String)

			rntd, ok := FindTodoElement(o)
			ntd := rntd.AsElement()
			if ok {
				ntd.SyncUISetData("todo", o)
				ntd.Set("event","update",ui.Bool(true))
			}
			if !ok {
				ntd = NewTodoElement(o).AsElement()
				/*ntd.OnUnmounted(ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
					ui.DEBUG(evt.Origin().ID + " todo element has been unmounted")
					return false
				})) */
				t.Watch("ui", "todo", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool { // escalate back to the todolist the data changes issued at the todo Element level
					var tdl ui.List
					res, ok := t.Get("ui", "todoslist")
					if !ok {
						tdl = ui.NewList()
					} else {
						tdl = res.(ui.List)
					}

					for i, rawtodo := range tdl {
						todo := rawtodo.(Todo)
						oldid, _ := todo.Get("id")
						title, _ := todo.Get("title")
						titlestr := title.(ui.String)
						if len(titlestr) == 0 {
							// t.AsElement().SetDataSetUI("todoslist", append(tdl[:i], tdl[i+1:]...)) // update state and refresh list representation
							ntd.Set("event", "delete", ui.Bool(true))
							break
						}
						if oldid == idstr {
							tdl[i] = evt.NewValue()
							t.SetDataSetUI("todoslist", tdl)
							break
						}
					}
					return false
				}))

				t.Watch("event", "delete", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
					var tdl ui.List
					res, ok := t.AsElement().Get("ui", "todoslist")
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
							continue
						}
						ntdl= append(ntdl,rawtodo)
						i++
					}
					ntdl = ntdl[:i]
					t.SetDataSetUI("todoslist", ntdl)
					return false
				}))
			}
			newChildren = append(newChildren, ntd)
		}
		
		t.SetChildrenElements(newChildren...)

		for _,v:= range newlist{
			o := v.(Todo)
			id, _ := o.Get("id")
			idstr := id.(ui.String)
			childrenSet[string(idstr)]= struct{}{}
		}
		
		for _,v:=range oldlist{
			o := v.(Todo)
			id, _ := o.Get("id")
			idstr := id.(ui.String)
			if _,ok:= childrenSet[string(idstr)];!ok{
				d,ok:= FindTodoElement(o)
				if ok{
					ui.Delete(d.AsElement())
				}
			}
		}

		
		
		return false
	}))


	return t.AsElement()
}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

func NewTodosListElement(id string, options ...string) TodosListElement {
	return TodosListElement{ui.BasicElement{doc.LoadFromStorage(newTodolistElement(id, options...))}}
}
