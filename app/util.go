package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"

	"github.com/ajaxray/geek-life/model"
)

var blankCell = func() *tview.TextView {
	cell := tview.NewTextView()
	cell.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	return cell
}()

func makeHorizontalLine(lineChar rune, color tcell.Color) *tview.TextView {
	hr := tview.NewTextView()
	hr.SetBackgroundColor(tcell.NewHexColor(0x0c0c0c))
	hr.SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
		// Draw a horizontal line across the middle of the box.
		style := tcell.StyleDefault.Foreground(color).Background(tcell.NewHexColor(0x0c0c0c))
		centerY := y + height/2
		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, centerY, lineChar, nil, style)
		}

		// Space for other content.
		return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
	})

	return hr
}

// CustomInputBox 创建一个自定义的输入框组件，使用Box包装InputField
type CustomInputBox struct {
	*tview.Box
	input *tview.InputField
	placeholder string
}

func NewCustomInputBox(placeholder string) *CustomInputBox {
	input := tview.NewInputField()
	input.SetPlaceholder(placeholder)
	
	box := &CustomInputBox{
		Box: tview.NewBox(),
		input: input,
		placeholder: placeholder,
	}
	
	// 设置Box的背景色
	box.SetBackgroundColor(tcell.NewHexColor(0x484848))
	
	// 设置输入框的透明背景
	input.SetBackgroundColor(tcell.ColorNone)
	input.SetFieldBackgroundColor(tcell.ColorNone)
	input.SetFieldTextColor(tcell.NewHexColor(0xC98E40))
	input.SetPlaceholderTextColor(tcell.NewHexColor(0xC98E40))
	
	return box
}

func (c *CustomInputBox) Draw(screen tcell.Screen) {
	// 先绘制背景Box
	c.Box.Draw(screen)
	
	// 获取Box的坐标
	x, y, width, height := c.GetInnerRect()
	
	// 绘制自定义背景
	bgStyle := tcell.StyleDefault.Background(tcell.NewHexColor(0x484848))
	for cy := y; cy < y+height; cy++ {
		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, cy, ' ', nil, bgStyle)
		}
	}
	
	// 设置输入框的位置并绘制
	c.input.SetRect(x, y, width, height)
	c.input.Draw(screen)
}

func (c *CustomInputBox) GetText() string {
	return c.input.GetText()
}

func (c *CustomInputBox) SetText(text string) {
	c.input.SetText(text)
}

func (c *CustomInputBox) SetDoneFunc(handler func(key tcell.Key)) {
	c.input.SetDoneFunc(handler)
}

func (c *CustomInputBox) SetChangedFunc(handler func(text string)) {
	c.input.SetChangedFunc(handler)
}

func (c *CustomInputBox) HasFocus() bool {
	return c.input.HasFocus()
}

func (c *CustomInputBox) Focus(delegate func(p tview.Primitive)) {
	c.input.Focus(delegate)
}

func (c *CustomInputBox) Blur() {
	c.input.Blur()
}

func (c *CustomInputBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.input.InputHandler()
}

// makeCustomStyledInput 创建一个具有自定义样式的输入框，专门用于 New Project 和 New Task
func makeCustomStyledInput(placeholder string) *tview.InputField {
	// 自定义颜色：#484848 背景，#B9C87F 文字
	bgColor := tcell.NewRGBColor(72, 72, 72)
	// #B9C87F = RGB(185, 200, 127)
	textColor := tcell.NewRGBColor(185, 200, 127)
	
	// 在创建输入框之前，临时修改全局样式
	originalContrastBg := tview.Styles.ContrastBackgroundColor
	tview.Styles.ContrastBackgroundColor = bgColor
	
	// 创建输入框
	input := tview.NewInputField()
	input.SetPlaceholder(placeholder)
	
	// 立即恢复全局样式
	tview.Styles.ContrastBackgroundColor = originalContrastBg
	
	// 强制样式应用函数 - 无论是否有焦点都保持自定义颜色
	applyCustomStyle := func() {
		// 重新临时设置全局样式
		originalBg := tview.Styles.ContrastBackgroundColor
		tview.Styles.ContrastBackgroundColor = bgColor
		
		input.SetPlaceholderTextColor(textColor)
		input.SetFieldTextColor(textColor)
		input.SetFieldBackgroundColor(bgColor)
		input.SetBackgroundColor(bgColor)
		input.SetFieldStyle(tcell.StyleDefault.
			Background(bgColor).
			Foreground(textColor))
		
		// 恢复全局样式
		tview.Styles.ContrastBackgroundColor = originalBg
	}
	
	// 初始设置
	applyCustomStyle()
	
	// 重写聚焦处理 - 保持自定义文字颜色，只改变背景
	input.SetFocusFunc(func() {
		// 聚焦时保持文字颜色为 #B9C87F，背景改为蓝色
		input.SetFieldBackgroundColor(tcell.ColorBlue)
		input.SetFieldTextColor(textColor) // 保持自定义文字颜色
		input.SetPlaceholderTextColor(textColor) // 保持自定义文字颜色
		input.SetBackgroundColor(tcell.ColorBlue)
		input.SetFieldStyle(tcell.StyleDefault.
			Background(tcell.ColorBlue).
			Foreground(textColor)) // 保持自定义文字颜色
	})
	
	// 失去焦点时：恢复自定义背景和文字颜色
	input.SetBlurFunc(func() {
		applyCustomStyle()
	})
	
	// 设置Done回调来处理ESC键
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			// ESC键时恢复自定义样式
			applyCustomStyle()
		}
		// ESC或Enter键时都可以触发失去焦点的行为
	})
	
	return input
}

