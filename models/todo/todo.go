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