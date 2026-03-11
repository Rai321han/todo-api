package todo

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"

	todoModel "todo-api/models/todo"
)

func TestTodoService_AddTodo(t *testing.T) {
	Convey("AddTodo validates input before calling repository", t, func() {
		svc := NewTodoService(&todoModel.TodoRepository{})

		result, err := svc.AddTodo(nil, 7)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
		So(err.Error(), ShouldEqual, "invalid todo input: not enough data")
		So(result.ID, ShouldEqual, 0)
	})

	Convey("AddTodo trims title and delegates create", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Create", func(_ *todoModel.TodoRepository, input *todoModel.Todo) (todoModel.Todo, error) {
			So(input.Title, ShouldEqual, "Buy milk")
			So(input.Description, ShouldEqual, "2 liters")
			So(input.IsCompleted, ShouldBeFalse)
			So(input.UserID, ShouldEqual, 9)
			return todoModel.Todo{ID: 11, Title: input.Title, Description: input.Description, IsCompleted: input.IsCompleted, UserID: input.UserID}, nil
		})
		defer patches.Reset()

		result, err := svc.AddTodo(&todoModel.Todo{Title: "  Buy milk  ", Description: "2 liters"}, 9)

		So(err, ShouldBeNil)
		So(result.ID, ShouldEqual, 11)
		So(result.Title, ShouldEqual, "Buy milk")
		So(result.UserID, ShouldEqual, 9)
	})

	Convey("AddTodo returns repository error", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)
		expectedErr := errors.New("db create failed")

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Create", func(_ *todoModel.TodoRepository, _ *todoModel.Todo) (todoModel.Todo, error) {
			return todoModel.Todo{}, expectedErr
		})
		defer patches.Reset()

		_, err := svc.AddTodo(&todoModel.Todo{Title: "Task"}, 3)

		So(errors.Is(err, ErrTodoCreateFailed), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, expectedErr.Error())
	})
}

func TestTodoService_GetTodoByID(t *testing.T) {
	Convey("GetTodoByID returns todo from repository", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "GetByID", func(_ *todoModel.TodoRepository, id, userID int) (todoModel.Todo, error) {
			So(id, ShouldEqual, 4)
			So(userID, ShouldEqual, 2)
			return todoModel.Todo{ID: 4, Title: "Read"}, nil
		})
		defer patches.Reset()

		got, err := svc.GetTodoByID(4, 2)

		So(err, ShouldBeNil)
		So(got.ID, ShouldEqual, 4)
		So(got.Title, ShouldEqual, "Read")
	})

	Convey("GetTodoByID propagates repository error", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)
		expectedErr := errors.New("read failed")

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "GetByID", func(_ *todoModel.TodoRepository, _, _ int) (todoModel.Todo, error) {
			return todoModel.Todo{}, expectedErr
		})
		defer patches.Reset()

		_, err := svc.GetTodoByID(1, 1)

		So(errors.Is(err, ErrTodoFetchFailed), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, expectedErr.Error())
	})

	Convey("GetTodoByID maps not found errors", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "GetByID", func(_ *todoModel.TodoRepository, _, _ int) (todoModel.Todo, error) {
			return todoModel.Todo{}, todoModel.ErrTodoNotFound
		})
		defer patches.Reset()

		_, err := svc.GetTodoByID(1, 1)

		So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
	})
}

func TestTodoService_GetAllTodos(t *testing.T) {
	Convey("GetAllTodos normalizes default options and calls repository", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "GetAll", func(_ *todoModel.TodoRepository, userID int, options todoModel.TodoListOptions) (todoModel.TodoListResponse, error) {
			So(userID, ShouldEqual, 6)
			So(options.SortBy, ShouldEqual, "created_at")
			So(options.Order, ShouldEqual, "desc")
			So(options.Page, ShouldEqual, defaultPage)
			So(options.Limit, ShouldEqual, defaultLimit)
			So(options.Search, ShouldEqual, "")
			return todoModel.TodoListResponse{Todos: []todoModel.Todo{{ID: 1, Title: "One"}}}, nil
		})
		defer patches.Reset()

		result, err := svc.GetAllTodos(6, todoModel.TodoListOptions{})

		So(err, ShouldBeNil)
		So(len(result.Todos), ShouldEqual, 1)
		So(result.Todos[0].ID, ShouldEqual, 1)
	})

	Convey("GetAllTodos applies title sorting defaults", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "GetAll", func(_ *todoModel.TodoRepository, _ int, options todoModel.TodoListOptions) (todoModel.TodoListResponse, error) {
			So(options.SortBy, ShouldEqual, "title")
			So(options.Order, ShouldEqual, "asc")
			So(options.Search, ShouldEqual, "hello")
			return todoModel.TodoListResponse{}, nil
		})
		defer patches.Reset()

		_, err := svc.GetAllTodos(1, todoModel.TodoListOptions{SortBy: " Title ", Search: "  hello  "})

		So(err, ShouldBeNil)
	})

	Convey("GetAllTodos wraps invalid options with ErrInvalidListOptions", t, func() {
		svc := NewTodoService(&todoModel.TodoRepository{})

		_, err := svc.GetAllTodos(1, todoModel.TodoListOptions{SortBy: "priority"})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidListOptions), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "invalid sort_by")
	})

	Convey("GetAllTodos propagates repository error", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)
		expectedErr := errors.New("query failed")

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "GetAll", func(_ *todoModel.TodoRepository, _ int, _ todoModel.TodoListOptions) (todoModel.TodoListResponse, error) {
			return todoModel.TodoListResponse{}, expectedErr
		})
		defer patches.Reset()

		_, err := svc.GetAllTodos(10, todoModel.TodoListOptions{})

		So(errors.Is(err, ErrTodoListFailed), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, expectedErr.Error())
	})
}

