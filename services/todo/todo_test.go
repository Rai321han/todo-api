package todo

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	todoModel "todo-api/models/todo"
)

// fakeTodoRepo is a test double for the TodoRepository interface that allows us to inject custom behavior for each method.
// We can set the function fields to simulate different scenarios and control the responses for our service tests.
type fakeTodoRepo struct {
	createFunc  func(todo *todoModel.Todo) (todoModel.Todo, error)
	getByIDFunc func(id, userID int) (todoModel.Todo, error)
	getAllFunc  func(userID int, options todoModel.TodoListOptions) (todoModel.TodoListResponse, error)
	updateFunc  func(id, userID int, todo *todoModel.Todo) (todoModel.Todo, error)
	deleteFunc  func(id, userID int) error
}

func (f *fakeTodoRepo) Create(todo *todoModel.Todo) (todoModel.Todo, error) {
	if f.createFunc != nil {
		return f.createFunc(todo)
	}
	return todoModel.Todo{}, nil
}

func (f *fakeTodoRepo) GetByID(id, userID int) (todoModel.Todo, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(id, userID)
	}
	return todoModel.Todo{}, nil
}

func (f *fakeTodoRepo) GetAll(userID int, options todoModel.TodoListOptions) (todoModel.TodoListResponse, error) {
	if f.getAllFunc != nil {
		return f.getAllFunc(userID, options)
	}
	return todoModel.TodoListResponse{}, nil
}

func (f *fakeTodoRepo) Update(id, userID int, todo *todoModel.Todo) (todoModel.Todo, error) {
	if f.updateFunc != nil {
		return f.updateFunc(id, userID, todo)
	}
	return todoModel.Todo{}, nil
}

func (f *fakeTodoRepo) Delete(id, userID int) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(id, userID)
	}
	return nil
}

func TestValidateTodoInput(t *testing.T) {
	Convey("validateTodoInput should reject nil payload", t, func() {
		err := validateTodoInput(nil)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
	})

	Convey("validateTodoInput should reject empty title", t, func() {
		err := validateTodoInput(&todoModel.Todo{Title: "   "})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
	})

	Convey("validateTodoInput should trim title and accept valid payload", t, func() {
		in := &todoModel.Todo{Title: "  Buy milk  "}

		err := validateTodoInput(in)

		So(err, ShouldBeNil)
		So(in.Title, ShouldEqual, "Buy milk")
	})
}

func TestNormalizeListOptions(t *testing.T) {
	Convey("normalizeListOptions should apply defaults", t, func() {
		normalized, err := normalizeListOptions(todoModel.TodoListOptions{})

		So(err, ShouldBeNil)
		So(normalized.SortBy, ShouldEqual, "created_at")
		So(normalized.Order, ShouldEqual, "desc")
		So(normalized.Page, ShouldEqual, defaultPage)
		So(normalized.Limit, ShouldEqual, defaultLimit)
	})

	Convey("normalizeListOptions should normalize and keep explicit title sort defaults", t, func() {
		normalized, err := normalizeListOptions(todoModel.TodoListOptions{SortBy: " TITLE "})

		So(err, ShouldBeNil)
		So(normalized.SortBy, ShouldEqual, "title")
		So(normalized.Order, ShouldEqual, "asc")
	})

	Convey("normalizeListOptions should reject invalid sort", t, func() {
		_, err := normalizeListOptions(todoModel.TodoListOptions{SortBy: "priority"})

		So(err, ShouldNotBeNil)
	})

	Convey("normalizeListOptions should reject invalid order", t, func() {
		_, err := normalizeListOptions(todoModel.TodoListOptions{SortBy: "title", Order: "up"})

		So(err, ShouldNotBeNil)
	})

	Convey("normalizeListOptions should reject invalid page and limit", t, func() {
		_, pageErr := normalizeListOptions(todoModel.TodoListOptions{Page: -1, Limit: 10})
		_, limitErr := normalizeListOptions(todoModel.TodoListOptions{Page: 1, Limit: 101})

		So(pageErr, ShouldNotBeNil)
		So(limitErr, ShouldNotBeNil)
	})

	Convey("normalizeListOptions should trim search", t, func() {
		normalized, err := normalizeListOptions(todoModel.TodoListOptions{Search: "  abc  "})

		So(err, ShouldBeNil)
		So(normalized.Search, ShouldEqual, "abc")
	})
}

