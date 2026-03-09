package controllers

import (
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
	if err := c.ParseForm(&newTodo); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid request body"))
		return
	}

	// Call the service layer to create the new todo item
	err := todoService.AddTodo(&newTodo)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Ctx.Output.Body([]byte("failed to create todo item"))
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = newTodo
	c.ServeJSON()
}

func (c *TodoController) GetByID() {
	// Implementation for retrieving a todo item by ID will go here
}

func (c *TodoController) GetAll() {
	// Implementation for retrieving all todo items will go here
}

func (c *TodoController) Update() {
	// Implementation for updating a todo item will go here
}

func (c *TodoController) Delete() {
	// Implementation for deleting a todo item will go here
}