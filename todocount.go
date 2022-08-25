package main

import (
	"strconv"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type TodoCount struct {
	ui.BasicElement
}

func (t TodoCount) SetCount(count int) TodoCount {
	t.AsElement().SetDataSetUI("count", ui.Number(count))
	return t
}

func TodoCountFromRef(ref *ui.Element) TodoCount{
	return TodoCount{ui.BasicElement{ref}}
}

func NewTodoCount(id string, options ...string) TodoCount {
	newtodocount := doc.Elements.NewConstructor("todocount", func(id string) *ui.Element {
		s := doc.NewSpan(id)
		s.AsElement().Watch("ui", "count", s.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			n, ok := evt.NewValue().(ui.Number)
			if !ok {
				return true
			}
			nn := int(n)
			i := "item"
			if nn > 1 {
				i = "items"
			}
			htmlstr := "<strong>" + strconv.Itoa(nn) + "<strong>" + " " + i + " left"
			doc.SetInnerHTML(s.AsElement(), htmlstr)
			return false
		}))

		doc.AddClass(s.AsElement(), "todo-count")
		return s.AsElement()
	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)
	return TodoCount{ui.BasicElement{doc.LoadFromStorage(newtodocount(id, options...))}}
}
