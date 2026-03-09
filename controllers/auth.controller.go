package controllers

import (
	"encoding/json"

	"todo-api/models/db"
	userModel "todo-api/models/user"
	"todo-api/service/auth"

	beego "github.com/beego/beego/v2/server/web"
)


type AuthController struct {
	beego.Controller
}

func (c *AuthController) Register() {
	var user userModel.User
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&user); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid request body"))
		return
	}

	repo := &userModel.UserRepository{DB: db.DB}
	
	authService := auth.NewAuthService(repo,"asdkhasdf")
	
	err := authService.Register(&user)

	if err != nil {
        c.Ctx.Output.SetStatus(400)
        c.Ctx.Output.Body([]byte(err.Error()))
        return
    }
    c.Ctx.Output.SetStatus(201)
    c.Data["json"] = map[string]string{"message": "user created"}
    c.ServeJSON()
}

func (c *AuthController) Login() {
    var input struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&input); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Ctx.Output.Body([]byte("invalid request body"))
		return
	}
	repo := &userModel.UserRepository{DB: db.DB}
	authService := auth.NewAuthService(repo,"asdkhasdf")
    token, err := authService.Login(input.Email, input.Password)
    if err != nil {
        c.Ctx.Output.SetStatus(401)
        c.Ctx.Output.Body([]byte(err.Error()))
        return
    }
    c.Data["json"] = map[string]string{"token": token}
    c.ServeJSON()
}
