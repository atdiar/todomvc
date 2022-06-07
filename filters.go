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

func NewFilterList(name string, id string, options ...string) Filters {
	newFilters := doc.Elements.NewConstructor("filters", func(name string, id string) *ui.Element {
		u := doc.NewUl(name, id)
		doc.AddClass(u.AsElement(), "filters")
		return u.AsElement()
	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

	e:= newFilters(name, id, options...)

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

		ui.UseRouter(e,func(r *ui.Router){
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
	return Filters{ui.BasicElement{doc.LoadFromStorage(e)}}
}

func ClearCompleteBtn(name string, id string) doc.Button {
	b := doc.NewButton(name, id, "button")
	b.SetText("Clear completed")
	doc.AddClass(b.AsElement(), "clear-completed")
	return b
}
