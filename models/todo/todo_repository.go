package todo

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrTodoNotFound = errors.New("todo not found")

type DB interface {
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
}

// TodoRepository provides methods to interact with the todos table in the database.
// It contains a reference to the database connection and allows for creating, retrieving, updating, and deleting items.
type TodoRepository struct {
	DB DB
}

// Create inserts a new item into the database.
// It takes a pointer to a task struct as input and returns an error if the operation fails or the created task item with its ID populated if successful.
func (r *TodoRepository) Create(todo *Todo) (Todo, error) {
	var createdTodo Todo

	query := `
		INSERT INTO todos (title, description, is_completed, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.DB.QueryRow(query, todo.Title, todo.Description, todo.IsCompleted, todo.UserID).
		Scan(&createdTodo.ID, &createdTodo.CreatedAt, &createdTodo.UpdatedAt)

	if err != nil {
		return Todo{}, fmt.Errorf("create todo: %w", err)
	}

	createdTodo.Title = todo.Title
	createdTodo.Description = todo.Description
	createdTodo.IsCompleted = todo.IsCompleted
	createdTodo.UserID = todo.UserID

	return createdTodo, nil
}

// GetByID retrieves a item from the database by its ID and user ID.
// It returns the item if found and accessible by the user, or an error if the item is not found.
func (r *TodoRepository) GetByID(id, userID int) (Todo, error) {
	var todo Todo
	query := `
		SELECT id, title, description, is_completed, user_id, created_at, updated_at
		FROM todos
		WHERE id = $1 AND user_id = $2
	`
	err := r.DB.QueryRow(query, id, userID).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Description,
		&todo.IsCompleted,
		&todo.UserID,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return Todo{}, ErrTodoNotFound
	} else if err != nil {
		return Todo{}, fmt.Errorf("get todo by id: %w", err)
	}
	return todo, nil
}

// GetAll retrieves all items for a specific user from the database.
// It returns a slice of items or an error if the operation fails.
func (r *TodoRepository) GetAll(userID int, options TodoListOptions) (TodoListResponse, error) {
	baseQuery, filterArgs, argPos := buildBaseQuery(userID, options)

	totalCount, err := r.countAllByUser(userID)
	if err != nil {
		return TodoListResponse{}, err
	}

	filteredCount, err := r.countFiltered(baseQuery, filterArgs)
	if err != nil {
		return TodoListResponse{}, err
	}

	listQuery, listArgs := buildListQuery(baseQuery, options, filterArgs, argPos)

	rows, err := r.DB.Query(listQuery, listArgs...)
	if err != nil {
		return TodoListResponse{}, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()

	todos, err := scanTodoRows(rows)
	if err != nil {
		return TodoListResponse{}, err
	}

	return TodoListResponse{
		TotalPages:  calculateTotalPages(filteredCount, options.Limit),
		CurrentPage: options.Page,
		Limit:       options.Limit,
		TotalCount:  totalCount,
		Todos:       todos,
	}, nil
}

// buildBaseQuery constructs the base SQL query for retrieving items based on the user ID and list options.
// It returns the base query string, a slice of query arguments, and the next argument position for additional filters.
func buildBaseQuery(userID int, options TodoListOptions) (string, []any, int) {
	baseQuery := `
		FROM todos
		WHERE user_id = $1
	`
	args := []any{userID}
	argPos := 2

	if options.Status != nil {
		baseQuery += fmt.Sprintf(" AND is_completed = $%d", argPos)
		args = append(args, *options.Status)
		argPos++
	}

	if options.Search != "" {
		baseQuery += fmt.Sprintf(" AND title ILIKE $%d", argPos)
		args = append(args, "%"+options.Search+"%")
		argPos++
	}

	return baseQuery, args, argPos
}

// buildOrderBy constructs the ORDER BY clause for the SQL query based on the list options.
func buildOrderBy(options TodoListOptions) string {
	if options.SortBy == "title" {
		if options.Order == "desc" {
			return "title DESC"
		}
		return "title ASC"
	}

	if options.Order == "asc" {
		return "created_at ASC"
	}

	return "created_at DESC"
}

// buildListQuery constructs the final SQL query for retrieving items based on the base query, list options, and query arguments.
// It returns the complete query string and the final slice of query arguments.
func buildListQuery(baseQuery string, options TodoListOptions, args []any, argPos int) (string, []any) {
	listQuery := `
		SELECT id, title, description, is_completed, user_id, created_at, updated_at
	` + baseQuery + " ORDER BY " + buildOrderBy(options) + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	listArgs := append(append([]any{}, args...), options.Limit, options.Offset())

	return listQuery, listArgs
}

func (r *TodoRepository) countAllByUser(userID int) (int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM todos
		WHERE user_id = $1
	`

	var totalCount int
	if err := r.DB.QueryRow(countQuery, userID).Scan(&totalCount); err != nil {
		return 0, fmt.Errorf("count todos: %w", err)
	}

	return totalCount, nil
}

// countFiltered executes a COUNT query based on the provided base query and arguments to determine the number of items that match the filtering criteria.
func (r *TodoRepository) countFiltered(baseQuery string, args []any) (int, error) {
	filteredCountQuery := "SELECT COUNT(*) " + baseQuery

	var filteredCount int
	if err := r.DB.QueryRow(filteredCountQuery, args...).Scan(&filteredCount); err != nil {
		return 0, fmt.Errorf("count filtered todos: %w", err)
	}

	return filteredCount, nil
}

// scanTodoRows iterates over the rows returned from a SQL query and scans each row into a Todo struct, accumulating the results into a slice of Todo items.
func scanTodoRows(rows *sql.Rows) ([]Todo, error) {
	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Description,
			&todo.IsCompleted,
			&todo.UserID,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan todo row: %w", err)
		}
		todos = append(todos, todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todo rows: %w", err)
	}

	return todos, nil
}

func calculateTotalPages(totalItems, limit int) int {
	if totalItems <= 0 || limit <= 0 {
		return 0
	}

	return (totalItems + limit - 1) / limit
}

// Update modifies an existing item in the database.
// Update can update the title, description or is_completed fields of the item.
// It takes a pointer to a struct as input and returns the updated item or an error if the operation fails.
func (r *TodoRepository) Update(id int, userId int, todo *Todo) (Todo, error) {
	var updatedTodo Todo

	query := `
	UPDATE todos SET 
	title = COALESCE($3, title), 
	description = COALESCE($4, description), 
	is_completed = COALESCE($5, is_completed), 
	updated_at = NOW()
	WHERE id = $1 AND user_id = $2
	RETURNING id, title, description, is_completed, user_id, created_at, updated_at
	`

	err := r.DB.QueryRow(query, id, userId, nullString(todo.Title), nullString(todo.Description), nullBool(todo.IsCompleted)).
		Scan(
			&updatedTodo.ID,
			&updatedTodo.Title,
			&updatedTodo.Description,
			&updatedTodo.IsCompleted,
			&updatedTodo.UserID,
			&updatedTodo.CreatedAt,
			&updatedTodo.UpdatedAt,
		)

	if err == sql.ErrNoRows {
		return Todo{}, ErrTodoNotFound
	} else if err != nil {
		return Todo{}, fmt.Errorf("update todo: %w", err)
	}
	return updatedTodo, nil
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullBool(b bool) any {
	return b
}

// Delete removes a item from the database based on its ID and user ID.
// It returns an error if the operation fails or if the item is not found.
func (r *TodoRepository) Delete(id int, userId int) error {
	query := `
	DELETE FROM todos WHERE id = $1 AND user_id = $2
	`
	result, err := r.DB.Exec(query, id, userId)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete todo rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}