func TestTodoServiceAddTodo(t *testing.T) {
	Convey("AddTodo should validate input", t, func() {
		svc := NewTodoService(&fakeTodoRepo{})

		_, err := svc.AddTodo(&todoModel.Todo{Title: "  "}, 1)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
	})

	Convey("AddTodo should create todo with user id", t, func() {
		repo := &fakeTodoRepo{
			createFunc: func(todo *todoModel.Todo) (todoModel.Todo, error) {
				So(todo.Title, ShouldEqual, "Task")
				So(todo.Description, ShouldEqual, "Desc")
				So(todo.UserID, ShouldEqual, 77)
				return todoModel.Todo{ID: 1, Title: todo.Title, Description: todo.Description, UserID: todo.UserID}, nil
			},
		}
		svc := NewTodoService(repo)

		created, err := svc.AddTodo(&todoModel.Todo{Title: "Task", Description: "Desc"}, 77)

		So(err, ShouldBeNil)
		So(created.ID, ShouldEqual, 1)
		So(created.UserID, ShouldEqual, 77)
	})

	Convey("AddTodo should wrap repository create errors", t, func() {
		repo := &fakeTodoRepo{createFunc: func(todo *todoModel.Todo) (todoModel.Todo, error) {
			return todoModel.Todo{}, errors.New("db down")
		}}
		svc := NewTodoService(repo)

		_, err := svc.AddTodo(&todoModel.Todo{Title: "Task"}, 1)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoCreateFailed), ShouldBeTrue)
	})
}

func TestTodoServiceGetTodoByID(t *testing.T) {
	Convey("GetTodoByID should map not found error", t, func() {
		repo := &fakeTodoRepo{getByIDFunc: func(id, userID int) (todoModel.Todo, error) {
			return todoModel.Todo{}, todoModel.ErrTodoNotFound
		}}
		svc := NewTodoService(repo)

		_, err := svc.GetTodoByID(1, 2)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
	})

	Convey("GetTodoByID should wrap unexpected errors", t, func() {
		repo := &fakeTodoRepo{getByIDFunc: func(id, userID int) (todoModel.Todo, error) {
			return todoModel.Todo{}, errors.New("db error")
		}}
		svc := NewTodoService(repo)

		_, err := svc.GetTodoByID(1, 2)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoFetchFailed), ShouldBeTrue)
	})

	Convey("GetTodoByID should return todo on success", t, func() {
		repo := &fakeTodoRepo{getByIDFunc: func(id, userID int) (todoModel.Todo, error) {
			return todoModel.Todo{ID: id, UserID: userID, Title: "Task"}, nil
		}}
		svc := NewTodoService(repo)

		got, err := svc.GetTodoByID(9, 3)

		So(err, ShouldBeNil)
		So(got.ID, ShouldEqual, 9)
		So(got.UserID, ShouldEqual, 3)
	})
}

