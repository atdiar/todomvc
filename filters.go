package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

type Filters struct {
	*ui.Element
}

func NewFilter(name string, id string, u ui.Link) *ui.Element {
	li := doc.Li.WithID(id)
	a := doc.Anchor.WithID(id+"-anchor")
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

var newFilters = doc.Elements.NewConstructor("filters", func(id string) *ui.Element {
	e := doc.Ul.WithID(id).AsElement()
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

		ui.UseRouter(evt.Origin(),func(r *ui.Router){
			filters := make([]*ui.Element,0,len(urls))
			for i,url:= range urls{
				urlstr:= string(url.(ui.String))
				name := string(names[i].(ui.String))
				lnk,ok:= r.RetrieveLink(urlstr)
				if ok{
					filters = append(filters,NewFilter(name,name+"-filter",lnk).AsElement())
				}
			}
			evt.Origin().SetChildrenElements(filters...)
		})
		return false
	}))
	return e
}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

func NewFilterList(id string, options ...string) Filters {
	return Filters{doc.LoadFromStorage(newFilters( id, options...))}
}

func ClearCompleteBtn(id string) doc.ButtonElement {
	b := doc.Button.WithID(id, "button")
	b.SetText("Clear completed")
	doc.AddClass(b.AsElement(), "clear-completed")
	return b
}
