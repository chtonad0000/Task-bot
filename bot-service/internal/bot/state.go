package bot

import "time"

type TaskCreationState struct {
	Step     int
	TaskName string
	Deadline time.Time
	Priority int
}

type TaskProgressUpdateState struct {
	TaskId int64
	Step   int
}

var userTaskStates = make(map[int64]*TaskCreationState)

var userTaskProgressUpdateStates = make(map[int64]*TaskProgressUpdateState)
