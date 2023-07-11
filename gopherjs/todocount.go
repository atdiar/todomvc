package main

import (
	"strconv"
	"strings"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type TodoCount struct {
	*ui.Element
}

func (t TodoCount) SetCount(count int) TodoCount {
	t.AsElement().SetDataSetUI("count", ui.Number(count))
	return t
}

func TodoCountFromRef(ref *ui.Element) TodoCount{
	return TodoCount{ref}
}

func NewTodoCount(d doc.Document, id string, options ...string) TodoCount {
	return TodoCount{newtodocount(d, id, options...)}
}

func newtodocount (document doc.Document, id string, options ...string) *ui.Element {
	s := document.Span.WithID(id,options...)
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
		htmlstr := strings.Join([]string{"<strong>",strconv.Itoa(nn),"<strong>"," ",i," left"},"")
		doc.SetInnerHTML(s.AsElement(), htmlstr)
		return false
	}).RunASAP())

	doc.AddClass(s.AsElement(), "todo-count")
	return s.AsElement()
}
