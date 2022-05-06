package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
)

// GOOS=js GOARCH=wasm go build -o  ../../app.wasm

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
	ToggleAllInput := doc.NewInput("checkbox", "toggle-all", "toggle-all")
	ToggleAllInput.AsElement().AddEventListener("click", ui.NewEventHandler(func(evt ui.Event) bool {
		togglestate, ok := ToggleAllInput.AsElement().GetData("checked")
		if !ok {
			ToggleAllInput.AsElement().Set("event", "toggled", ui.Bool(true))
			return false
		}
		ts := togglestate.(ui.Bool)
		ToggleAllInput.AsElement().Set("event", "toggled", !ts)
		return false
	}), doc.NativeEventBridge)

	ToggleLabel := doc.NewLabel("toggle-all-Label", "toggle-all-label").For(ToggleAllInput.AsElement())
	TodosList := NewTodosListElement("todo-list", "todo-list", doc.EnableLocalPersistence())
	todolistview, ok := TodosList.AsViewElement()
	if !ok {
		panic("Expected TodosList to be constructed as a ViewElement")
	}
	MainSection.SetChildren(ToggleAllInput, ToggleLabel, TodosList)

	// 5. Build MainFooter
	TodoCount := NewTodoCount("todo-count", "todo-count")

	FilterList := NewFilterList("filters", "filters")
	// links
	router := ui.NewRouter("/", todolistview)
	router.Hijack("/", "/all")

	linkall := router.NewLink(todolistview, "all")
	linkactive := router.NewLink(todolistview, "active")
	linkcompleted := router.NewLink(todolistview, "completed")

	allFilter := NewFilter("All", "all-filter", linkall)
	activeFilter := NewFilter("Active", "active-filter", linkactive)
	completedFilter := NewFilter("Completed", "completed-filter", linkcompleted)
	FilterList.SetFilterList(allFilter, activeFilter, completedFilter)

	ClearCompleteButton := ClearCompleteBtn("clear-complete", "clear-complete")
	ClearCompleteButton.AsElement().AddEventListener("click", ui.NewEventHandler(func(evt ui.Event) bool {
		ClearCompleteButton.AsElement().Set("event", "clear", ui.Bool(true))
		return false
	}), doc.NativeEventBridge)
	MainFooter.SetChildren(TodoCount, FilterList, ClearCompleteButton)

	// 6.Build AppFooter
	editinfo := doc.NewParagraph("editinfo", "editinfo").SetText("Double-click to edit a todo")
	createdWith := doc.NewParagraph("createdWith", "createdWith").SetText("Created with: ").SetChildren(doc.NewAnchor("particleui", "particleui").SetHREF("http://github.com/atdiar/particleui").SetText("ParticleUI"))
	AppFooter.SetChildren(editinfo, createdWith)

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

	AppSection.AsElement().Watch("event", "toggled", ToggleAllInput.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		status := evt.NewValue().(ui.Bool)
		tdl := TodosList.GetList()
		for _, todo := range tdl {
			t := todo.(Todo)
			t.Set("completed", status)
			todo, _ := FindTodoElement(t)
			todo.AsElement().SetDataSetUI("todo", t)
		}
		return false
	}))

	AppSection.AsElement().WatchASAP("event", "mounted", MainFooter.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tdl := TodosList.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : block")
		}
		return false
	}))

	MainSection.AsElement().Watch("ui", "todoslist", TodosList.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tdl := TodosList.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainSection.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainSection.AsElement(), "display : block")
		}
		return false
	}))

	router.ListenAndServe("popstate", doc.GetWindow().AsElement(), doc.NativeEventBridge)

}
