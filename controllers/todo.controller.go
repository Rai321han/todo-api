package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"todo-api/models/db"
	"todo-api/models/todo"
	service "todo-api/service/todo"
	"todo-api/utils"

	beego "github.com/beego/beego/v2/server/web"
)

type TodoController struct {
	beego.Controller
}

func (c *TodoController) todoService() *service.TodoService {
	todoRepo := &todo.TodoRepository{DB: db.DB}
	return service.NewTodoService(todoRepo)
}

func (c *TodoController) currentUserID() int {
	return c.Ctx.Input.GetData("user_id").(int)
}

func (c *TodoController) parseTodoID() (int, error) {
	id, err := c.GetInt(":id")
	if err != nil {
		return 0, fmt.Errorf("invalid id")
	}

	return id, nil
}

func (c *TodoController) handleTodoError(action string, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, service.ErrInvalidTodoInput),
		errors.Is(err, service.ErrInvalidTodoID),
		errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidListOptions):
		utils.RespondWithError(c.Ctx, 400, err.Error())
	case errors.Is(err, service.ErrTodoNotFound):
		utils.RespondWithError(c.Ctx, 404, "todo item not found")
	default:
		log.Printf("todo %s failed: %v", action, err)
		utils.RespondWithError(c.Ctx, 500, fmt.Sprintf("failed to %s", action))
	}
}

func decodeJSONBody(body io.Reader, dst interface{}) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("invalid request body: request body must contain a single JSON object")
	}

	return nil
}

// Create handles the HTTP POST request to create a new todo item. 
// It parses the request body to extract the todo details, validates the input, and then calls the service layer to add the new todo item to the database.
// If successful, it returns the created todo item with a 201 status code.
// If there are any errors during parsing, validation, or creation, it responds with appropriate error messages and status codes.
func (c *TodoController) Create() {
	todoService := c.todoService()
	var newTodo todo.Todo

	if err := decodeJSONBody(c.Ctx.Request.Body, &newTodo); err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}

	userId := c.currentUserID()
	createdTodo, err := todoService.AddTodo(&newTodo, userId)
	if err != nil {
		c.handleTodoError("create todo item", err)
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = createdTodo
	c.ServeJSON()
}


// GetByID handles the HTTP GET request to retrieve a specific todo item by its ID.
// It extracts the ID from the URL path, validates it, and then calls the service layer to fetch the corresponding todo item from the database.
// If the item is found and accessible by the user, it returns the item with a 200 status code.
// If the item is not found or there are any errors during retrieval, it responds with appropriate error messages and status codes.
func (c *TodoController) GetByID() {
	todoService := c.todoService()
	userId := c.currentUserID()
	id, err := c.parseTodoID()
	if err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}

	item, err := todoService.GetTodoByID(id, userId)
	if err != nil {
		c.handleTodoError("retrieve todo item", err)
		return
	}

	c.Data["json"] = item
	c.ServeJSON()
}


// GetAll handles the HTTP GET request to retrieve all todo items for the authenticated user.
// If successful, it returns a list of todo items with a 200 status code.
// If there are any errors during retrieval or if no items are found, it responds with appropriate error messages and status codes.
func (c *TodoController) GetAll() {
	todoService := c.todoService()
	userId := c.currentUserID()

	options, err := parseTodoListOptions(c)
	if err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}

	todos, err := todoService.GetAllTodos(userId, options)
	if err != nil {
		c.handleTodoError("retrieve todo items", err)
		return
	}

	c.Data["json"] = todos
	c.ServeJSON()
}


// parseTodoListOptions extracts and validates query parameters from the HTTP request to construct a TodoListOptions struct.
// It handles parameters for filtering by status, sorting, pagination, and search.
// If any parameters are invalid, it returns an error with a descriptive message.
func parseTodoListOptions(c *TodoController) (todo.TodoListOptions, error) {
	options := todo.TodoListOptions{
		SortBy: c.GetString("sort_by"),
		Order:  c.GetString("order"),
		Search: c.GetString("search"),
	}

	statusParam := strings.ToLower(strings.TrimSpace(c.GetString("status")))
	if statusParam != "" {
		switch statusParam {
		case "completed":
			completed := true
			options.Status = &completed
		case "pending":
			completed := false
			options.Status = &completed
		default:
			return todo.TodoListOptions{}, fmt.Errorf("invalid status. allowed values: completed, pending")
		}
	}

	pageParam := strings.TrimSpace(c.GetString("page"))
	if pageParam != "" {
		page, err := strconv.Atoi(pageParam)
		if err != nil {
			return todo.TodoListOptions{}, fmt.Errorf("page must be a valid integer")
		}
		options.Page = page
	}

	limitParam := strings.TrimSpace(c.GetString("limit"))
	if limitParam != "" {
		limit, err := strconv.Atoi(limitParam)
		if err != nil {
			return todo.TodoListOptions{}, fmt.Errorf("limit must be a valid integer")
		}
		options.Limit = limit
	}

	return options, nil
}


// Update handles the HTTP PUT request to update a specific todo item by its ID.
// It extracts the ID from the URL path, validates it, and then parses the request body to get the updated details of the todo item.
// If the update is successful, it returns the updated todo item with a 200 status code.
// If there are any errors during parsing, validation, or updating, it responds with appropriate error messages and status codes.
func (c *TodoController) Update() {
	todoService := c.todoService()
	userId := c.currentUserID()
	id, err := c.parseTodoID()
	if err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}

	var updatedData todo.Todo
	if err := decodeJSONBody(c.Ctx.Request.Body, &updatedData); err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}
	
	updatedTodo, err := todoService.UpdateTodo(id, userId, &updatedData)

	if err != nil {
		c.handleTodoError("update todo item", err)
		return
	}

	c.Data["json"] = updatedTodo
	c.ServeJSON()
}


// Delete handles the HTTP DELETE request to remove a specific todo item by its ID.
// If the deletion is successful, it returns a 204 No Content status code.
// If there are any errors during deletion or if the item is not found, it responds with appropriate error messages and status codes.
func (c *TodoController) Delete() {
	todoService := c.todoService()
	userId := c.currentUserID()
	id, err := c.parseTodoID()
	if err != nil {
		utils.RespondWithError(c.Ctx, 400, err.Error())
		return
	}

	err = todoService.DeleteTodo(id, userId)
	if err != nil {
		c.handleTodoError("delete todo item", err)
		return
	}

	c.Ctx.Output.SetStatus(204) // No Content
}