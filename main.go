package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

// GOOS=js GOARCH=wasm go build -o  server/assets/app.wasm

var Children = ui.Children
var E = ui.New
var Listen = ui.Listen
var CSS = func(classes ...string) func(*ui.Element)*ui.Element{
	return func(e *ui.Element) *ui.Element{
		for _,class:= range classes{
			doc.AddClass(e,class)
		}
		return e
	}
}

var InitRouter = func(e *ui.Element) *ui.Element{
	e.OnFirstTimeMounted(ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		v,ok:= evt.Origin().AsViewElement()
		if !ok{
			panic("Router cannot be instantiated with non-ViewElement objects")
		}
		router := ui.NewRouter("/",v ,doc.RouterConfig)
		router.Hijack("/", "/all")
		return false
	}))
	
	return e
}


func main() {
	
	

	// 1. Create a new document
	Document := doc.NewDocument("Todo-App")
	AppSection := doc.NewSection("todoapp", "todoapp")
	AppFooter := doc.NewFooter("infofooter", "infofooter")

	// 2. AppSection components
	MainHeader := doc.NewHeader("header", "header")
	MainSection := doc.NewSection("main", "main")
	MainFooter := doc.NewFooter("footer", "footer")

	// 3. MainHeader components
	MainHeading := doc.NewH1("todo", "apptitle").SetText("Todo")
	todosinput := NewTodoInput("todo", "new-todo")

	// 4. MainSection componnents
	ToggleAllInput := doc.NewInput("checkbox", "toggle-all", "toggle-all")
	toggleallhandler:= ui.NewEventHandler(func(evt ui.Event) bool {
		ToggleAllInput:= evt.Target()
		togglestate, ok := ToggleAllInput.GetData("checked")
		if !ok {
			ToggleAllInput.Set("event", "toggled", ui.Bool(true))
			return false
		}
		ts := togglestate.(ui.Bool)
		ToggleAllInput.Set("event", "toggled", !ts)
		return false
	})

	ToggleLabel := doc.NewLabel("toggle-all-Label", "toggle-all-label").For(ToggleAllInput.AsElement())
	TodosList := NewTodosListElement("todo-list", "todo-list", doc.EnableLocalPersistence())


	// 5. MainFooter components
	TodoCount := NewTodoCount("todo-count", "todo-count")

	FilterList := NewFilterList("filters", "filters")


	ClearCompleteButton := ClearCompleteBtn("clear-complete", "clear-complete")
	ClearCompleteHandler := ui.NewEventHandler(func(evt ui.Event) bool {
		ClearCompleteButton:= evt.Target()
		ClearCompleteButton.Set("event", "clear", ui.Bool(true))
		return false
	})

	// 6.AppFooter components
	editinfo := doc.NewParagraph("editinfo", "editinfo").SetText("Double-click to edit a todo")
	createdWith := doc.NewParagraph("createdWith", "createdWith").SetText("Created with: ").SetChildren(doc.NewAnchor("particleui", "particleui").SetHREF("http://github.com/atdiar/particleui").SetText("ParticleUI"))


	ui.New(Document,
		Children(
			E(AppSection,
				CSS("todoapp"),
				Children(
					E(MainHeader,
						CSS("header"),
						Children(
							E(MainHeading),
							E(todosinput,
								CSS("new-todo"),
							),
						),
					),
					E(MainSection,
						CSS("main"),
						Children(
							E(ToggleAllInput,
								CSS("toggle-all"),
								Listen("click",toggleallhandler,doc.NativeEventBridge),
							),
							E(ToggleLabel),
							E(TodosList,
								InitRouter,
							),
						),
					),
					E(MainFooter,
						CSS("footer"),
						Children(
							E(TodoCount),
							E(FilterList),
							E(ClearCompleteButton, 
								Listen("click",ClearCompleteHandler,doc.NativeEventBridge),
							),
						),
					),
				),
			),
			E(AppFooter,
				CSS("info"),
				Children(
					E(editinfo),
					E(createdWith),
				),
			),
		),
	)


	// COMPONENTS DATA RELATIONSHIPS



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
		tdl := TodosList.GetList()
		ntdl := ui.NewList()
		for _, todo := range tdl {
			t := todo.(Todo)
			c, _ := t.Get("completed")
			cpl := c.(ui.Bool)
			if !cpl {
				ntdl = append(ntdl, todo)
			}
		}

		TodosList.AsElement().SetDataSetUI("todoslist", ntdl)
		return false
	}))

	AppSection.AsElement().Watch("data", "todoslist", TodosList.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		l := TodosList.GetList()

		if len(l) == 0 {
			doc.SetInlineCSS(MainFooter.AsElement(), "display:none")
		} else {
			doc.SetInlineCSS(MainFooter.AsElement(), "display:block")
		}

		countcomplete := 0
		allcomplete := true
		for _, todo := range l {
			t := todo.(Todo)
			completed, ok := t.Get("completed")
			if !ok {
				allcomplete = false
			}
			c := completed.(ui.Bool)
			if !c {
				allcomplete = false
			} else {
				countcomplete++
			}
		}

		TodoCount.SetCount(len(l) - countcomplete)

		if countcomplete == 0 {
			doc.SetInlineCSS(ClearCompleteButton.AsElement(), "display:none")
		} else {
			doc.SetInlineCSS(ClearCompleteButton.AsElement(), "display:block")
		}

		if allcomplete {
			ToggleAllInput.AsElement().SetDataSetUI("checked", ui.Bool(true))
		} else {
			ToggleAllInput.AsElement().SetDataSetUI("checked", ui.Bool(false))
		}
		return false
	}))

	AppSection.AsElement().Watch("event", "toggled", ToggleAllInput, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		status := evt.NewValue().(ui.Bool)
		tdl := TodosList.GetList()
		for i, todo := range tdl {
			t := todo.(Todo)
			c:= ui.Copy(t).(Todo)
			c.Set("completed", status)
			tdl[i]=c
			//todo, _ := FindTodoElement(t)
			//todo.AsElement().SetDataSetUI("todo", t)
		}
		TodosList.SetList(tdl)
		return false
	}))

	AppSection.AsElement().WatchASAP("event", "mounted", MainFooter, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tdl := TodosList.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : block")
		}
		return false
	}))

	AppSection.AsElement().WatchASAP("ui","filterslist",TodosList,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		FilterList.AsElement().SetUI("filterslist",evt.NewValue())
		return false
	}))

	MainSection.AsElement().Watch("ui", "todoslist", TodosList, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tdl := TodosList.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainSection.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainSection.AsElement(), "display : block")
		}
		return false
	}))

	ui.GetRouter().ListenAndServe("popstate", doc.GetWindow().AsElement(), doc.NativeEventBridge)

}
