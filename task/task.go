package task

type Status string

const (
	Pending  Status = "pending"
	Finished Status = "finished"
)

func (s *Status) String() string {
	switch *s {
	case Pending:
		return "pending"
	case Finished:
		return "finished"
	default:
		return "unknown"
	}
}

type Task struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Reward int64  `json:"reward"`
}
