package models

type Node struct {
	Value string `json:"value"`
	Left  *Node  `json:"left,omitempty"`
	Right *Node  `json:"right,omitempty"`
}

type Task struct {
	ID        string  `json:"id"`
	Arg1      string  `json:"arg1"`
	Arg2      string  `json:"arg2"`
	Operation string  `json:"operation"`
	Result    float64 `json:"result,omitempty"`
	Completed bool    `json:"completed"`
}

type Result struct {
	TaskID string  `json:"task_id"`
	Value  float64 `json:"value"`
	Error  string  `json:"error,omitempty"`
}

type Expression struct {
	Name   string  `json:"name"`
	Status int     `json:"status"`
	Id     int     `json:"id"`
	Result float64 `json:"result"`
	Node   *Node   `json:"node,omitempty"`
}
