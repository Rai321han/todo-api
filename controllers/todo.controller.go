package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"todo-api/models/db"
	"todo-api/models/todo"
	service "todo-api/service/todo"

	beego "github.com/beego/beego/v2/server/web"
)

type TodoController struct {
	beego.Controller
}

// Create handles the HTTP POST request to create a new todo item. 
// It parses the request body to extract the todo details, validates the input, and then calls the service layer to add the new todo item to the database.
// If successful, it returns the created todo item with a 201 status code.
// If there are any errors during parsing, validation, or creation, it responds with appropriate error messages and status codes.
func (c *TodoController) Create() {
	todoRepo := &todo.TodoRepository{DB: db.DB}
	todoService := service.NewTodoService(todoRepo)
	var newTodo todo.Todo

	// Decode json from the request body into the newTodo struct
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&newTodo); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid request body"))
		return
	}

	// Call the service layer to create the new todo item
	userId := c.Ctx.Input.GetData("user_id").(int)
	createdTodo, err := todoService.AddTodo(&newTodo, userId)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte("failed to create todo item"))
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
	todoRepo := &todo.TodoRepository{DB: db.DB}
	todoService := service.NewTodoService(todoRepo)


	userId := c.Ctx.Input.GetData("user_id").(int)

	// Extract the ID from the URL path
	id, err := c.GetInt(":id")
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid ID"))
		return
	}

	todo, err := todoService.GetTodoByID(id, userId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte("failed to retrieve todo item"))
		return
	} else if todo.ID == 0 {
		c.Ctx.Output.SetStatus(404)
		c.Ctx.Output.Body([]byte("todo item not found"))
		return
	}
	c.Data["json"] = todo
	c.ServeJSON()
}


// GetAll handles the HTTP GET request to retrieve all todo items for the authenticated user.
// If successful, it returns a list of todo items with a 200 status code.
// If there are any errors during retrieval or if no items are found, it responds with appropriate error messages and status codes.
func (c *TodoController) GetAll() {
	todoRepo := &todo.TodoRepository{DB: db.DB}
	todoService := service.NewTodoService(todoRepo)

	userId := c.Ctx.Input.GetData("user_id").(int)

	options, err := parseTodoListOptions(c)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte(err.Error()))
		return
	}

	todos, err := todoService.GetAllTodos(userId, options)
	if err != nil {
		if errors.Is(err, service.ErrInvalidListOptions) {
			c.Ctx.Output.SetStatus(400)
			c.Ctx.Output.Body([]byte(err.Error()))
			return
		}

		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte("failed to retrieve todo items"))
		return
	}

	c.Data["json"] = todos
	c.ServeJSON()
}

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
	todoRepo := &todo.TodoRepository{DB: db.DB}
	todoService := service.NewTodoService(todoRepo)

	userId := c.Ctx.Input.GetData("user_id").(int)

	// Extract the ID from the URL path
	id, err := c.GetInt(":id")
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid ID"))
		return
	}

	var updatedData todo.Todo
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&updatedData); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid request body"))
		return
	}
	
	updatedTodo, err := todoService.UpdateTodo(id, userId, &updatedData)

	if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Ctx.Output.Body([]byte("failed to update todo."))
	}

	if updatedTodo.ID == 0 {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte("todo item not found"))
		return
	}
	c.Data["json"] = updatedTodo
	c.ServeJSON()
}


// Delete handles the HTTP DELETE request to remove a specific todo item by its ID.
// If the deletion is successful, it returns a 204 No Content status code.
// If there are any errors during deletion or if the item is not found, it responds with appropriate error messages and status codes.
func (c *TodoController) Delete() {
	todoRepo := &todo.TodoRepository{DB: db.DB}
	todoService := service.NewTodoService(todoRepo)

	userId := c.Ctx.Input.GetData("user_id").(int)

	// Extract the ID from the URL path
	id, err := c.GetInt(":id")
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid ID"))
		return
	}

	err = todoService.DeleteTodo(id, userId)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte("failed to delete todo item"))
		return
	}

	c.Ctx.Output.SetStatus(204) // No Content
}