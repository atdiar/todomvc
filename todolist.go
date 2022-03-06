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

func (t TodosListElement) SetList(tdl ui.List) TodosListElement {
	t.AsElement().SetDataSetUI("todoslist", tdl)
	return t
}

func NewTodosListElement(name string, id string, options ...string) TodosListElement {
	newTodolistElement := doc.Elements.NewConstructor("todoslist", func(name string, id string) *ui.Element {
		t := doc.NewUl("todoslist", "todoslist")
		doc.AddClass(t.AsElement(), "todo-list")

		tview := ui.NewViewElement(t.AsElement(), ui.NewView("all"), ui.NewView("active"), ui.NewView("completed"))
		tview.OnActivation("all", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			tview.AsElement().SetUI("filter", ui.String("all"))
			// reload list
			res, ok := t.AsElement().Get("data", "todoslist")
			if ok {
				tdl := res.(ui.List)
				t.AsElement().SetDataSetUI("todoslist", tdl)
			}
			return false
		}))
		tview.OnActivation("active", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			tview.AsElement().SetUI("filter", ui.String("active"))
			// reload list
			res, ok := t.AsElement().Get("data", "todoslist")
			if ok {
				tdl := res.(ui.List)
				t.AsElement().SetDataSetUI("todoslist", tdl)
			}
			return false
		}))
		tview.OnActivation("completed", ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			tview.AsElement().SetUI("filter", ui.String("completed"))
			// reload list
			res, ok := t.AsElement().Get("data", "todoslist")
			if ok {
				tdl := res.(ui.List)
				t.AsElement().SetDataSetUI("todoslist", tdl)
			}
			return false
		}))

		t.AsElement().Watch("ui", "todoslist", t.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			// Handles list change, for instance, on new todo insertion
			t.AsElement().DeleteChildren()

			list := evt.NewValue().(ui.List) // TODO :  diff old list and new list
			snapshotlist := ui.NewList()
			snapshotlist = append(snapshotlist, list...)
			filter := "all"
			f, ok := t.AsElement().Get("ui", "filter")
			if ok {
				rf := f.(ui.String)
				filter = string(rf)
			}

			for _, v := range snapshotlist {
				// Let's get each todo
				o := v.(Todo)
				id, _ := o.Get("id")
				idstr := id.(ui.String)
				cplte, _ := o.Get("completed")
				complete := cplte.(ui.Bool)

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

				ntd := NewTodoElement(o).AsElement()
				t.AsElement().AppendChild(ntd)

				t.AsElement().Watch("data", "todo", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool { // escalate back to the todolist the data changes issued at the todo Element level
					var tdl ui.List
					res, ok := t.AsElement().Get("data", "todoslist")
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
							t.AsElement().SetDataSetUI("todoslist", append(tdl[:i], tdl[i+1:]...)) // update state and refresh list representation
							break
						}
						if oldid == idstr {
							tdl[i] = evt.NewValue()
							t.AsElement().SetDataSetUI("todoslist", tdl) // update state and refresh list representation
							break
						}
					}
					return false
				}))

				t.AsElement().Watch("event", "delete", ntd, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
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
							t.AsElement().SetDataSetUI("todoslist", tdl) // refresh list representation
							break
						}
					}
					return false
				}))

			}
			return false
		}))

		return t.AsElement()
	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

	return TodosListElement{ui.BasicElement{doc.LoadElement(newTodolistElement(name, id, options...))}}
}