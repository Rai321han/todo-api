package todo

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrTodoNotFound = errors.New("todo not found")

// TodoRepository provides methods to interact with the todos table in the database.
// It contains a reference to the database connection and allows for creating, retrieving, updating, and deleting items.
type TodoRepository struct {
	DB *sql.DB
}

// Create inserts a new item into the database. 
// It takes a pointer to a todo struct as input and returns an error if the operation fails or the created Todo item with its ID populated if successful.
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
func (r *TodoRepository) GetAll(userID int, options TodoListOptions) ([]Todo, error) {
	baseQuery := `
		SELECT id, title, description, is_completed, user_id, created_at, updated_at
		FROM todos
		WHERE user_id = $1
	`

	args := []interface{}{userID}
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

	orderBy := "created_at DESC"
	if options.SortBy == "title" {
		if options.Order == "desc" {
			orderBy = "title DESC"
		} else {
			orderBy = "title ASC"
		}
	} else {
		if options.Order == "asc" {
			orderBy = "created_at ASC"
		} else {
			orderBy = "created_at DESC"
		}
	}

	baseQuery += " ORDER BY " + orderBy
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, options.Limit, options.Offset())

	rows, err := r.DB.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer rows.Close()

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
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todo rows: %w", err)
	}
	return todos, nil
}


func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullBool(b bool) interface{} {
	return b
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