package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

// GOOS=js GOARCH=wasm go build -o  ../../app.wasm
var DEBUG = log.Print

type Todo = ui.Object

func NewTodo(title ui.String) Todo {
	NewID := ui.NewIDgenerator(time.Now().UnixNano())
	o := ui.NewObject()
	o.Set("id", ui.String(NewID()))
	o.Set("completed", ui.Bool(false))
	o.Set("title", title)
	return o
}

type TodoElement struct {
	ui.BasicElement
}

func (t TodoElement) SetComplete(b bool) TodoElement {
	res, ok := t.AsElement().Get("data", "todo")
	if !ok {
		return t
	}
	todo, ok := res.(ui.Object)
	if !ok {
		return t
	}

	todo.Set("completed", ui.Bool(b))
	t.AsElement().SetDataSetUI("todo", todo)

	return t
}

func NewTodoElement(t Todo) TodoElement {
	todoid, ok := t.Get("id")
	if !ok {
		return TodoElement{}
	}
	todoidstr, ok := todoid.(ui.String)
	if !ok {
		return TodoElement{}
	}

	newtodo := doc.Elements.NewConstructor("todo", func(name string, id string) *ui.Element {
		d := doc.NewDiv(name, id)
		doc.AddClass(d.AsElement(), "view")

		i := doc.NewInput("checkbox", "completed", id+"-completed")
		doc.AddClass(i.AsElement(), "toggle")

		l := doc.NewLabel(id, id+"-lbl")

		b := doc.NewButton(id, id+"-btn", "button")
		doc.AddClass(b.AsElement(), "destroy")

		d.SetChildren(i, l, b)
		li := doc.NewListItem("li-"+id, "li-"+id).SetValue(d.AsElement())

		li.AsElement().Watch("ui", "todo", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			t, ok := evt.NewValue().(ui.Object)
			if !ok {
				return true
			}

			_, ok = t.Get("id")
			if !ok {
				return true
			}

			todocomplete, ok := t.Get("completed")
			if !ok {
				return true
			}
			todocompletebool := todocomplete.(ui.Bool)

			if todocompletebool {
				doc.AddClass(li.AsElement(), "complete")
			} else {
				doc.RemoveClass(li.AsElement(), "complete")
			}

			todotitle, ok := t.Get("title")
			if !ok {
				return true
			}
			todotitlestr := todotitle.(ui.String)

			i.AsElement().SetUI("checked", todocompletebool)
			l.SetText(string(todotitlestr))

			return false
		}))

		li.AsElement().Watch("event", "toggle", li, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			res, ok := li.AsElement().GetData("todo")
			if !ok {
				return true
			}
			todo := res.(ui.Object)

			b, ok := todo.Get("completed")
			if !ok {
				return true
			}
			complete := !(b.(ui.Bool))

			todo.Set("completed", ui.Bool(complete))

			li.AsElement().SetDataSetUI("todo", todo)
			return false
		}))

		i.AsElement().AddEventListener("click", ui.NewEventHandler(func(evt ui.Event) bool {
			li.AsElement().Set("event", "toggle", ui.Bool(true))
			return false
		}), doc.NativeEventBridge)

		b.AsElement().AddEventListener("click", ui.NewEventHandler(func(evt ui.Event) bool {
			li.AsElement().Set("event", "delete", ui.Bool(true))
			DEBUG("click", evt.Target().ID)
			return false
		}), doc.NativeEventBridge)

		return li.AsElement()

	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)

	ntd := doc.LoadElement(newtodo("todo-"+string(todoidstr), "todo-"+string(todoidstr)))
	ntd.SetDataSetUI("todo", t)

	return TodoElement{ui.BasicElement{ntd}}
}

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

		t.AsElement().Watch("ui", "todoslist", t.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
			// Handles list change, for instance, on new todo insertion
			t.AsElement().DeleteChildren()
			list := evt.NewValue().(ui.List) // TODO :  diff old list and new list
			snapshotlist:=ui.NewList()
			snapshotlist= append(snapshotlist,list...)

			for _, v := range snapshotlist {
				// Let's get each todo
				o := v.(Todo)
				id, ok := o.Get("id")
				if !ok {panic("Unexpected todo format")}
				idstr := id.(ui.String)

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
					snapshottdl:= ui.NewList()
					snapshottdl= append(snapshottdl,tdl...)
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

type TodoCount struct{
	ui.BasicElement
}

func(t TodoCount) SetCount(count int) TodoCount{
	t.AsElement().SetDataSetUI("count",ui.Number(count))
	return t
}

func NewTodoCount(name string,id string, options ...string) TodoCount{
	newtodocount := doc.Elements.NewConstructor("todocount", func(name string, id string)*ui.Element{
		s:= doc.NewSpan(name,id)
		s.AsElement().Watch("ui","count",s.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
				n,ok:= evt.NewValue().(ui.Number)
				if !ok{
					return true
				}
				nn:= int(n)
				i:="item"
				if nn>1{
					i="items"
				}
				htmlstr:= "<strong>" + strconv.Itoa(nn) + "<strong>" + " " + i +" left"
				doc.SetInnerHTML(s.AsElement(),htmlstr)
				return false
		}))

		doc.AddClass(s.AsElement(),"todo-count")
		return s.AsElement()
	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)
	return TodoCount{ui.BasicElement{doc.LoadElement(newtodocount(name,id,options...))}}
}

type Filters struct{
	ui.BasicElement
}

func NewFilter(name string,id string, u ui.Link) ui.BasicElement{
	li:=doc.NewListItem(name,id)
	a:= doc.NewAnchor(name,id+"-anchor")
	a.FromLink(u)
	li.AsElement().AppendChild(a)
	a.AsElement().Watch("ui","active",a.AsElement(),ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		b := evt.NewValue().(ui.Bool)
		if b{
			doc.AddClass(a.AsElement(),"selected")
		} else{
			doc.RemoveClass(a.AsElement(),"selected")
		}
		return false
	}))
	a.SetText(name)
	return li.BasicElement
}

