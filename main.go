package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
	. "github.com/atdiar/particleui/drivers/js/declarative"
)

// GOOS=js GOARCH=wasm go build -o  server/assets/app.wasm

func main() {


		

	var AppSection *ui.Element
	var MainSection *ui.Element
	var MainFooter *ui.Element
	var todosinput *ui.Element
	var ToggleAllInput *ui.Element
	var TodosList *ui.Element
	var TodoCount *ui.Element
	var FilterList *ui.Element
	var ClearCompleteButton *ui.Element
	
	toggleallhandler:= ui.NewEventHandler(func(evt ui.Event) bool {
		togglestate, ok := evt.Target().GetData("checked")
		if !ok {
			evt.Target().Set("event", "toggled", ui.Bool(true))
			return false
		}
		ts := togglestate.(ui.Bool)
		evt.Target().Set("event", "toggled", !ts)
		return false
	})

	ClearCompleteHandler := ui.NewEventHandler(func(evt ui.Event) bool {
		ClearCompleteButton:= evt.Target()
		ClearCompleteButton.Set("event", "clear", ui.Bool(true))
		return false
	})

	document:= doc.NewDocument("Todo-App")
	//defer document.ListenAndServe()


	ui.New(document.Body(),
		Children(
			E(doc.AriaChangeAnnouncer),
			E(doc.Section("todoapp"),
				Ref(&AppSection),
				CSS("todoapp"),
				Children(
					E(doc.Header("header"),
						CSS("header"),
						Children(
							E(doc.H1("apptitle").SetText("Todo")),
							E(NewTodoInput("new-todo"),
								Ref(&todosinput),
								CSS("new-todo"),
							),
						),
					),
					E(doc.Section("main"),
						Ref(&MainSection),
						CSS("main"),
						Children(
							E(doc.Input("checkbox","toggle-all"),
								Ref(&ToggleAllInput),
								CSS("toggle-all"),
								Listen("click",toggleallhandler),
							),
							E(doc.Label("toggle-all-label").For("toggle-all")),
							E(NewTodosListElement("todo-list", doc.EnableLocalPersistence()),
								Ref(&TodosList),
								InitRouter(Hijack("/","/all"),doc.RouterConfig),
							),
						),
					),
					E(doc.Footer("footer"),
						Ref(&MainFooter),
						CSS("footer"),
						Children(
							E(NewTodoCount("todo-count"), Ref(&TodoCount)),
							E(NewFilterList("filters"), Ref(&FilterList)),
							E(ClearCompleteBtn("clear-complete"),
								Ref(&ClearCompleteButton),
								Listen("click",ClearCompleteHandler),
							),
						),
					),
				),
			),
			E(doc.Footer("infofooter"),
				CSS("info"),
				Children(
					E(doc.Paragraph("editinfo").SetText("Double-click to edit a todo")),
					E(doc.Paragraph("createdWith").SetText("Created with: "),
						Children(
							E(doc.Anchor("particleui").SetHREF("http://github.com/atdiar/particleui").SetText("ParticleUI")),
						),
					),
				),
			),
		),
	)


	// COMPONENTS DATA RELATIONSHIPS

	// 4. Watch for new todos to insert
	AppSection.AsElement().Watch("event", "newtodo", todosinput.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tlist:= TodoListFromRef(TodosList)
		tdl := tlist.GetList()

		s, ok := evt.NewValue().(ui.String)
		if !ok || s == "" {
			return true
		}
		t := NewTodo(s)
		tdl = append(tdl, t)
		tlist.SetList(tdl)

		return false
	}))

	AppSection.AsElement().Watch("event", "clear", ClearCompleteButton.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tlist:= TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		ntdl := ui.NewList()
		for _, todo := range tdl {
			t := todo.(Todo)
			c, _ := t.Get("completed")
			cpl := c.(ui.Bool)
			if !cpl {
				ntdl = append(ntdl, todo)
			}
		}

		tlist.SetList(ntdl)
		return false
	}))

	AppSection.AsElement().Watch("ui", "todoslist", TodosList.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tlist:= TodoListFromRef(TodosList)
		l := tlist.GetList()

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
		tc:= TodoCountFromRef(TodoCount)
		tc.SetCount(len(l) - countcomplete)

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
	}).RunASAP())

	AppSection.AsElement().Watch("event", "toggled", ToggleAllInput, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		status := evt.NewValue().(ui.Bool)

		tlist:= TodoListFromRef(TodosList)

		tdl := tlist.GetList()
		for i, todo := range tdl {
			t := todo.(Todo)
			t.Set("completed", status)
			tdl[i]=t
			/*todo, ok := FindTodoElement(t)
			if ok{
				todo.AsElement().SyncUISetData("todo", t)
			}*/
			
		}
		tlist.SetList(tdl)
		return false
	}))

	AppSection.AsElement().Watch("event", "mounted", MainFooter, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		
		tlist:= TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : block")
		}
		return false
	}).RunASAP())

	AppSection.AsElement().Watch("ui","filterslist",TodosList,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		FilterList.AsElement().SetUI("filterslist",evt.NewValue())
		return false
	}).RunASAP())

	MainSection.AsElement().Watch("ui", "todoslist", TodosList, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tlist:= TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainSection.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainSection.AsElement(), "display : block")
		}
		return false
	}).RunASAP())

	document.ListenAndServe()


}