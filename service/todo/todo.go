package todo

import (
	"errors"
	"fmt"
	"strings"

	todoModel "todo-api/models/todo"
)

type TodoService struct {
	repo *todoModel.TodoRepository
}


// Define custom error variables for better error handling and clarity
var (
	ErrInvalidTodoInput  = errors.New("invalid todo input")
	ErrInvalidTodoID     = errors.New("invalid todo id")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidListOptions = errors.New("invalid todo list options")
	ErrTodoNotFound      = errors.New("todo not found")
	ErrTodoCreateFailed  = errors.New("failed to create todo item")
	ErrTodoFetchFailed   = errors.New("failed to retrieve todo item")
	ErrTodoListFailed    = errors.New("failed to retrieve todo items")
	ErrTodoUpdateFailed  = errors.New("failed to update todo item")
	ErrTodoDeleteFailed  = errors.New("failed to delete todo item")
)

// Define constants for default pagination values and maximum limits
const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)


// NewTodoService creates a new instance of TodoService with the provided repository.
func NewTodoService(repo *todoModel.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}


// AddTodo validates the input and creates a new todo item for the specified user.
// It returns the created todo item or an error if the operation fails.
func (s *TodoService) AddTodo(todoData *todoModel.Todo, userId int) (todoModel.Todo, error) {
	if err := validateTodoInput(todoData); err != nil {
		return todoModel.Todo{}, err
	}

	todo := &todoModel.Todo{
		Title:       todoData.Title,
		Description: todoData.Description,
		IsCompleted: todoData.IsCompleted,
		UserID:      userId,
	}

	createdTodo, err := s.repo.Create(todo)
	if err != nil {
		return todoModel.Todo{}, fmt.Errorf("%w: %v", ErrTodoCreateFailed, err)
	}

	return createdTodo, nil
}

// GetTodoByID retrieves a todo item by its ID and the associated user ID.
// It returns the todo item or an error if the item is not found or if the operation fails.
func (s *TodoService) GetTodoByID(id, userID int) (todoModel.Todo, error) {
	todo, err := s.repo.GetByID(id, userID)
	if err != nil {
		if errors.Is(err, todoModel.ErrTodoNotFound) {
			return todoModel.Todo{}, ErrTodoNotFound
		}

		return todoModel.Todo{}, fmt.Errorf("%w: %v", ErrTodoFetchFailed, err)
	}
	return todo, nil
}

// GetAllTodos retrieves a list of todo items for the specified user ID based on the provided list options.
// It returns a slice of todo items or an error if the operation fails.
func (s *TodoService) GetAllTodos(userID int, options todoModel.TodoListOptions) (todoModel.TodoListResponse, error) {
	normalizedOptions, err := normalizeListOptions(options)
	if err != nil {
		return todoModel.TodoListResponse{}, fmt.Errorf("%w: %v", ErrInvalidListOptions, err)
	}

	todos, err := s.repo.GetAll(userID, normalizedOptions)
	if err != nil {
		return todoModel.TodoListResponse{}, fmt.Errorf("%w: %v", ErrTodoListFailed, err)
	}
	return todos, nil
}

// UpdateTodo validates the input and updates an existing todo item identified by its ID and associated user ID.
// It returns the updated todo item or an error if the item is not found or if the operation fails.
func (s *TodoService) UpdateTodo(id, userID int, todo *todoModel.Todo) (todoModel.Todo, error) {
	if err := validateTodoInput(todo); err != nil {
		return todoModel.Todo{}, err
	}

	updatedTodo, err := s.repo.Update(id, userID, todo)
	if err != nil {
		if errors.Is(err, todoModel.ErrTodoNotFound) {
			return todoModel.Todo{}, ErrTodoNotFound
		}

		return todoModel.Todo{}, fmt.Errorf("%w: %v", ErrTodoUpdateFailed, err)
	}
	return updatedTodo, nil
}

// DeleteTodo removes a todo item identified by its ID and associated user ID.
// It returns an error if the item is not found or if the operation fails.
func (s *TodoService) DeleteTodo(id, userID int) error {
	err := s.repo.Delete(id, userID)
	if err != nil {
		if errors.Is(err, todoModel.ErrTodoNotFound) {
			return ErrTodoNotFound
		}

		return fmt.Errorf("%w: %v", ErrTodoDeleteFailed, err)
	}
	return nil
}

// validateTodoInput checks if the provided todo item has valid data.
// It ensures that the title is not empty and trims any leading or trailing whitespace.
// It returns an error if the input is invalid.
func validateTodoInput(todo *todoModel.Todo) error {
	if todo == nil {
		return fmt.Errorf("%w: not enough data", ErrInvalidTodoInput)
	}

	todo.Title = strings.TrimSpace(todo.Title)
	if todo.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidTodoInput)
	}

	return nil
}


// normalizeListOptions processes and validates the provided list options for retrieving todo items.
// It sets default values for sorting and pagination if they are not provided and ensures that the values are valid.
// It returns the normalized list options or an error if the options are invalid.
// Example usage:
// options := todoModel.TodoListOptions{
//     SortBy: "title",
//     Order:  "asc",
//     Page:   1,
//     Limit:  20,
// }
// normalizedOptions, err := normalizeListOptions(options)
// if err != nil {
//     // Handle error
// }
func normalizeListOptions(options todoModel.TodoListOptions) (todoModel.TodoListOptions, error) {
	// Set default sorting field to "created_at" if not provided
	if options.SortBy == "" {
		options.SortBy = "created_at"
	}

	sortBy := strings.ToLower(strings.TrimSpace(options.SortBy))
	if sortBy != "created_at" && sortBy != "title" {
		return todoModel.TodoListOptions{}, fmt.Errorf("invalid sort_by. allowed values: created_at, title")
	}

	// Set the normalized sortBy back to options to ensure consistent formatting
	options.SortBy = sortBy

	order := strings.ToLower(strings.TrimSpace(options.Order))

	// Set the order based on the sort_by field if not provided, otherwise validate the provided order
	// Default to ascending order for title and descending order for created_at to show newest items first
	// This logic ensures that if the client does not specify an order, it will default to a sensible choice based on the sorting field.
	// If the client does specify an order, it must be either "asc" or "desc", otherwise an error is returned.
	if order == "" {
		if options.SortBy == "title" {
			options.Order = "asc"
		} else {
			// Default to descending order for created_at to show newest items first
			options.Order = "desc"
		}
	} else if order != "asc" && order != "desc" {
		return todoModel.TodoListOptions{}, fmt.Errorf("invalid order. allowed values: asc, desc")
	} else {
		// Set the normalized order back to options to ensure consistent formatting
		options.Order = order 
	}

	if options.Page == 0 {
		options.Page = defaultPage
	}
	if options.Limit == 0 {
		options.Limit = defaultLimit
	}

	if options.Page < 1 {
		return todoModel.TodoListOptions{}, fmt.Errorf("page must be greater than or equal to 1")
	}
	if options.Limit < 1 || options.Limit > maxLimit {
		return todoModel.TodoListOptions{}, fmt.Errorf("limit must be between 1 and %d", maxLimit)
	}

	options.Search = strings.TrimSpace(options.Search)

	return options, nil
}