func(f Filters) SetFilterList(filters ...ui.AnyElement) Filters{
	f.SetChildren(filters...)
	return f
}

func NewFilterList(name string, id string, options ...string) Filters{
	newFilters:= doc.Elements.NewConstructor("filters",func(name string,id string)*ui.Element{
		u:= doc.NewUl(name,id)
		doc.AddClass(u.AsElement(),"filters")
		return u.AsElement()
	}, doc.AllowSessionStoragePersistence, doc.AllowAppLocalStoragePersistence)
	return Filters{ui.BasicElement{doc.LoadElement(newFilters(name,id,options...))}}
}

func ClearCompleteBtn(name string,id string) doc.Button{
	b:= doc.NewButton(name,id,"button")
	b.SetText("Clear completed")
	doc.AddClass(b.AsElement(),"clear-completed")
	return b
}

func NewTodoInput(name string, id string) doc.Input {
	todosinput := doc.NewInput("text", name, id)
	doc.SetAttribute(todosinput.AsElement(), "placeholder", "What needs to be done?")
	doc.SetAttribute(todosinput.AsElement(), "autofocus", "")
	doc.SetAttribute(todosinput.AsElement(), "onfocus", "this.value=''")

	todosinput.AsElement().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
		s := ui.String(evt.Value())
		str := strings.TrimSpace(string(s)) // Trim value
		todosinput.AsElement().SetDataSetUI("value", ui.String(str))
		return false
	}), doc.NativeEventBridge)

	todosinput.AsElement().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
		if evt.Value() == "Enter" {
			evt.PreventDefault()
			if todosinput.Value() != "" {
				todosinput.AsElement().Set("event", "newtodo", todosinput.Value())
			}
			todosinput.AsElement().SetDataSetUI("value", ui.String(""))
			todosinput.Clear()
		}
		return false
	}), doc.NativeEventBridge)

	return todosinput
}

