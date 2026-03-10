package todo


type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	UserID      int    `json:"user_id"`
	Description string `json:"description"`
	IsCompleted   bool   `json:"is_completed"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TodoListOptions contains query options for listing todos.
type TodoListOptions struct {
	Status *bool
	SortBy string
	Order  string
	Search string
	Page   int
	Limit  int
}

func (o TodoListOptions) Offset() int {
	return (o.Page - 1) * o.Limit
}