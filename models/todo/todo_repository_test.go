package todo

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
)

func setup() (*TodoRepository, sqlmock.Sqlmock, func()) {
	db, mock, _ := sqlmock.New()
	repo := &TodoRepository{DB: db}

	cleanup := func() {
		db.Close()
	}

	return repo, mock, cleanup
}

func TestTodoRepository_Create(t *testing.T) {
	Convey("TodoRepository Create", t, func() {
		repo, mock, cleanup := setup()
		defer cleanup()

		todoInput := &Todo{
			Title:       "Write tests",
			Description: "Add unit test for create todo",
			IsCompleted: false,
			UserID:      7,
		}

		Convey("Success: should create todo and return populated fields", func() {
			now := time.Now()

			mock.ExpectQuery("INSERT INTO todos").
				WithArgs(todoInput.Title, todoInput.Description, todoInput.IsCompleted, todoInput.UserID).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
						AddRow(11, now, now),
				)

			result, err := repo.Create(todoInput)

			So(err, ShouldBeNil)
			So(result.ID, ShouldEqual, 11)
			So(result.Title, ShouldEqual, todoInput.Title)
			So(result.Description, ShouldEqual, todoInput.Description)
			So(result.IsCompleted, ShouldEqual, todoInput.IsCompleted)
			So(result.UserID, ShouldEqual, todoInput.UserID)
			So(result.CreatedAt, ShouldResemble, now)
			So(result.UpdatedAt, ShouldResemble, now)

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should wrap DB error", func() {
			mock.ExpectQuery("INSERT INTO todos").
				WithArgs(todoInput.Title, todoInput.Description, todoInput.IsCompleted, todoInput.UserID).
				WillReturnError(errors.New("insert failed"))

			result, err := repo.Create(todoInput)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create todo")
			So(err.Error(), ShouldContainSubstring, "insert failed")
			So(result, ShouldResemble, Todo{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should return error if scan fails", func() {
			mock.ExpectQuery("INSERT INTO todos").
				WithArgs(todoInput.Title, todoInput.Description, todoInput.IsCompleted, todoInput.UserID).
				WillReturnRows(
					sqlmock.NewRows([]string{"id"}). // missing columns -> scan error
										AddRow(11),
				)

			result, err := repo.Create(todoInput)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "create todo")
			So(result, ShouldResemble, Todo{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

func TestTodoRepository_GetByID(t *testing.T) {
	Convey("TodoRepository GetByID", t, func() {
		repo, mock, cleanup := setup()
		defer cleanup()

		userId := 11
		todoId := 1

		now := time.Now()

		todoInput := &Todo{
			ID:          todoId,
			Title:       "Write tests",
			Description: "Add unit test for create todo",
			IsCompleted: false,
			UserID:      userId,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		Convey("SUCCESS: should return todo", func() {
			mock.ExpectQuery("SELECT id, title, description").
				WithArgs(todoId, userId).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "title", "description", "is_completed", "user_id", "created_at", "updated_at"}).
						AddRow(todoInput.ID, todoInput.Title, todoInput.Description, todoInput.IsCompleted, todoInput.UserID, todoInput.CreatedAt, todoInput.UpdatedAt),
				)

			result, err := repo.GetByID(todoId, userId)
			So(err, ShouldBeNil)
			So(result, ShouldResemble, *todoInput)
			So(result.ID, ShouldEqual, todoId)
			So(result.UserID, ShouldEqual, userId)

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("FAILURE: should return ErrTodoNotFound if no rows", func() {
			mock.ExpectQuery("SELECT id, title, description").
				WithArgs(todoId, userId).
				WillReturnError(sql.ErrNoRows)

			result, err := repo.GetByID(todoId, userId)
			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
			So(result, ShouldResemble, Todo{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("FAILURE: should wrap DB error", func() {
			mock.ExpectQuery("SELECT id, title, description").
				WithArgs(todoId, userId).
				WillReturnError(errors.New("query failed"))

			result, err := repo.GetByID(todoId, userId)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "get todo by id")
			So(err.Error(), ShouldContainSubstring, "query failed")
			So(result, ShouldResemble, Todo{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

func TestTodoRepository_GetAll(t *testing.T) {
	Convey("TodoRepository GetAll", t, func() {
		repo, mock, cleanup := setup()
		defer cleanup()

		userID := 7
		status := false
		options := TodoListOptions{
			Status: &status,
			SortBy: "title",
			Order:  "desc",
			Search: "Write",
			Page:   1,
			Limit:  2,
		}

		now := time.Now()

		Convey("Success: should return paginated todo list", func() {
			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID, status, "%Write%").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

			mock.ExpectQuery("SELECT id, title, description").
				WithArgs(userID, status, "%Write%", options.Limit, options.Offset()).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "title", "description", "is_completed", "user_id", "created_at", "updated_at"}).
						AddRow(1, "Write tests", "unit tests", false, userID, now, now).
						AddRow(2, "Write docs", "api docs", false, userID, now, now),
				)

			result, err := repo.GetAll(userID, options)

			So(err, ShouldBeNil)
			So(result.CurrentPage, ShouldEqual, options.Page)
			So(result.Limit, ShouldEqual, options.Limit)
			So(result.TotalCount, ShouldEqual, 5)
			So(result.TotalPages, ShouldEqual, 2)
			So(len(result.Todos), ShouldEqual, 2)
			So(result.Todos[0].Title, ShouldEqual, "Write tests")

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should return error when counting all todos fails", func() {
			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID).
				WillReturnError(errors.New("count failed"))

			result, err := repo.GetAll(userID, options)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "count todos")
			So(err.Error(), ShouldContainSubstring, "count failed")
			So(result, ShouldResemble, TodoListResponse{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should return error when counting filtered todos fails", func() {
			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID, status, "%Write%").
				WillReturnError(errors.New("filtered count failed"))

			result, err := repo.GetAll(userID, options)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "count filtered todos")
			So(err.Error(), ShouldContainSubstring, "filtered count failed")
			So(result, ShouldResemble, TodoListResponse{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should wrap query error when listing todos fails", func() {
			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID, status, "%Write%").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

			mock.ExpectQuery("SELECT id, title, description").
				WithArgs(userID, status, "%Write%", options.Limit, options.Offset()).
				WillReturnError(errors.New("list failed"))

			result, err := repo.GetAll(userID, options)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "list todos")
			So(err.Error(), ShouldContainSubstring, "list failed")
			So(result, ShouldResemble, TodoListResponse{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should return scan error when todo rows are invalid", func() {
			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

			mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
				WithArgs(userID, status, "%Write%").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

			mock.ExpectQuery("SELECT id, title, description").
				WithArgs(userID, status, "%Write%", options.Limit, options.Offset()).
				WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow(1),
				)

			result, err := repo.GetAll(userID, options)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "scan todo row")
			So(result, ShouldResemble, TodoListResponse{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

func TestTodoRepository_Update(t *testing.T) {
	Convey("TodoRepository Update", t, func() {
		repo, mock, cleanup := setup()
		defer cleanup()

		id := 10
		userID := 7
		now := time.Now()

		Convey("Success: should update and return todo", func() {
			input := &Todo{
				Title:       "Updated title",
				Description: "Updated description",
				IsCompleted: true,
			}

			mock.ExpectQuery("UPDATE todos SET").
				WithArgs(id, userID, input.Title, input.Description, input.IsCompleted).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "title", "description", "is_completed", "user_id", "created_at", "updated_at"}).
						AddRow(id, input.Title, input.Description, input.IsCompleted, userID, now, now),
				)

			result, err := repo.Update(id, userID, input)

			So(err, ShouldBeNil)
			So(result.ID, ShouldEqual, id)
			So(result.UserID, ShouldEqual, userID)
			So(result.Title, ShouldEqual, input.Title)
			So(result.Description, ShouldEqual, input.Description)
			So(result.IsCompleted, ShouldEqual, true)

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Success: should pass nil for empty title and description", func() {
			input := &Todo{IsCompleted: true}

			mock.ExpectQuery("UPDATE todos SET").
				WithArgs(id, userID, nil, nil, input.IsCompleted).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "title", "description", "is_completed", "user_id", "created_at", "updated_at"}).
						AddRow(id, "Existing title", "Existing description", true, userID, now, now),
				)

			result, err := repo.Update(id, userID, input)

			So(err, ShouldBeNil)
			So(result.ID, ShouldEqual, id)
			So(result.IsCompleted, ShouldBeTrue)

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should return ErrTodoNotFound when no rows are updated", func() {
			input := &Todo{Title: "No todo", IsCompleted: false}

			mock.ExpectQuery("UPDATE todos SET").
				WithArgs(id, userID, input.Title, nil, input.IsCompleted).
				WillReturnError(sql.ErrNoRows)

			result, err := repo.Update(id, userID, input)

			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
			So(result, ShouldResemble, Todo{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should wrap DB error", func() {
			input := &Todo{Title: "Title", Description: "Desc", IsCompleted: false}

			mock.ExpectQuery("UPDATE todos SET").
				WithArgs(id, userID, input.Title, input.Description, input.IsCompleted).
				WillReturnError(errors.New("update failed"))

			result, err := repo.Update(id, userID, input)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "update todo")
			So(err.Error(), ShouldContainSubstring, "update failed")
			So(result, ShouldResemble, Todo{})

			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}

func TestTodoRepository_Delete(t *testing.T) {
	Convey("TodoRepository Delete", t, func() {
		repo, mock, cleanup := setup()
		defer cleanup()

		id := 9
		userID := 3

		Convey("Success: should delete todo", func() {
			mock.ExpectExec("DELETE FROM todos").
				WithArgs(id, userID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.Delete(id, userID)

			So(err, ShouldBeNil)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should wrap DB exec error", func() {
			mock.ExpectExec("DELETE FROM todos").
				WithArgs(id, userID).
				WillReturnError(errors.New("delete failed"))

			err := repo.Delete(id, userID)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "delete todo")
			So(err.Error(), ShouldContainSubstring, "delete failed")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should wrap rows affected error", func() {
			mock.ExpectExec("DELETE FROM todos").
				WithArgs(id, userID).
				WillReturnResult(sqlmock.NewErrorResult(errors.New("rows failed")))

			err := repo.Delete(id, userID)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "delete todo rows affected")
			So(err.Error(), ShouldContainSubstring, "rows failed")
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})

		Convey("Failure: should return ErrTodoNotFound when no rows are deleted", func() {
			mock.ExpectExec("DELETE FROM todos").
				WithArgs(id, userID).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.Delete(id, userID)

			So(err, ShouldNotBeNil)
			So(errors.Is(err, ErrTodoNotFound), ShouldBeTrue)
			So(mock.ExpectationsWereMet(), ShouldBeNil)
		})
	})
}