func main() {
	// 1. Create a new document
	Document := doc.NewDocument("Todo-App")
	AppSection := doc.NewSection("todoapp", "todoapp")
	AppFooter := doc.NewFooter("infofooter", "infofooter")

	Document.SetChildren(AppSection, AppFooter)

	// 2. Build AppSection
	MainHeader := doc.NewHeader("header", "header")
	MainSection := doc.NewSection("main", "main")
	MainFooter := doc.NewFooter("footer", "footer")

	AppSection.SetChildren(MainHeader, MainSection, MainFooter)

	// 3. Build MainHeader
	MainHeading := doc.NewH1("todo", "apptitle").SetText("Todo")
	todosinput := NewTodoInput("todo", "new-todo")

	MainHeader.SetChildren(MainHeading, todosinput)

	// 4. Build MainSection
	ToggleAllInput := doc.NewInput("checkbox", "toogle-all", "toggle-all")
	ToggleAllInput.AsElement().AddEventListener("click",ui.NewEventHandler(func(evt ui.Event)bool{
		togglestate,ok := ToggleAllInput.AsElement().GetData("checked")
		if !ok{
			ToggleAllInput.AsElement().Set("event","toggled",ui.Bool(true))
			return false
		}
		ts := togglestate.(ui.Bool)
		ToggleAllInput.AsElement().Set("event","toggled",!ts)
		return false
	}),doc.NativeEventBridge)

	ToggleLabel := doc.NewLabel("toggle-all-Label", "toggle-all-label").For(ToggleAllInput.AsElement())
	TodosList := NewTodosListElement("todo-list", "todo-list", doc.EnableSessionPersistence())
	todolistview:= ui.NewViewElement(TodosList.AsElement(),ui.NewView("all"),ui.NewView("active"),ui.NewView("completed"))
	MainSection.SetChildren(ToggleAllInput, ToggleLabel, TodosList)

	// 5. Build MainFooter
	TodoCount := NewTodoCount("todo-count","todo-count")

	FilterList := NewFilterList("filters","filters")
	// links
	router:= ui.NewRouter("/",todolistview)
	linkall:= router.NewLink(todolistview,"all")
	linkactive:= router.NewLink(todolistview,"active")
	linkcompleted:= router.NewLink(todolistview,"completed")

	allFilter:= NewFilter("All","all-filter",linkall) // TODO Set LInk
	activeFilter:= NewFilter("Active","active-filter", linkactive) // TODO
	completedFilter:= NewFilter("Completed","completed-filter",linkcompleted) // TODO
	FilterList.SetFilterList(allFilter,activeFilter,completedFilter)

	ClearCompleteButton:= ClearCompleteBtn("clear-complete","clear-complete")
	ClearCompleteButton.AsElement().AddEventListener("click",ui.NewEventHandler(func(evt ui.Event)bool{
		ClearCompleteButton.AsElement().Set("event","clear",ui.Bool(true))
		return false
	}),doc.NativeEventBridge)

	MainFooter.SetChildren(TodoCount,FilterList,ClearCompleteButton)

	// x.Build AppFooter

	//css
	doc.AddClass(AppSection.AsElement(), "todoapp")
	doc.AddClass(AppFooter.AsElement(), "info")
	doc.AddClass(MainHeader.AsElement(), "header")
	doc.AddClass(MainSection.AsElement(), "main")
	doc.AddClass(MainFooter.AsElement(), "footer")
	doc.AddClass(todosinput.AsElement(), "new-todo")
	doc.AddClass(ToggleAllInput.AsElement(), "toggle-all")
	//doc.AddClass(TodosList.AsElement(),"todo-list")

	// 4. Watch for new todos to insert
	AppSection.AsElement().Watch("event", "newtodo", todosinput.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tdl := TodosList.GetList()

		s, ok := evt.NewValue().(ui.String)
		if !ok || s == "" {
			return true
		}
		t := NewTodo(s)
		tdl = append(tdl, t)

		TodosList.AsElement().SetDataSetUI("todoslist", tdl)

		return false
	}))

	AppSection.AsElement().Watch("event", "clear", ClearCompleteButton.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		TodosList.AsElement().SetDataSetUI("todoslist", ui.NewList())
		return false
	}))

	AppSection.AsElement().Watch("data", "todoslist", TodosList.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		l:= evt.NewValue().(ui.List) // we know it's a list, otherwise it can just panic, it's ok

		if len(l) == 0{
			doc.SetInlineCSS(MainFooter.AsElement(),"display:none")
		} else{
			doc.SetInlineCSS(MainFooter.AsElement(),"display:block")
		}

		countcomplete := 0
		allcomplete:= true
		for _,todo:=range l{
			t:= todo.(Todo)
			completed,ok:= t.Get("completed")
			if !ok{
				allcomplete = false
			}
			c:=completed.(ui.Bool)
			if !c{
				allcomplete = false
			} else{
				countcomplete++
			}
		}

		TodoCount.SetCount(countcomplete)

		if countcomplete ==0{
			doc.SetInlineCSS(ClearCompleteButton.AsElement(),"display:none")
		} else{
			doc.SetInlineCSS(ClearCompleteButton.AsElement(),"display:block")
		}

		if allcomplete{
				ToggleAllInput.AsElement().SetDataSetUI("checked",ui.Bool(true))
		} else{
			ToggleAllInput.AsElement().SetDataSetUI("checked",ui.Bool(false))
		}
		return false
	}))

	AppSection.AsElement().Watch("event","toggled",ToggleAllInput.AsElement(),ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		status := evt.NewValue().(ui.Bool)
		tdl := TodosList.GetList()
		for i,todo:=range tdl{
			t:= todo.(Todo)
			t.Set("completed",status)
			tdl[i] = t
		}
		TodosList.SetList(tdl)
		return false
	}))

	router.ListenAndServe("popstate",Document.AsElement(),doc.NativeEventBridge)

	c := make(chan struct{}, 0)
	<-c
}

