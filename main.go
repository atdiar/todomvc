package main

import (
	"github.com/atdiar/particleui"
	"github.com/atdiar/particleui/drivers/js"
	. "github.com/atdiar/particleui/drivers/js/declarative"
)

// GOOS=js GOARCH=wasm go build -o  server/assets/app.wasm

func App() doc.Document {


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
		var ischecked bool
		_,ok:= evt.Target().Get("ui","checked")
		if !ok{
			chk := doc.GetAttribute(evt.Target(),"checked")
			if chk != "null"{
				ischecked = true
			}
			evt.Target().SyncUI("checked",ui.Bool(!ischecked))
		}

		evt.Target().TriggerEvent("toggled")
		return false
	})

	ClearCompleteHandler := ui.NewEventHandler(func(evt ui.Event) bool {
		ClearCompleteButton:= evt.Target()
		ClearCompleteButton.TriggerEvent( "clear", ui.Bool(true))
		return false
	})

	

	document:= doc.NewDocument("Todo-App")
	
	

	// TODO the HEAD should be pregenerated at build time
	E(document.Head(),
		Children(
			E(document.Link.WithID("todocss").
				SetAttribute("rel","stylesheet").
				SetAttribute("href","/css/todomvc.css"),
			),
			E(document.Script.WithID("wasmVM").
				Defer().
				Src("/wasm_exec.js"),
			),
			E(document.Script.WithID("goruntime").
				Defer().
				SetInnerHTML(
					`
						const go = new Go();
						WebAssembly.instantiateStreaming(fetch("/app.wasm"), go.importObject).then((result) => {
							go.run(result.instance);
						});
					`,
				),
			),
		),
	)	

	
	ui.New(document.Body(),
		Children(
			E(doc.AriaChangeAnnouncerFor(document)),
			E(document.Section.WithID("todoapp"),
				Ref(&AppSection),
				CSS("todoapp"),
				Children(
					E(document.Header.WithID("header"),
						CSS("header"),
						Children(
							E(document.H1.WithID("apptitle").SetText("Todo")),
							E(NewTodoInput("new-todo"),
								Ref(&todosinput),
								CSS("new-todo"),
							),
						),
					),
					E(document.Section.WithID("main"),
						Ref(&MainSection),
						CSS("main"),
						Children(
							E(document.Input.WithID("toggle-all","checkbox"),
								Ref(&ToggleAllInput),
								CSS("toggle-all"),
								Listen("click",toggleallhandler),
							),
							E(document.Label().For(ToggleAllInput)),
							E(NewTodoList(document,"todo-list", doc.EnableLocalPersistence()),
								Ref(&TodosList),
								InitRouter(Hijack("/","/all"),doc.RouterConfig),
							),
						),
					),
					E(document.Footer.WithID("footer"),
						Ref(&MainFooter),
						CSS("footer"),
						Children(
							E(NewTodoCount(document,"todo-count"), Ref(&TodoCount)),
							E(NewFilterList(document, "filters"), Ref(&FilterList)),
							E(ClearCompleteBtn("clear-complete"),
								Ref(&ClearCompleteButton),
								Listen("click",ClearCompleteHandler),
							),
						),
					),
				),
			),
			E(document.Footer(),
				CSS("info"),
				Children(
					E(document.Paragraph().SetText("Double-click to edit a todo")),
					E(document.Paragraph().SetText("Created with: "),
						Children(
							E(document.Anchor().SetHREF("http://github.com/atdiar/particleui").SetText("ZUI")),
						),
					),
				),
			),
		),
	)


	// COMPONENTS DATA RELATIONSHIPS

	// 4. Watch for new todos to insert
	AppSection.WatchEvent( "newtodo", todosinput.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tlist:= TodoListFromRef(TodosList)
		tdl := tlist.GetList()

		s, ok := evt.NewValue().(ui.String)
		if !ok || s == "" {
			panic("BAD TODO")
		}
		t := NewTodo(s)
		tdl = append(tdl, t)
		tlist.SetList(tdl)

		return false
	}))

	AppSection.WatchEvent( "clear", ClearCompleteButton.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
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

	AppSection.WatchEvent("updated", TodosList.AsElement(), ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tlist:= TodoListFromRef(TodosList)
		l := tlist.GetList()

		if len(l) == 0 {
			doc.SetInlineCSS(MainFooter.AsElement(), "display:none")
		} else {
			doc.SetInlineCSS(MainFooter.AsElement(), "display:block")
		}

		countcomplete := 0
		allcomplete := len(l) > 0

		for _, todo := range l {
			t := todo.(Todo)
			completed, ok := t.Get("completed")
			if !ok {
				panic("todo should have a completed property")
			}
			c := completed.(ui.Bool)
			if !c {
				allcomplete = false
			} else {
				countcomplete++
			}
		}

		tc:= TodoCountFromRef(TodoCount)
		var itemsleft = len(l)-countcomplete
		tc.SetCount(itemsleft)

		if itemsleft > 0 {
			allcomplete =false
			doc.SetInlineCSS(ClearCompleteButton.AsElement(), "display:none")
		} else {
			doc.SetInlineCSS(ClearCompleteButton.AsElement(), "display:block")
		}

		if allcomplete {
			ToggleAllInput.AsElement().SetUI("checked", ui.Bool(true))
		} else {
			ToggleAllInput.AsElement().SetUI("checked", ui.Bool(false))
		}
		return false
	}).RunASAP())

	AppSection.WatchEvent("toggled", ToggleAllInput, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		chk,ok:= evt.Origin().Get("ui","checked")
		if !ok{
			panic(("checked prop should be present"))
		}
		status := chk.(ui.Bool)

		tlist:= TodoListFromRef(TodosList)


		tdl := tlist.GetList()
/*
		for i, todo := range tdl {
			t := todo.(Todo)
			t.Set("completed", !status)
			tdl[i]=t
			/*tde,ok:=FindTodoElement(doc.GetDocument(evt.Origin().Root()),t)
			if !ok{
				panic("todo element not found which should not be possible")
			}
			tde.SyncUISetData("todo",t)
			
		}
		tlist.SetList(tdl) // TODO try to syncUISetData and SetlIst
*/
		for i, todo := range tdl {
			t := todo.(Todo)
			t.Set("completed", !status)
			tdl[i]=t
			tde,ok:=FindTodoElement(doc.GetDocument(evt.Origin()),t)
			if !ok{
				panic("todo element not found which should not be possible")
			}
			tde.SetDataSetUI("todo",t)
		}
		tlist.UpdateList(tdl)
		return false
	}))

	AppSection.WatchEvent("mounted", MainFooter, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		
		tlist:= TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainFooter.AsElement(), "display : block")
		}
		return false
	}).RunASAP())

	AppSection.Watch("ui","filterslist",TodosList,ui.NewMutationHandler(func(evt ui.MutationEvent)bool{
		FilterList.AsElement().SetUI("filterslist",evt.NewValue())
		return false
	}).RunASAP())

	MainSection.AsElement().WatchEvent("update", TodosList, ui.NewMutationHandler(func(evt ui.MutationEvent) bool {
		tlist:= TodoListFromRef(TodosList)
		tdl := tlist.GetList()
		if len(tdl) == 0 {
			doc.SetInlineCSS(MainSection.AsElement(), "display : none")
		} else {
			doc.SetInlineCSS(MainSection.AsElement(), "display : block")
		}
		return false
	}).RunASAP())

	
	return document

}

func main(){
	ListenAndServe := doc.NewBuilder(App)
	ListenAndServe(nil)
}