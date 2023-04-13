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
	t.AsElement().SetDataSetUI("todoslist", tdl)
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

	tview.AsElement().Watch("ui","filter", tview,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{

		o:= ui.NewObject()
		o.Set("filter", evt.NewValue())
		tdl,ok:= evt.Origin().Get("ui","todoslist")
		if !ok{
			o.Set("todoslist", ui.NewList())
		} else{
			o.Set("todoslist",tdl)
		}
		evt.Origin().TriggerEvent("renderlist",o)
		
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
		list:= evt.NewValue().(ui.List)
		for _,td:= range list{
			t:= td.(Todo)
			ntd, ok:= FindTodoElement(doc.GetDocument(evt.Origin()),t)
			if !ok{
				ntd = TodosListElement{evt.Origin()}.NewTodo(t)
			} else{
				ntd.SetDataSetUI("todo",t)
			}
		}

		o.Set("todoslist", evt.NewValue())
		evt.Origin().TriggerEvent("renderlist",o)
		
		return false
	}))
	
	tview.OnActivated("all", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().Set("ui","filter", ui.String("all"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-all")
		return false
	}))
	tview.OnActivated("active", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().Set("ui","filter", ui.String("active"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-active")

		return false
	}))
	tview.OnActivated("completed", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		evt.Origin().Set("ui","filter", ui.String("completed"))
		doc.GetDocument(evt.Origin()).Window().SetTitle("TODOMVC-completed")
		return false
	}))

	t.WatchEvent("renderlist", t, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
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


		newChildren := make([]*ui.Element, 0, len(newlist))
		childrenSet := make(map[string]struct{},len(newlist))

		for _, v := range newlist {
			// Let's get each todo
			o := v.(Todo)
			id, _ := o.Get("id")
			idstr := id.(ui.String)
			if displayWhen(filter)(o){
				
				ntd, ok := FindTodoElement(doc.GetDocument(evt.Origin()),o)
				if !ok {
					ntd = TodosListElement{t}.NewTodo(o)
				}	
				newChildren = append(newChildren, ntd.AsElement())
			}

			childrenSet[idstr.String()] = struct{}{}
		}

		t.SetChildrenElements(newChildren...)
		

		
		if true{
			// cleanup elements rcorresponding to deleted todos
			for _,v:=range oldlist{
				o := v.(Todo)
				id, _ := o.Get("id")
				idstr := id.(ui.String)
	
				if _,ok:= childrenSet[string(idstr)];!ok{
					d,ok:= FindTodoElement(doc.GetDocument(evt.Origin()),o)
					if ok{
						ui.Delete(d.AsElement())
					}
				}
			}
		}
		
		
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
		res, ok := t.Get("ui", "todoslist")
		if !ok {
			tdl = ui.NewList()
		} else {
			tdl = res.(ui.List)
		}

		newval := evt.NewValue()

		for i, rawtodo := range tdl {
			todo := rawtodo.(Todo)
			oldid, _ := todo.Get("id")
			title, _ := todo.Get("title")
			titlestr := title.(ui.String)
			if len(titlestr) == 0 {
				// t.AsElement().SetDataSetUI("todoslist", append(tdl[:i], tdl[i+1:]...)) // update state and refresh list representation
				ntd.TriggerEvent("delete", ui.Bool(true))
				break
			}
			if oldid == idstr {
				if !ui.Equal(tdl[i], newval){
					tdl[i] = newval
					t.SetList(tdl)
				}				
				break
			}
		}
		return false
	}))

	t.WatchEvent( "delete", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
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
		t.SetList(ntdl) // TODO updatedlist event (with specfic puprose value so that the corresponding DOM element gets removed)
		return false
	}))

	return ntd
}