/*
func main() {
	// 1. Create a new document
	root2 := doc.NewDocument("TestAppID2")

	// 2. Create an Input box that will  allow to create new todos
	todosinput := doc.NewInput("text", "todo", "newtodo", doc.EnableSessionPersistence())
	doc.SetAttribute(todosinput.Element(), "placeholder", "What needs to be done?")
	doc.SetAttribute(todosinput.Element(), "autofocus", "")
	doc.SetAttribute(todosinput.Element(), "onfocus", "this.value=''")
	root2.Element().AppendChild(todosinput.Element())

	// 3. TODO definition
	type Todo = ui.Object
	NewTodo := func(title ui.String) Todo {
		o := ui.NewObject()
		o.Set("id", ui.String(doc.NewID()))
		o.Set("completed", ui.Bool(false))
		o.Set("title", ui.String(title))
		return o
	}

	// 4. List
	l := doc.NewUl("todoslist", "todoslist", doc.EnableSessionPersistence())
	root2.Element().AppendChild(l.Element())

	// 5. Watch for new todos to insert
	root2.Element().Watch("data", "newtodo", todosinput.Element(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		var tdl ui.List
		todoslist, ok := l.Element().Get("data", "todoslist")
		if !ok {
			tdl = ui.NewList()
		}
		tdl, ok = todoslist.(ui.List)
		if !ok {
			tdl = ui.NewList()
		}

		s, ok := evt.NewValue().(ui.String)
		if !ok || s == "" {
			return true
		}
		if s != "" {
			t := NewTodo(s)
			tdl = append(tdl, t)
			log.Print(tdl)
			l.Element().SetData("todoslist", tdl) // todo SetData only completed require another step
		}

		// Store todo raw for display
		v:= l.Values()
		v= append(v,evt.NewValue())
		l.FromValues(v...)

		return false
	}))

	// UI event handlers
	todosinput.Element().AddEventListener("change", ui.NewEventHandler(func(evt ui.Event) bool {
		s := ui.String(evt.Value())
		todosinput.Element().SetDataSetUI("value", s)
		return false
	}), doc.NativeEventBridge)

	todosinput.Element().AddEventListener("keyup", ui.NewEventHandler(func(evt ui.Event) bool {
		if evt.Value() == "Enter" {
			evt.PreventDefault()
			if todosinput.Value() != ""{
				todosinput.Element().SetDataSetUI("newtodo", todosinput.Value())
			}
			todosinput.Element().SetDataSetUI("value", ui.String(""))
			todosinput.Blur()
		}
		return false
	}), doc.NativeEventBridge)

	c := make(chan struct{}, 0)
	<-c
}
*/

