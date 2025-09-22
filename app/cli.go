package main

import (
	"fmt"
	"os"
	"unicode"

	"github.com/asdine/storm/v3"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	flag "github.com/spf13/pflag"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
	repo "github.com/ajaxray/geek-life/repository/storm"
	"github.com/ajaxray/geek-life/util"
)

var (
	app              *tview.Application
	layout, contents *tview.Flex

	statusBar         *StatusBar
	projectPane       *ProjectPane
	taskPane          *TaskPane
	taskDetailPane    *TaskDetailPane
	projectDetailPane *ProjectDetailPane

	db          *storm.DB
	projectRepo repository.ProjectRepository
	taskRepo    repository.TaskRepository

	// Flag variables
	dbFile string
)

func init() {
	flag.StringVarP(&dbFile, "db-file", "d", "", "Specify DB file path manually.")
}

func main() {
	app = tview.NewApplication()
	
	// 设置全局主题：使用单线边框，但不强制背景色
	blackColor := tcell.NewHexColor(0x0c0c0c)
	tview.Styles.PrimitiveBackgroundColor = blackColor
	// 不覆盖按钮和输入框的背景色设置
	// tview.Styles.ContrastBackgroundColor = blackColor
	// tview.Styles.MoreContrastBackgroundColor = blackColor
	
	// 设置边框为单线样式
	tview.Borders.Horizontal = '─'
	tview.Borders.Vertical = '│'
	tview.Borders.TopLeft = '┌'
	tview.Borders.TopRight = '┐'
	tview.Borders.BottomLeft = '└'
	tview.Borders.BottomRight = '┘'
	tview.Borders.HorizontalFocus = '─'
	tview.Borders.VerticalFocus = '│'
	tview.Borders.TopLeftFocus = '┌'
	tview.Borders.TopRightFocus = '┐'
	tview.Borders.BottomLeftFocus = '└'
	tview.Borders.BottomRightFocus = '┘'
	
	flag.Parse()

	db = util.ConnectStorm(dbFile)
	defer func() {
		if err := db.Close(); err != nil {
			util.LogIfError(err, "Error in closing storm Db")
		}
	}()

	if flag.NArg() > 0 && flag.Arg(0) == "migrate" {
		migrate(db)
		fmt.Println("Database migrated successfully!")
	} else {
		projectRepo = repo.NewProjectRepository(db)
		taskRepo = repo.NewTaskRepository(db)

		titleBar := makeTitleBar()
		contentPages := prepareContentPages()
		statusBarPane := prepareStatusBar(app)
		
		layout = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(titleBar, 2, 0, false).
			AddItem(contentPages, 0, 1, true).
			AddItem(statusBarPane, 1, 0, false)
		
		layout.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))

		setKeyboardShortcuts()

		if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
			panic(err)
		}
	}

}

func migrate(database *storm.DB) {
	util.FatalIfError(database.ReIndex(&model.Project{}), "Error in migrating Projects")
	util.FatalIfError(database.ReIndex(&model.Task{}), "Error in migrating Tasks")

	fmt.Println("Migration completed. Start geek-life normally.")
	os.Exit(0)
}

func setKeyboardShortcuts() *tview.Application {
	return app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// 首先检查是否在输入框中，如果是则直接返回事件，屏蔽所有快捷键
		if ignoreKeyEvt() {
			return event
		}

		// 只有不在输入框时才处理快捷键
		// Global shortcuts
		switch unicode.ToLower(event.Rune()) {
		case 'p':
			app.SetFocus(projectPane)
			contents.RemoveItem(taskDetailPane)
			return nil
		case 'q':
			// 退出功能
			return nil
		case 't':
			app.SetFocus(taskPane)
			contents.RemoveItem(taskDetailPane)
			return nil
		}

		// Handle based on current focus. Handlers may modify event
		switch {
		case projectPane.HasFocus():
			event = projectPane.handleShortcuts(event)
		case taskPane.HasFocus():
			event = taskPane.handleShortcuts(event)
			if event != nil && projectDetailPane.isShowing() {
				event = projectDetailPane.handleShortcuts(event)
			}
		case taskDetailPane.HasFocus():
			event = taskDetailPane.handleShortcuts(event)
		}

		return event
	})
}

func prepareContentPages() *tview.Flex {
	projectPane = NewProjectPane(projectRepo)
	taskPane = NewTaskPane(projectRepo, taskRepo)
	projectDetailPane = NewProjectDetailPane()
	taskDetailPane = NewTaskDetailPane(taskRepo)

	contents = tview.NewFlex().
		AddItem(projectPane, 25, 1, true).
		AddItem(taskPane, 0, 2, false)
		
	contents.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	return contents

}

func makeTitleBar() *tview.Flex {
	titleText := tview.NewTextView()
	titleText.SetText("[lime::b]geek-life")
	titleText.SetDynamicColors(true)
	titleText.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	titleText.SetTextAlign(tview.AlignLeft)
	
	versionInfo := tview.NewTextView()
	versionInfo.SetText("[::d]Version: 0.1.2")
	versionInfo.SetTextAlign(tview.AlignRight)
	versionInfo.SetDynamicColors(true)
	versionInfo.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))

	titleBar := tview.NewFlex()
	titleBar.SetDirection(tview.FlexColumn)
	titleBar.AddItem(titleText, 0, 2, false)
	titleBar.AddItem(versionInfo, 0, 1, false)
	titleBar.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	titleBar.SetBorder(false)
	
	return titleBar
}

func AskYesNo(text string, f func()) {

	activePane := app.GetFocus()
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				f()
			}
			app.SetRoot(layout, true).EnableMouse(true)
			app.SetFocus(activePane)
		})

	pages := tview.NewPages().
		AddPage("background", layout, true, true).
		AddPage("modal", modal, true, true)
	_ = app.SetRoot(pages, true).EnableMouse(true)
}