func TestTodoService_UpdateTodo(t *testing.T) {
	Convey("UpdateTodo validates input", t, func() {
		svc := NewTodoService(&todoModel.TodoRepository{})

		_, err := svc.UpdateTodo(1, 1, &todoModel.Todo{Title: "   "})

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
		So(err.Error(), ShouldEqual, "invalid todo input: title is required")
	})

	Convey("UpdateTodo calls repository with trimmed title", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Update", func(_ *todoModel.TodoRepository, id, userID int, td *todoModel.Todo) (todoModel.Todo, error) {
			So(id, ShouldEqual, 8)
			So(userID, ShouldEqual, 7)
			So(td.Title, ShouldEqual, "Refined title")
			return todoModel.Todo{ID: 8, Title: td.Title, UserID: 7}, nil
		})
		defer patches.Reset()

		updated, err := svc.UpdateTodo(8, 7, &todoModel.Todo{Title: "  Refined title "})

		So(err, ShouldBeNil)
		So(updated.ID, ShouldEqual, 8)
		So(updated.Title, ShouldEqual, "Refined title")
	})

	Convey("UpdateTodo propagates repository error", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)
		expectedErr := errors.New("update failed")

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Update", func(_ *todoModel.TodoRepository, _, _ int, _ *todoModel.Todo) (todoModel.Todo, error) {
			return todoModel.Todo{}, expectedErr
		})
		defer patches.Reset()

		_, err := svc.UpdateTodo(1, 2, &todoModel.Todo{Title: "Task"})

		So(errors.Is(err, ErrTodoUpdateFailed), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, expectedErr.Error())
	})

	Convey("UpdateTodo maps not found errors", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Update", func(_ *todoModel.TodoRepository, _, _ int, _ *todoModel.Todo) (todoModel.Todo, error) {
			return todoModel.Todo{}, todoModel.ErrTodoNotFound
		})
		defer patches.Reset()

		_, err := svc.UpdateTodo(1, 2, &todoModel.Todo{Title: "Task"})

		So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
	})
}

func TestTodoService_DeleteTodo(t *testing.T) {
	Convey("DeleteTodo returns nil on success", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Delete", func(_ *todoModel.TodoRepository, id, userID int) error {
			So(id, ShouldEqual, 12)
			So(userID, ShouldEqual, 3)
			return nil
		})
		defer patches.Reset()

		err := svc.DeleteTodo(12, 3)

		So(err, ShouldBeNil)
	})

	Convey("DeleteTodo maps repository errors", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Delete", func(_ *todoModel.TodoRepository, _, _ int) error {
			return errors.New("delete failed")
		})
		defer patches.Reset()

		err := svc.DeleteTodo(12, 3)

		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrTodoDeleteFailed), ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "delete failed")
	})

	Convey("DeleteTodo maps not found errors", t, func() {
		repo := &todoModel.TodoRepository{}
		svc := NewTodoService(repo)

		patches := gomonkey.ApplyMethod(reflect.TypeOf(repo), "Delete", func(_ *todoModel.TodoRepository, _, _ int) error {
			return todoModel.ErrTodoNotFound
		})
		defer patches.Reset()

		err := svc.DeleteTodo(12, 3)

		So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
	})
}

func TestValidateTodoInput(t *testing.T) {
	Convey("validateTodoInput rejects nil", t, func() {
		err := validateTodoInput(nil)
		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
		So(err.Error(), ShouldEqual, "invalid todo input: not enough data")
	})

	Convey("validateTodoInput rejects empty title", t, func() {
		err := validateTodoInput(&todoModel.Todo{Title: "  "})
		So(err, ShouldNotBeNil)
		So(errors.Is(err, ErrInvalidTodoInput), ShouldBeTrue)
		So(err.Error(), ShouldEqual, "invalid todo input: title is required")
	})

	Convey("validateTodoInput trims valid title", t, func() {
		td := &todoModel.Todo{Title: "  task  "}
		err := validateTodoInput(td)
		So(err, ShouldBeNil)
		So(td.Title, ShouldEqual, "task")
	})
}

func TestNormalizeListOptions(t *testing.T) {
	Convey("normalizeListOptions applies defaults", t, func() {
		options, err := normalizeListOptions(todoModel.TodoListOptions{})

		So(err, ShouldBeNil)
		So(options.SortBy, ShouldEqual, "created_at")
		So(options.Order, ShouldEqual, "desc")
		So(options.Page, ShouldEqual, defaultPage)
		So(options.Limit, ShouldEqual, defaultLimit)
	})

	Convey("normalizeListOptions title sort defaults to asc", t, func() {
		options, err := normalizeListOptions(todoModel.TodoListOptions{SortBy: "TITLE", Search: "  key  "})

		So(err, ShouldBeNil)
		So(options.SortBy, ShouldEqual, "title")
		So(options.Order, ShouldEqual, "asc")
		So(options.Search, ShouldEqual, "key")
	})

	Convey("normalizeListOptions rejects invalid sort_by", t, func() {
		_, err := normalizeListOptions(todoModel.TodoListOptions{SortBy: "status"})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "invalid sort_by")
	})

	Convey("normalizeListOptions rejects invalid page", t, func() {
		_, err := normalizeListOptions(todoModel.TodoListOptions{Page: -1, Limit: 10})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "page must be")
	})

	Convey("normalizeListOptions rejects invalid limit", t, func() {
		_, err := normalizeListOptions(todoModel.TodoListOptions{Page: 1, Limit: maxLimit + 1})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "limit must be between")
	})
}

