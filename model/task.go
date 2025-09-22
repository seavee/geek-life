package model

// Task represent a task - the building block of the TaskManager app
type Task struct {
	ID          int64  `storm:"id,increment",json:"id"`
	ProjectID   int64  `storm:"index",json:"project_id"`
	UUID        string `storm:"unique",json:"uuid,omitempty"`
	Title       string `json:"text"`
	Details     string `json:"notes"`
	Completed   bool   `storm:"index",json:"completed"`
	CompletedAt int64  `storm:"index",json:"completed_at,omitempty"`
	DueDate     int64  `storm:"index",json:"due_date,omitempty"`
}
