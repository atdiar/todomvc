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
	res, ok := t.AsElement().Get("data", "todoslist")
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

		t.AsElement().SetUI("filterslist",filterslist)
	})

	tview.AsElement().Watch("ui","filter", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		o:= ui.NewObject()
		o.Set("filter", evt.NewValue())
		tdl,ok:= evt.Origin().Get("ui","todoslist")
		if !ok{
			ui.DEBUG("could not find todoslist")
			o.Set("todoslist", ui.NewList())
		} else{
			o.Set("todoslist",tdl)
		}
		evt.Origin().SetUI("todoslistview",o)
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
		//ui.DEBUG("todolist ui prop being set")
		//ui.DEBUG(evt.NewValue())

		o.Set("todoslist", evt.NewValue())
		evt.Origin().Set("ui", "todoslistview",o)
		return false
	}))
	
	tview.OnActivation("all", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().SetUI("filter", ui.String("all"))
		doc.GetWindow().SetTitle("TODOMVC-all")
		return false
	}))
	tview.OnActivation("active", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().SetUI("filter", ui.String("active"))
		doc.GetWindow().SetTitle("TODOMVC-active")

		return false
	}))
	tview.OnActivation("completed", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().SetUI("filter", ui.String("completed"))
		doc.GetWindow().SetTitle("TODOMVC-completed")
		return false
	}))

	t.AsElement().Watch("ui", "todoslistview", t, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		ui.DEBUG("todolistview is being rendered", evt.Origin().Mounted(), evt.NewValue(),"\n old:",evt.OldValue())
		t:= evt.Origin()
		// Handles list change, for instance, on new todo insertion
		//t.RemoveChildren() // TODO delete detached elements


		o:= evt.NewValue().(ui.Object)
		
		filter:= string(o.MustGetString("filter"))
		list:= o.MustGetList("todoslist").Filter(displayWhen(filter))
		copylist := ui.Copy(list).(ui.List)


		newChildren := make([]*ui.Element, 0, len(list))
		childrenSet := make(map[string]struct{},len(list))

		for _, v := range list {
			// Let's get each todo
			o := v.(Todo)
			id, _ := o.Get("id")
			idstr := id.(ui.String)
			

			rntd, ok := FindTodoElement(o)
			ntd := rntd.AsElement()
			if ok {
				ntd.SyncUISetData("todo", o)
			}
			if !ok {
				ntd = NewTodoElement(o).AsElement()
				t.Watch("data", "todo", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool { // escalate back to the todolist the data changes issued at the todo Element level
					var tdl ui.List
					res, ok := t.Get("data", "todoslist")
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
							t.SetDataSetUI("todoslist", tdl) // DEBUG changed from SyncUISetData
							break
						}
					}
					return false
				}))

				t.Watch("event", "delete", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
					var tdl ui.List
					res, ok := t.AsElement().Get("data", "todoslist")
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
							evt.Origin().Parent.DeleteChild(evt.Origin())
							continue
						}
						ntdl[i] = rawtodo
						i++
					}
					ntdl = ntdl[:i]
					t.SetDataSetUI("todoslist", ntdl) // DEBUG changed from SyncUISetData
					return false
				}))
			}
			newChildren = append(newChildren, ntd)
			childrenSet[string(idstr)] =struct{}{}
		}
		
		/*for _,v:=range copylist{
			o := v.(Todo)
			id, _ := o.Get("id")
			idstr := id.(ui.String)
			if _,ok:= childrenSet[string(idstr)];!ok{
				d,ok:= FindTodoElement(o)
				if ok{
					ui.Delete(d.AsElement())
				}
			}
		}*/
		
		//ui.DEBUG(copylist)
		if len(newChildren) == 0 && len(copylist)!= 0{ 
			ui.DEBUG("filtered list ", len(list), list,)
			ui.DEBUG(filter, "no Element")
		}
		t.SetChildrenElements(newChildren...)
		return false
	}))


	return t.AsElement()
}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

func NewTodosListElement(id string, options ...string) TodosListElement {
	return TodosListElement{ui.BasicElement{doc.LoadFromStorage(newTodolistElement(id, options...))}}
}
