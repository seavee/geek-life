package storm

import (
	"time"

	"github.com/asdine/storm/v3"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
)

type taskRepository struct {
	DB *storm.DB
}

// NewTaskRepository will create an object that represent the repository.Task interface
func NewTaskRepository(db *storm.DB) repository.TaskRepository {
	return &taskRepository{db}
}

func (t *taskRepository) GetAll() ([]model.Task, error) {
	panic("implement me")
}

func (t *taskRepository) GetAllByProject(project model.Project) ([]model.Task, error) {
	var tasks []model.Task
	//err = db.Find("ProjetID", project.ID, &tasks, storm.Limit(10), storm.Skip(10), storm.Reverse())
	err := t.DB.Find("ProjectID", project.ID, &tasks)

	return tasks, err
}

func (t *taskRepository) GetAllByDate(date time.Time) ([]model.Task, error) {
	var tasks []model.Task

	if date.IsZero() {
		var allTasks []model.Task
		err := t.DB.AllByIndex("ProjectID", &allTasks)
		for _, t := range allTasks {
			if t.DueDate == 0 {
				tasks = append(tasks, t)
			}
		}

		return tasks, err
	} else {
		err := t.DB.Find("DueDate", getRoundedDueDate(date), &tasks)
		return tasks, err
	}
}

func (t *taskRepository) GetAllByDateRange(from, to time.Time) ([]model.Task, error) {
	var tasks []model.Task

	err := t.DB.Range("DueDate", getRoundedDueDate(from), getRoundedDueDate(to), &tasks)
	return tasks, err
}

func (t *taskRepository) GetAllCompletedByDate(date time.Time) ([]model.Task, error) {
	var tasks []model.Task
	
	// 获取指定日期的开始和结束时间戳
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).Unix()
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location()).Unix()
	
	// 查询在指定日期内完成的任务
	err := t.DB.Range("CompletedAt", startOfDay, endOfDay, &tasks)
	if err != nil {
		return tasks, err
	}
	
	// 过滤确保任务确实是已完成的
	var completedTasks []model.Task
	for _, task := range tasks {
		if task.Completed && task.CompletedAt >= startOfDay && task.CompletedAt <= endOfDay {
			completedTasks = append(completedTasks, task)
		}
	}
	
	return completedTasks, nil
}

func (t *taskRepository) GetByID(ID string) (model.Task, error) {
	panic("implement me")
}

func (t *taskRepository) GetByUUID(UUID string) (model.Task, error) {
	panic("implement me")
}

func (t *taskRepository) Create(project model.Project, title, details, UUID string, dueDate int64) (model.Task, error) {
	task := model.Task{
		ProjectID: project.ID,
		Title:     title,
		Details:   details,
		UUID:      UUID,
		DueDate:   dueDate,
	}

	err := t.DB.Save(&task)
	return task, err
}

func (t *taskRepository) Update(task *model.Task) error {
	return t.DB.Update(task)
}

func (t *taskRepository) UpdateField(task *model.Task, field string, value interface{}) error {
	return t.DB.UpdateField(task, field, value)
}

func (t *taskRepository) Delete(task *model.Task) error {
	return t.DB.DeleteStruct(task)
}

func getRoundedDueDate(date time.Time) int64 {
	if date.IsZero() {
		return 0
	}

	return date.Unix()
}
