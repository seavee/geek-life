package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
)

// ProjectYearGroup 用于按年份分组项目
type ProjectYearGroup struct {
	Year     int
	Projects []ProjectWithIndex
}

// ProjectWithIndex 项目及其在原始列表中的索引
type ProjectWithIndex struct {
	Project model.Project
	Index   int
}

// ProjectPane Displays projects and dynamic lists
type ProjectPane struct {
	*tview.Flex
	projects            []model.Project
	list                *tview.List
	newProject          *tview.InputField
	repo                repository.ProjectRepository
	activeProject       *model.Project
	projectListStarting int // The index in list where project names starts
}

// NewProjectPane initializes
func NewProjectPane(repo repository.ProjectRepository) *ProjectPane {
	pane := ProjectPane{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		list:       tview.NewList().ShowSecondaryText(false),
		newProject: makeLightTextInput("+[New Project]"),
		repo:       repo,
	}
	
	// 设置项目列表背景色和文字色
	pane.list.SetSelectedBackgroundColor(tcell.ColorWhite)
	pane.list.SetSelectedTextColor(tcell.ColorBlack)
	pane.list.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))

	pane.newProject.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			pane.addNewProject()
		case tcell.KeyEsc:
			pane.newProject.SetText("")
			app.SetFocus(projectPane)
		}
	})

	// 输入框正常高度
	pane.AddItem(pane.list, 0, 1, true)
	pane.AddItem(pane.newProject, 1, 0, false)

	pane.SetBorder(true).SetTitle("Projects").SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	pane.loadListItems(false)

	return &pane
}

func (pane *ProjectPane) addNewProject() {
	name := pane.newProject.GetText()
	if len(name) < 3 {
		statusBar.showForSeconds("[red::]Project name should be at least 3 character", 5)
		return
	}

	project, err := pane.repo.Create(name, "")
	if err != nil {
		statusBar.showForSeconds("[red::]Failed to create Project:"+err.Error(), 5)
	} else {
		statusBar.showForSeconds(fmt.Sprintf("[yellow::]Project %s created. Press n to start adding new tasks.", name), 10)
		pane.projects = append(pane.projects, project)
		pane.addProjectToList(len(pane.projects)-1, true)
		pane.newProject.SetText("")
	}
}

func (pane *ProjectPane) addDynamicLists() {
	pane.addSection("Dynamic Lists")
	pane.list.AddItem("  • Today", "", 0, func() { taskPane.LoadDynamicList("today") })
	pane.list.AddItem("  • Tomorrow", "", 0, func() { taskPane.LoadDynamicList("tomorrow") })
	pane.list.AddItem("  • Upcoming", "", 0, func() { taskPane.LoadDynamicList("upcoming") })
	pane.list.AddItem("  • Unscheduled", "", 0, func() { taskPane.LoadDynamicList("unscheduled") })
}

func (pane *ProjectPane) addProjectList() {
	pane.addSection("Projects")
	pane.projectListStarting = pane.list.GetItemCount()

	var err error
	pane.projects, err = pane.repo.GetAll()
	if err != nil {
		statusBar.showForSeconds("Could not load Projects: "+err.Error(), 5)
		return
	}

	// 按年份分组显示项目
	pane.addProjectsByYear()

	pane.list.SetCurrentItem(2) // Keep "Today" selected on start
}

// getProjectYears 获取项目的年份信息
func (pane *ProjectPane) getProjectYears(project model.Project) []int {
	tasks, err := taskRepo.GetAllByProject(project)
	if err != nil {
		return []int{}
	}

	yearSet := make(map[int]bool)
	for _, task := range tasks {
		if task.DueDate > 0 {
			year := time.Unix(task.DueDate, 0).Year()
			yearSet[year] = true
		}
	}

	years := make([]int, 0, len(yearSet))
	for year := range yearSet {
		years = append(years, year)
	}
	sort.Ints(years)
	return years
}

