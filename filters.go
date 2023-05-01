package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type Filters struct {
	*ui.Element
}

func NewFilter(document doc.Document, name string, id string, u ui.Link, options ...string) *ui.Element {
	li := document.Li.WithID(id, options...)
	a := document.Anchor.WithID(id+"-anchor")
	a.FromLink(u)
	li.AsElement().AppendChild(a)
	a.AsElement().Watch("ui", "active", a, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		b := evt.NewValue().(ui.Bool)
		if b {
			doc.AddClass(evt.Origin(), "selected")
		} else {
			doc.RemoveClass(evt.Origin(), "selected")
		}
		return false
	}))
	a.SetText(name)
	return li.AsElement()
}

func newFilters(document doc.Document, id string, options ...string) *ui.Element {
	e := document.Ul.WithID(id, options...).AsElement()
	doc.AddClass(e, "filters")

	e.Watch("ui","filterslist",e, ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		l:= evt.NewValue().(ui.Object)

		nameslist,ok := l.Get("names")
		if !ok{
			panic("Bad format for filter list object.")
		}
		names := nameslist.(ui.List)

		urllist,ok:= l.Get("urls")
		if !ok{
			panic("Bad format for filter list object")
		}
		urls := urllist.(ui.List)

		evt.Origin().OnRouterInit(func(r *ui.Router){
			filters := make([]*ui.Element,0,len(urls))
			for i,url:= range urls{
				urlstr:= string(url.(ui.String))
				name := string(names[i].(ui.String))
				lnk,ok:= r.RetrieveLink(urlstr)
				if ok{
					filters = append(filters,NewFilter(document, name,name+"-filter",lnk).AsElement())
				}
			}
			evt.Origin().SetChildrenElements(filters...)
		})
		return false
	}))
	return e
}

func NewFilterList(document doc.Document, id string, options ...string) Filters {
	return Filters{doc.LoadFromStorage(newFilters( document, id, options...))}
}

func ClearCompleteBtn(id string) doc.ButtonElement {
	b := doc.Button.WithID(id, "button")
	b.SetText("Clear completed")
	doc.AddClass(b.AsElement(), "clear-completed")
	return b
}
