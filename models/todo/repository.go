package todo

import (
	"database/sql"
	"fmt"
)


const (
	ErrTodoNotFound string = "todo not found"
)

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
		return Todo{}, err
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
		return Todo{}, fmt.Errorf(ErrTodoNotFound)
	} else if err != nil {
		return Todo{}, err
	}
	return todo, nil
}


// GetAll retrieves all items for a specific user from the database.
// It returns a slice of items or an error if the operation fails.
func (r *TodoRepository) GetAll(userID int) ([]Todo, error) {
	query := `
		SELECT id, title, description, is_completed, user_id, created_at, updated_at
		FROM todos
		WHERE user_id = $1
	`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		todos = append(todos, todo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return todos, nil
}


// Update modifies an existing item in the database.
// Update can update the title, description or is_completed fields of the item.
// It takes a pointer to a struct as input and returns the updated item or an error if the operation fails.
func (r *TodoRepository) Update(id int, userId int, todo *Todo) (Todo, error) {
	var updatedTodo Todo

	query := `
	UPDATE todos SET title = $3, description = $4, is_completed = $5, updated_at = NOW()
	WHERE id = $1 AND user_id = $2
	RETURNING id, title, description, is_completed, user_id, created_at, updated_at
	`	

	err := r.DB.QueryRow(query, id, userId, todo.Title, todo.Description, todo.IsCompleted).
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
		return Todo{}, fmt.Errorf(ErrTodoNotFound)
	} else if err != nil {
		return Todo{}, err
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
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf(ErrTodoNotFound)
	}
	return nil
}