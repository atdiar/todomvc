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

func NewTodosListElement(id string, options ...string) TodosListElement {
	newTodolistElement := doc.Elements.NewConstructor("todoslist", func(id string) *ui.Element {
		t := doc.NewUl(id)
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
		
		tview.OnActivation("all", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			evt.Origin().SetUI("filter", ui.String("all"))
			doc.GetWindow().SetTitle("TODOMVC-all")
			// reload list
			evt.Origin().RemoveChildren()
			res, ok := evt.Origin().Get("data", "todoslist")
			if ok {
				tdl := res.(ui.List)
				evt.Origin().SetDataSetUI("todoslist", tdl)
			}
			return false
		}))
		tview.OnActivation("active", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			evt.Origin().SetUI("filter", ui.String("active"))
			doc.GetWindow().SetTitle("TODOMVC-active")
			// reload list
			evt.Origin().RemoveChildren()
			res, ok := evt.Origin().Get("data", "todoslist")
			if ok {
				tdl := res.(ui.List)
				evt.Origin().SetDataSetUI("todoslist", tdl)
			}
			return false
		}))
		tview.OnActivation("completed", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			evt.Origin().SetUI("filter", ui.String("completed"))
			doc.GetWindow().SetTitle("TODOMVC-completed")
			// reload list
			evt.Origin().RemoveChildren()
			res, ok := evt.Origin().Get("data", "todoslist")
			if ok {
				tdl := res.(ui.List)
				evt.Origin().SetDataSetUI("todoslist", tdl)
			}
			return false
		}))

		t.AsElement().Watch("ui", "todoslist", t.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			t:= evt.Origin()
			// Handles list change, for instance, on new todo insertion
			t.RemoveChildren() // TODO delete detached elements


			list := evt.NewValue().(ui.List)
			//snapshotlist := ui.NewList()
			//snapshotlist = append(snapshotlist, list...)
			filter := "all"
			f, ok := t.Get("ui", "filter")
			if ok {
				rf := f.(ui.String)
				filter = string(rf)
			}

			newChildren := make([]*ui.Element, 0, len(list))
			childrenSet := make(map[ui.String]struct{},len(list))

			for _, v := range list {
				// Let's get each todo
				o := v.(Todo)
				id, _ := o.Get("id")
				idstr := id.(ui.String)
				cplte, _ := o.Get("completed")
				complete := cplte.(ui.Bool)
				childrenSet[idstr]=struct{}{}

				if filter == "active" {
					if complete {
						continue
					}
				}

				if filter == "completed" {
					if !complete {
						continue
					}
				}

				rntd, ok := FindTodoElement(o)
				ntd := rntd.AsElement()
				if ok {
					ntd.SetUI("todo", o)
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
								t.SyncUISetData("todoslist", tdl) // update state and refresh list representation TODO use Update method
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
						snapshottdl := ui.NewList()
						snapshottdl = append(snapshottdl, tdl...)
						for i, rawtodo := range snapshottdl {
							todo := rawtodo.(Todo)
							oldid, _ := todo.Get("id")
							if oldid == idstr {
								tdl = append(tdl[:i], tdl[i+1:]...)
								//t.AsElement().SetDataSetUI("todoslist", tdl) // refresh list representation
								ntd.Parent.DeleteChild(ntd)
								t.SyncUISetData("todoslist", tdl)
								break
							}
						}
						return false
					}))
				}
				newChildren = append(newChildren, ntd)
			}
			oldlist,ok:=evt.OldValue().(ui.List)
			if ok{
				for _,v:=range oldlist{
					o := v.(Todo)
					id, _ := o.Get("id")
					idstr := id.(ui.String)
					if _,ok:= childrenSet[idstr];!ok{
						d,ok:= FindTodoElement(o)
						if ok{
							ui.Delete(d.AsElement())
						}
					}
				}
			}
			t.SetChildrenElements(newChildren...)
			return false
		}))

		return t.AsElement()
	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

	return TodosListElement{ui.BasicElement{doc.LoadFromStorage(newTodolistElement(id, options...))}}
}