func makeLightTextInput(placeholder string) *tview.InputField {
	// 对于 New Project 和 New Task，使用自定义样式
	if placeholder == "+[New Project]" || placeholder == "+[New Task]" {
		return makeCustomStyledInput(placeholder)
	}
	
	// 其他输入框使用原来的样式
	input := tview.NewInputField()
	input.SetPlaceholder(placeholder)
	
	// 默认状态：深灰色背景 + 绿色文字
	input.SetPlaceholderTextColor(tcell.ColorGreen)
	input.SetFieldTextColor(tcell.ColorGreen)
	input.SetFieldBackgroundColor(tcell.NewHexColor(0x404040))
	input.SetBackgroundColor(tcell.NewHexColor(0x404040))
	
	// 强制设置默认样式
	input.SetFieldStyle(tcell.StyleDefault.
		Background(tcell.NewHexColor(0x404040)).
		Foreground(tcell.ColorGreen))
	
	// 聚焦时：蓝色背景 + 白色文字
	input.SetFocusFunc(func() {
		input.SetFieldBackgroundColor(tcell.ColorBlue)
		input.SetFieldTextColor(tcell.ColorWhite)
		input.SetPlaceholderTextColor(tcell.ColorWhite)
		// 强制设置聚焦样式
		input.SetFieldStyle(tcell.StyleDefault.
			Background(tcell.ColorBlue).
			Foreground(tcell.ColorWhite))
	})
	
	// 失去焦点时：深灰色背景 + 绿色文字
	input.SetBlurFunc(func() {
		input.SetFieldBackgroundColor(tcell.NewHexColor(0x404040))
		input.SetFieldTextColor(tcell.ColorGreen)
		input.SetPlaceholderTextColor(tcell.ColorGreen)
		// 强制设置失焦样式
		input.SetFieldStyle(tcell.StyleDefault.
			Background(tcell.NewHexColor(0x404040)).
			Foreground(tcell.ColorGreen))
	})
	
	return input
}

// If input text is a valid date, parse it. Or get current date
func parseDateInputOrCurrent(inputText string) time.Time {
	if dateTime, err := time.Parse(dateLayoutISO, inputText); err == nil {
		return toDate(dateTime)
	}

	return toDate(time.Now())
}

func toDate(dateTime time.Time) time.Time {
	return time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, time.Local)
}

func makeButton(label string, handler func()) *tview.Button {
	btn := tview.NewButton(label)
	btn.SetSelectedFunc(handler)
	btn.SetLabelColor(tcell.ColorWhite)
	btn.SetBackgroundColor(tcell.ColorCornflowerBlue)  // 保持蓝色用于普通按钮
	btn.SetBorder(false)

	return btn
}

func ignoreKeyEvt() bool {
	// 检查项目输入框
	if projectPane != nil && projectPane.newProject != nil && projectPane.newProject.HasFocus() {
		return true
	}
	
	// 检查任务输入框
	if taskPane != nil && taskPane.newTask != nil && taskPane.newTask.HasFocus() {
		return true
	}
	
	// 检查重命名输入框
	if taskDetailPane != nil && taskDetailPane.header != nil && 
	   taskDetailPane.header.renameText != nil && taskDetailPane.header.renameText.HasFocus() {
		return true
	}
	
	// 检查日期输入框
	if taskDetailPane != nil && taskDetailPane.taskDate != nil && taskDetailPane.taskDate.HasFocus() {
		return true
	}
	
	// 检查femto编辑器
	focused := app.GetFocus()
	if focused != nil {
		if _, ok := focused.(*femto.View); ok {
			return true
		}
	}
	
	return false
}

// yetToImplement - to use as callback for unimplemented features
// `yetToImplement` is unused (deadcode)
// func yetToImplement(feature string) func() {
// 	message := fmt.Sprintf("[yellow]%s is yet to implement. Please Check in next version.", feature)
// 	return func() { statusBar.showForSeconds(message, 5) }
// }

func removeThirdCol() {
	contents.RemoveItem(taskDetailPane)
	contents.RemoveItem(projectDetailPane)
}

func getTaskTitleColor(task model.Task) string {
	if task.Completed {
		// 已完成任务使用 #797a12
		return "#797a12"
	} else {
		// 未完成任务统一使用 #7DEF1A，不受截止日期影响
		return "#7DEF1A"
	}
}

func makeTaskListingTitle(task model.Task) string {
	checkbox := "[ []"
	if task.Completed {
		checkbox = "[x[]"
	}

	return fmt.Sprintf("[%s]%s %s", getTaskTitleColor(task), checkbox, task.Title)
}

// `findProjectByID` is unused (deadcode)
// func findProjectByID(id int64) *model.Project {
// 	for i := range projectPane.projects {
// 		if projectPane.projects[i].ID == id {
// 			return &projectPane.projects[i]
// 		}
// 	}

// 	return nil
// }