/*func main() {
	// 1. Main Element
	root2 := doc.NewDocument("test2", "TestAppID2")

	d:= doc.NewDiv("todoses","todocontainer",doc.EnableSessionPersistence())

	root2.Element().Watch("data","mutation",d.Element(),ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		log.Print("TEST !!!!!", evt.NewValue())
		return false
	}))

	d.Element().SetData("mutation",ui.String(root2.Element().ID))

	c := make(chan struct{}, 0)
	<-c
}
*/

/*
func main() {

	root := doc.NewDocument("test3", "TestAppID")

	rd := doc.NewDiv("test", "rootview") //.SetText("This is the view at initialization...")

	rd1 := doc.NewDiv("test", "d1").SetText("x top View A")
	rd2 := doc.NewDiv("test", "d2").SetText("x top View B")
	view1 := ui.NewView("view1", rd1.Element())
	view2 := ui.NewView("view2", rd2.Element())
	v := ui.NewViewElement(rd.Element(), view1, view2)

	rd3 := doc.NewDiv("test", "d3")
	rd4 := doc.NewDiv("test", "d4").SetText("xxxx    nested viewA")
	rd5 := doc.NewDiv("test", "d5").SetText("xxxx    nested viewB")
	rd6 := doc.NewDiv("test", "d6").SetText("xxxx    nested viewC")
	view4 := ui.NewView("nested1", rd4.Element())
	view5 := ui.NewView("nested2", rd5.Element())
	view6 := ui.NewView("nested3", rd6.Element())
	v2 := ui.NewViewElement(rd3.Element(), view4, view5, view6)

	rd2.Element().AppendChild(v2.Element())

	// By construction v2 is nested in v

	root.Element().AppendChild(rd.Element())

	router := ui.NewRouter("/", v)
	nd := doc.NewDiv("notfound", "divnotfound").SetText("notfound")
	router.OnNotfound(ui.NewView("notfound", nd.Element()))

	n := 0

	eh := ui.NewEventHandler(func(evt ui.Event) bool {
		n++
		//router.GoTo("/test"+ strconv.Itoa(n%3+1)+"/nested" + strconv.Itoa(n%2+1))
		router.GoTo("/view2/d3/nested" + strconv.Itoa(n%3+1))
		return false
	})

	root.Element().AddEventListener("click", eh, doc.NativeEventBridge)

	router.ListenAndServe("popstate", doc.GetWindow().Element(), doc.NativeEventBridge)

	c := make(chan struct{}, 0)
	<-c
}

*/

/*

//=========================== Event+ View + Routing test =======================

func main() {

	root := doc.NewDocument("test", "TestAppID")

	rd:= doc.NewDiv("test","divtest") //.SetText("This is the view at initialization...")

	rd2:= doc.NewDiv("test","divtest2").SetText("this is but a test nB02...")
	rd3:= doc.NewDiv("test","divtes3t").SetText("this is but a test nB03...")

	view2:=ui.NewView("test2",rd2.Element())
	view3:= ui.NewView("test3",rd3.Element())
	v:= ui.NewViewElement(rd.Element(),view2).AddView(view3)

	root.Element().AppendChild(v.Element())

	router := ui.NewRouter("/",v)
	log.Println("v route: ",v.Element().Route())
	nd:= doc.NewDiv("notfound","divnotfound").SetText("notfound")
	router.OnNotfound(ui.NewView("notfound",nd.Element()))

	n:=0

	eh:= ui.NewEventHandler(func(evt ui.Event)bool{
		n++
		router.GoTo("/test"+ strconv.Itoa(n%3+1))
		log.Print("click")
		return false
	})

	root.Element().AddEventListener("click",eh,doc.NativeEventBridge)

	router.ListenAndServe("popstate",doc.GetWindow().Element(),doc.NativeEventBridge)

	c := make(chan struct{}, 0)
	<-c
}

*/