// addProjectsByYear 按年份分组添加项目
func (pane *ProjectPane) addProjectsByYear() {
	// 收集所有项目的年份信息
	yearGroups := make(map[int][]ProjectWithIndex)
	defaultProjects := make([]ProjectWithIndex, 0)

	for i, project := range pane.projects {
		years := pane.getProjectYears(project)
		if len(years) == 0 {
			// 没有任务或所有任务都没有日期
			defaultProjects = append(defaultProjects, ProjectWithIndex{project, i})
		} else {
			// 为每个年份添加项目
			for _, year := range years {
				yearGroups[year] = append(yearGroups[year], ProjectWithIndex{project, i})
			}
		}
	}

	// 获取所有年份并排序
	allYears := make([]int, 0, len(yearGroups))
	for year := range yearGroups {
		allYears = append(allYears, year)
	}
	sort.Ints(allYears)

	// 首先显示默认组
	if len(defaultProjects) > 0 {
		pane.addSection("Default")
		for _, projectWithIndex := range defaultProjects {
			pane.addProjectToList(projectWithIndex.Index, false)
		}
		pane.list.AddItem("", "", 0, nil) // 空行分隔
	}

	// 按年份显示项目
	for _, year := range allYears {
		pane.addSection(fmt.Sprintf("%d", year))
		for _, projectWithIndex := range yearGroups[year] {
			pane.addProjectToList(projectWithIndex.Index, false)
		}
		if year != allYears[len(allYears)-1] { // 不是最后一个年份
			pane.list.AddItem("", "", 0, nil) // 空行分隔
		}
	}
}

func (pane *ProjectPane) addProjectToList(i int, selectItem bool) {
	// To avoid overriding of loop variables - https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/
	pane.list.AddItem("  • "+pane.projects[i].Title, "", 0, func(idx int) func() {
		return func() { pane.activateProject(idx) }
	}(i))

	if selectItem {
		pane.list.SetCurrentItem(-1)
		pane.activateProject(i)
	}
}

func (pane *ProjectPane) addSection(name string) {
	pane.list.AddItem("[gray]  "+name, "", 0, nil)
	// 只有Dynamic Lists和Projects需要横线，年份组不需要
	if name == "Dynamic Lists" || name == "Projects" {
		pane.list.AddItem("[gray]  "+strings.Repeat("─", 25), "", 0, nil)
	}
}

func (pane *ProjectPane) handleShortcuts(event *tcell.EventKey) *tcell.EventKey {
	switch unicode.ToLower(event.Rune()) {
	case 'j':
		pane.list.SetCurrentItem(pane.list.GetCurrentItem() + 1)
		return nil
	case 'k':
		pane.list.SetCurrentItem(pane.list.GetCurrentItem() - 1)
		return nil
	case 'n':
		app.SetFocus(pane.newProject)
		return nil
	}

	return event
}

func (pane *ProjectPane) activateProject(idx int) {
	pane.activeProject = &pane.projects[idx]
	taskPane.LoadProjectTasks(*pane.activeProject)

	removeThirdCol()
	projectDetailPane.SetProject(pane.activeProject)
	contents.AddItem(projectDetailPane, 25, 0, false)
	app.SetFocus(taskPane)
}

// RemoveActivateProject deletes the currently active project
func (pane *ProjectPane) RemoveActivateProject() {
	if pane.activeProject != nil && pane.repo.Delete(pane.activeProject) == nil {

		for i := range taskPane.tasks {
			_ = taskRepo.Delete(&taskPane.tasks[i])
		}
		taskPane.ClearList()

		statusBar.showForSeconds("[lime]Removed Project: "+pane.activeProject.Title, 5)
		removeThirdCol()

		pane.loadListItems(true)
	}
}

func (pane *ProjectPane) loadListItems(focus bool) {
	pane.list.Clear()
	pane.addDynamicLists()
	pane.list.AddItem("", "", 0, nil)
	pane.addProjectList()

	if focus {
		app.SetFocus(pane)
	}
}

// GetActiveProject provides pointer to currently active project
func (pane *ProjectPane) GetActiveProject() *model.Project {
	return pane.activeProject
}
