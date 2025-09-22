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
	renameInput         *tview.InputField
	repo                repository.ProjectRepository
	activeProject       *model.Project
	projectListStarting int // The index in list where project names starts
	renameMode          bool             // 是否处于重命名模式
	renameIndex         int              // 正在重命名的项目索引
}

// NewProjectPane initializes
func NewProjectPane(repo repository.ProjectRepository) *ProjectPane {
	pane := ProjectPane{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		list:        tview.NewList().ShowSecondaryText(false),
		newProject:  makeLightTextInput("+[New Project]"),
		renameInput: makeLightTextInput("Rename Project"),
		repo:        repo,
		renameMode:  false,
		renameIndex: -1,
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

	pane.renameInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			pane.finishRenaming()
		case tcell.KeyEsc:
			pane.cancelRenaming()
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
	// 根据项目工作状态选择显示符号
	var symbol string
	if pane.projects[i].Working {
		symbol = "  [orange]★[-] " // 橙色星号表示正在工作，[-]恢复默认颜色，添加空格间距
	} else {
		symbol = "  • " // 普通圆点，添加空格间距
	}
	
	pane.list.AddItem(symbol+pane.projects[i].Title, "", 0, func(idx int) func() {
		return func() { pane.activateProject(idx) }
	}(i))

	if selectItem {
		pane.list.SetCurrentItem(-1)
		pane.activateProject(i)
	}
}

func (pane *ProjectPane) addSection(name string) {
	var handler func()
	// 检查是否是年份分组
	if len(name) == 4 && name >= "2000" && name <= "2999" {
		// 年份分组可以点击，加载该年份的所有任务
		handler = func() {
			year := name
			pane.loadTasksByYear(year)
		}
		// 年份使用加粗样式，使其更突出
		pane.list.AddItem("[::b]  "+name, "", 0, handler)
	} else {
		// 其他分组不可点击
		pane.list.AddItem("[gray]  "+name, "", 0, nil)
	}
	
	// 只有Dynamic Lists和Projects需要横线，年份组不需要
	if name == "Dynamic Lists" || name == "Projects" {
		pane.list.AddItem("[gray]  "+strings.Repeat("─", 25), "", 0, nil)
	}
}

func (pane *ProjectPane) handleShortcuts(event *tcell.EventKey) *tcell.EventKey {
	// 如果处于重命名模式，不处理其他快捷键
	if pane.renameMode {
		return event
	}

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
	case 'r':
		pane.startRenaming()
		return nil
	case 'b':
		pane.toggleWorkingStatus()
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
	// 保持焦点在项目面板，这样用户点击项目后可以直接按 'r' 重命名
	app.SetFocus(pane)
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

// startRenaming 开始重命名当前选中的项目
func (pane *ProjectPane) startRenaming() {
	currentItem := pane.list.GetCurrentItem()
	
	// 检查当前选中的是否是项目（不是动态列表或分隔符）
	if currentItem < pane.projectListStarting {
		statusBar.showForSeconds("[red::]Please select a project to rename", 3)
		return
	}
	
	// 找到对应的项目索引
	projectIndex := pane.findProjectIndexByListItem(currentItem)
	if projectIndex == -1 {
		return
	}
	
	// 设置重命名模式
	pane.renameMode = true
	pane.renameIndex = projectIndex
	
	// 设置输入框的当前文本为项目名称
	pane.renameInput.SetText(pane.projects[projectIndex].Title)
	
	// 替换输入框显示
	pane.RemoveItem(pane.newProject)
	pane.AddItem(pane.renameInput, 1, 0, false)
	
	// 焦点转到重命名输入框
	app.SetFocus(pane.renameInput)
	
	statusBar.showForSeconds("[yellow::]Renaming mode: Edit name and press Esc to save, or Esc to cancel", 5)
}

// finishRenaming 完成重命名操作
func (pane *ProjectPane) finishRenaming() {
	if !pane.renameMode || pane.renameIndex == -1 {
		return
	}
	
	newName := pane.renameInput.GetText()
	if len(newName) < 3 {
		statusBar.showForSeconds("[red::]Project name should be at least 3 characters", 5)
		return
	}
	
	// 更新项目名称
	pane.projects[pane.renameIndex].Title = newName
	err := pane.repo.Update(&pane.projects[pane.renameIndex])
	if err != nil {
		statusBar.showForSeconds("[red::]Failed to rename project: "+err.Error(), 5)
		pane.cancelRenaming()
		return
	}
	
	// 显示成功消息
	statusBar.showForSeconds(fmt.Sprintf("[yellow::]Project renamed to '%s'", newName), 5)
	
	// 重新加载列表
	pane.loadListItems(true)
	
	// 选中刚才重命名的项目
	pane.selectProjectByName(newName)
	pane.exitRenameMode()
}

// cancelRenaming 取消重命名操作
func (pane *ProjectPane) cancelRenaming() {
	pane.exitRenameMode()
	app.SetFocus(pane)
}

// exitRenameMode 退出重命名模式
func (pane *ProjectPane) exitRenameMode() {
	pane.renameMode = false
	pane.renameIndex = -1
	
	// 恢复原来的输入框
	pane.RemoveItem(pane.renameInput)
	pane.AddItem(pane.newProject, 1, 0, false)
	pane.renameInput.SetText("")
}

// findProjectIndexByListItem 根据列表项索引找到对应的项目索引
func (pane *ProjectPane) findProjectIndexByListItem(listIndex int) int {
	// 获取当前选中项的文本
	main, _ := pane.list.GetItemText(listIndex)
	
	// 移除调试信息
	
	// 简化检查：只要包含项目符号就认为是项目
	if !strings.Contains(main, "•") && !strings.Contains(main, "⚡") && !strings.Contains(main, "·") && !strings.Contains(main, "★") {
		statusBar.showForSeconds("[red::]Please select a project (not a section or empty line)", 3)
		return -1
	}
	
	// 提取项目名称，通过找到符号位置来提取（考虑符号后可能有空格）
	var projectName string
	if idx := strings.Index(main, "★"); idx != -1 {
		projectName = strings.TrimSpace(main[idx+len("★"):])
	} else if idx := strings.Index(main, "⚡"); idx != -1 {
		projectName = strings.TrimSpace(main[idx+len("⚡"):])
	} else if idx := strings.Index(main, "•"); idx != -1 {
		projectName = strings.TrimSpace(main[idx+len("•"):])
	} else if idx := strings.Index(main, "·"); idx != -1 {
		projectName = strings.TrimSpace(main[idx+len("·"):])
	}
	
	// 在项目列表中查找匹配的项目
	for j, project := range pane.projects {
		if project.Title == projectName {
			return j
		}
	}
	
	statusBar.showForSeconds("[red::]Project not found", 3)
	return -1
}

// selectProjectByName 在列表中选中指定名称的项目
func (pane *ProjectPane) selectProjectByName(projectName string) {
	// 遍历列表项，寻找匹配的项目名称
	for i := 0; i < pane.list.GetItemCount(); i++ {
		main, _ := pane.list.GetItemText(i)
		
		// 检查是否是项目项且名称匹配（考虑不同的符号格式）
		if (strings.Contains(main, "•") || strings.Contains(main, "⚡") || strings.Contains(main, "·") || strings.Contains(main, "★")) && i >= pane.projectListStarting {
			var itemProjectName string
			if idx := strings.Index(main, "★"); idx != -1 {
				itemProjectName = strings.TrimSpace(main[idx+len("★"):])
			} else if idx := strings.Index(main, "⚡"); idx != -1 {
				itemProjectName = strings.TrimSpace(main[idx+len("⚡"):])
			} else if idx := strings.Index(main, "•"); idx != -1 {
				itemProjectName = strings.TrimSpace(main[idx+len("•"):])
			} else if idx := strings.Index(main, "·"); idx != -1 {
				itemProjectName = strings.TrimSpace(main[idx+len("·"):])
			}
			if itemProjectName == projectName {
				pane.list.SetCurrentItem(i)
				return
			}
		}
	}
}

// toggleWorkingStatus 切换当前选中项目的工作状态
func (pane *ProjectPane) toggleWorkingStatus() {
	currentItem := pane.list.GetCurrentItem()
	
	// 移除调试信息
	
	// 找到对应的项目索引（findProjectIndexByListItem 内部会检查是否为有效项目）
	projectIndex := pane.findProjectIndexByListItem(currentItem)
	if projectIndex == -1 {
		return
	}
	
	// 如果要设置为工作状态，先清除其他项目的工作状态（确保只有一个项目处于工作中）
	if !pane.projects[projectIndex].Working {
		for i := range pane.projects {
			if pane.projects[i].Working {
				pane.projects[i].Working = false
				err := pane.repo.Update(&pane.projects[i])
				if err != nil {
					statusBar.showForSeconds("[red::]Failed to update project: "+err.Error(), 5)
					return
				}
			}
		}
	}
	
	// 切换当前项目的工作状态
	pane.projects[projectIndex].Working = !pane.projects[projectIndex].Working
	err := pane.repo.Update(&pane.projects[projectIndex])
	if err != nil {
		statusBar.showForSeconds("[red::]Failed to update project: "+err.Error(), 5)
		return
	}
	
	// 显示状态消息
	projectName := pane.projects[projectIndex].Title
	if pane.projects[projectIndex].Working {
		statusBar.showForSeconds(fmt.Sprintf("[green::]'%s' is now your working project ⚡", projectName), 5)
	} else {
		statusBar.showForSeconds(fmt.Sprintf("[yellow::]'%s' is no longer your working project", projectName), 5)
	}
	
	// 重新加载列表并保持选中当前项目
	pane.loadListItems(true)
	pane.selectProjectByName(projectName)
}

// loadTasksByYear 按年份加载所有相关任务
func (pane *ProjectPane) loadTasksByYear(year string) {
	// 清除当前活动项目
	pane.activeProject = nil
	removeThirdCol()
	
	// 在TaskPane中加载该年份的所有任务
	taskPane.LoadTasksByYear(year)
	
	// 切换焦点到任务面板
	app.SetFocus(taskPane)
	
	// 显示状态消息
	statusBar.showForSeconds(fmt.Sprintf("[yellow::]Displaying all tasks for year %s", year), 5)
}