/*
=========================== Mutation test ======================================


func main() {

	root := doc.NewDocument("test", "TestAppID")
	div := doc.NewDiv("someDiv", "someDiv", doc.EnableSessionPersistence())


	root.AppendChild(div.Element())

	div.Element().Watch("ui", "mutatediv", div.Element(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		s := evt.NewValue()
		b, ok := s.(ui.Bool)
		if !ok {
			return true
		}
		if b {
			log.Print("greeting....")
			div.SetText("Hello, Earthlings!")
		}else{
			log.Print("Byes...")
			text2 := doc.NewTextNode().SetValue("Bye, noobs!")
			//div.Element().Mutate(ui.AppendChildCommand(text2.Element()))
			div.Element().AppendChild(text2.Element())
		}
		return false
	}))
	v,ok := div.Element().GetData("mutatediv")
	log.Print("get data value mutatediv")
	if !ok {
		log.Print("mutatediv is not present in persistent storage.")
		div.Element().SetDataSetUI("mutatediv", ui.Bool(true))
	} else{
		log.Print("data/mutatediv exists")
		b, ok := v.(ui.Bool)
		if !ok {
			log.Print("wrong type , expected ui.Bool")
			div.Element().SetDataSetUI("mutatediv", ui.Bool(true))
		}
		div.Element().SetDataSetUI("mutatediv", !b)
	}


	c := make(chan struct{}, 0)
	<-c
}
*/

/* =========================== Session Storage -- to simplify =====================



func main() {

	root := doc.NewDocument("test", "TestAppID")

	div:=doc.NewDiv("someDiv","someDiv",doc.EnableSessionPersistence())
	log.Print(div.Element().GetData("mutatediv"))
	text := doc.NewTextNode().SetValue("Hello, Earthlings!")

	// storage Test
	wd := doc.GetWindow()
	nwd,ok:=wd.Element().Native.(doc.NativeElement)
	if !ok{
		log.Print("unable to retrieve native window")
		return
	}
	nwd.JSValue().Get("sessionStorage").Set("test",true)
	root.AppendChild(div.Element())


	div.Element().Watch("data","mutatediv",div.Element(),ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		s:=evt.NewValue()
		b,ok:= s.(ui.Bool)
		if !ok{
			return true
		}
		if b{
			div.Element().AppendChild(text.Element())
			return false
		}
		div.Element().RemoveChild(text.Element())
		return false
	}))
	v,ok:= div.Element().GetData("mutatediv")
	log.Print(div.Element().Properties)
	if !ok{
		div.Element().SetData("mutatediv",ui.Bool(true))
	}

	b,ok:=v.(ui.Bool)
	if !ok{
		div.Element().SetData("mutatediv",ui.Bool(true))
	}
	div.Element().SetData("mutatediv",!b)

	c := make(chan struct{}, 0)
	<-c
}
*/

/* ================================================================================

func main() {

	root := doc.NewDocument("test", "TestAppID")

	div:=doc.NewDiv("someDiv","someDiv")
	text := doc.NewTextNode().SetValue("Hello, Earthlings!")
	div.Element().AppendChild(text.Element())

	root.AppendChild(div.Element())


	div.Element().Watch("data","somevalue",div.Element(),ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		s:=evt.NewValue()
		str,ok:= s.(ui.String)
		if !ok{
			return true
		}
		text.SetValue(str)
		return false
	}))

	div.Element().SetData("somevalue",ui.String("Bye, Earthlings !! With love!"))

	c := make(chan struct{}, 0)
	<-c
}
*/
