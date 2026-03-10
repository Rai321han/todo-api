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

func NewTodoService(repo *todoModel.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) AddTodo(todo *todoModel.Todo) error {
	if err := validateTodoInput(todo); err != nil {
		return err
	}

	createdTodo, err := s.repo.Create(todo)
	if err != nil {
		return err
	}

	*todo = createdTodo
	return nil
}

func (s *TodoService) GetTodoByID(id, userID int) (todoModel.Todo, error) {
	return s.repo.GetByID(id, userID)
}

func (s *TodoService) GetAllTodos(userID int) ([]todoModel.Todo, error) {
	return s.repo.GetAll(userID)
}

func (s *TodoService) UpdateTodo(id, userID int, todo *todoModel.Todo) (todoModel.Todo, error) {
	if err := validateTodoInput(todo); err != nil {
		return todoModel.Todo{}, err
	}

	updatedTodo, err := s.repo.Update(id, userID, todo)
	if err != nil {
		if errors.Is(err, todoModel.ErrTodoNotFound) {
			return todoModel.Todo{}, todoModel.ErrTodoNotFound
		}
		return todoModel.Todo{}, err
	}

	return updatedTodo, nil
}

func (s *TodoService) DeleteTodo(id, userID int) error {
	err := s.repo.Delete(id, userID)
	if err != nil {
		if errors.Is(err, todoModel.ErrTodoNotFound) {
			return todoModel.ErrTodoNotFound
		}
		return err
	}

	return nil
}

func validateTodoInput(todo *todoModel.Todo) error {
	if todo == nil {
		return fmt.Errorf("Not enough data.")
	}

	todo.Title = strings.TrimSpace(todo.Title)
	if todo.Title == "" {
		return fmt.Errorf("title is required")
	}

	return nil
}