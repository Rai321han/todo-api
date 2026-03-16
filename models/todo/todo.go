package todo

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	UserID      int    `json:"user_id"`
	Description string `json:"description"`
	IsCompleted bool   `json:"is_completed"`
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

// TodoListResponse represents a paginated todo list API response.
type TodoListResponse struct {
	TotalPages  int    `json:"total_pages"`
	CurrentPage int    `json:"current_page"`
	Limit       int    `json:"limit"`
	TotalCount  int    `json:"total_count"`
	Todos       []Todo `json:"todos"`
}

// Offset calculates the offset for pagination based on the current page and limit.
func (o TodoListOptions) Offset() int {
	return (o.Page - 1) * o.Limit
}
