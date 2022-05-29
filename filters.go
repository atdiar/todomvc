package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type Filters struct {
	ui.BasicElement
}

func NewFilter(name string, id string, u ui.Link) ui.BasicElement {
	li := doc.NewListItem(name, id)
	a := doc.NewAnchor(name, id+"-anchor")
	a.FromLink(u)
	li.AsElement().AppendChild(a)
	a.AsElement().Watch("ui", "active", a.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		b := evt.NewValue().(ui.Bool)
		if b {
			doc.AddClass(a.AsElement(), "selected")
		} else {
			doc.RemoveClass(a.AsElement(), "selected")
		}
		return false
	}))
	a.SetText(name)
	return li.BasicElement
}

func (f Filters) SetFilterList(filters ...ui.AnyElement) Filters {
	f.SetChildren(filters...)
	return f
}

func NewFilterList(name string, id string, options ...string) Filters {
	newFilters := doc.Elements.NewConstructor("filters", func(name string, id string) *ui.Element {
		u := doc.NewUl(name, id)
		doc.AddClass(u.AsElement(), "filters")
		return u.AsElement()
	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)
	return Filters{ui.BasicElement{doc.LoadFromStorage(newFilters(name, id, options...))}}
}

func ClearCompleteBtn(name string, id string) doc.Button {
	b := doc.NewButton(name, id, "button")
	b.SetText("Clear completed")
	doc.AddClass(b.AsElement(), "clear-completed")
	return b
}