func TestTodoServiceGetAllTodos(t *testing.T) {
	Convey("GetAllTodos should reject invalid list options", t, func() {
		svc := NewTodoService(&fakeTodoRepo{})

		_, err := svc.GetAllTodos(1, todoModel.TodoListOptions{SortBy: "invalid"})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidListOptions), ShouldBeTrue)
	})

	Convey("GetAllTodos should normalize options and return response", t, func() {
		repo := &fakeTodoRepo{getAllFunc: func(userID int, options todoModel.TodoListOptions) (todoModel.TodoListResponse, error) {
			So(userID, ShouldEqual, 5)
			So(options.SortBy, ShouldEqual, "title")
			So(options.Order, ShouldEqual, "asc")
			So(options.Page, ShouldEqual, defaultPage)
			So(options.Limit, ShouldEqual, defaultLimit)
			return todoModel.TodoListResponse{CurrentPage: options.Page, Limit: options.Limit, Todos: []todoModel.Todo{{ID: 1}}}, nil
		}}
		svc := NewTodoService(repo)

		resp, err := svc.GetAllTodos(5, todoModel.TodoListOptions{SortBy: "title"})

		So(err, ShouldBeNil)
		So(len(resp.Todos), ShouldEqual, 1)
		So(resp.Todos[0].ID, ShouldEqual, 1)
	})

	Convey("GetAllTodos should wrap repository errors", t, func() {
		repo := &fakeTodoRepo{getAllFunc: func(userID int, options todoModel.TodoListOptions) (todoModel.TodoListResponse, error) {
			return todoModel.TodoListResponse{}, errors.New("query error")
		}}
		svc := NewTodoService(repo)

		_, err := svc.GetAllTodos(1, todoModel.TodoListOptions{})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoListFailed), ShouldBeTrue)
	})
}

func TestTodoServiceUpdateTodo(t *testing.T) {
	Convey("UpdateTodo should validate input", t, func() {
		svc := NewTodoService(&fakeTodoRepo{})

		_, err := svc.UpdateTodo(1, 1, &todoModel.Todo{Title: "   "})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
	})

	Convey("UpdateTodo should map not found error", t, func() {
		repo := &fakeTodoRepo{updateFunc: func(id, userID int, todo *todoModel.Todo) (todoModel.Todo, error) {
			return todoModel.Todo{}, todoModel.ErrTodoNotFound
		}}
		svc := NewTodoService(repo)

		_, err := svc.UpdateTodo(1, 1, &todoModel.Todo{Title: "Task"})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
	})

	Convey("UpdateTodo should wrap unexpected errors", t, func() {
		repo := &fakeTodoRepo{updateFunc: func(id, userID int, todo *todoModel.Todo) (todoModel.Todo, error) {
			return todoModel.Todo{}, errors.New("update failed")
		}}
		svc := NewTodoService(repo)

		_, err := svc.UpdateTodo(1, 1, &todoModel.Todo{Title: "Task"})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoUpdateFailed), ShouldBeTrue)
	})

	Convey("UpdateTodo should return updated todo", t, func() {
		repo := &fakeTodoRepo{updateFunc: func(id, userID int, todo *todoModel.Todo) (todoModel.Todo, error) {
			return todoModel.Todo{ID: id, UserID: userID, Title: todo.Title}, nil
		}}
		svc := NewTodoService(repo)

		updated, err := svc.UpdateTodo(10, 2, &todoModel.Todo{Title: "Task"})

		So(err, ShouldBeNil)
		So(updated.ID, ShouldEqual, 10)
		So(updated.UserID, ShouldEqual, 2)
	})
}

func TestTodoServiceDeleteTodo(t *testing.T) {
	Convey("DeleteTodo should map not found error", t, func() {
		repo := &fakeTodoRepo{deleteFunc: func(id, userID int) error {
			return todoModel.ErrTodoNotFound
		}}
		svc := NewTodoService(repo)

		err := svc.DeleteTodo(1, 2)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
	})

	Convey("DeleteTodo should wrap unexpected errors", t, func() {
		repo := &fakeTodoRepo{deleteFunc: func(id, userID int) error {
			return errors.New("delete failed")
		}}
		svc := NewTodoService(repo)

		err := svc.DeleteTodo(1, 2)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoDeleteFailed), ShouldBeTrue)
	})

	Convey("DeleteTodo should return nil on success", t, func() {
		svc := NewTodoService(&fakeTodoRepo{})

		err := svc.DeleteTodo(1, 2)

		So(err, ShouldBeNil)
	})
}
