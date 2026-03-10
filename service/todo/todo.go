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

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

func NewTodoService(repo *todoModel.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

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

func (s *TodoService) GetAllTodos(userID int, options todoModel.TodoListOptions) ([]todoModel.Todo, error) {
	normalizedOptions, err := normalizeListOptions(options)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidListOptions, err)
	}

	todos, err := s.repo.GetAll(userID, normalizedOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTodoListFailed, err)
	}
	return todos, nil
}

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


func normalizeListOptions(options todoModel.TodoListOptions) (todoModel.TodoListOptions, error) {
	if options.SortBy == "" {
		options.SortBy = "created_at"
	}

	sortBy := strings.ToLower(strings.TrimSpace(options.SortBy))
	if sortBy != "created_at" && sortBy != "title" {
		return todoModel.TodoListOptions{}, fmt.Errorf("invalid sort_by. allowed values: created_at, title")
	}
	options.SortBy = sortBy

	order := strings.ToLower(strings.TrimSpace(options.Order))
	if order == "" {
		if options.SortBy == "title" {
			options.Order = "asc"
		} else {
			options.Order = "desc"
		}
	} else if order != "asc" && order != "desc" {
		return todoModel.TodoListOptions{}, fmt.Errorf("invalid order. allowed values: asc, desc")
	} else {
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