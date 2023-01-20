package routes

import (
	"fmt"

	"github.com/francoganga/pagoda_bun/models"
	"github.com/francoganga/pagoda_bun/pkg/context"
	"github.com/francoganga/pagoda_bun/pkg/controller"
	"github.com/francoganga/pagoda_bun/pkg/msg"

	"github.com/labstack/echo/v4"
)

type (
	resetPassword struct {
		controller.Controller
	}

	resetPasswordForm struct {
		Password        string `form:"password" validate:"required"`
		ConfirmPassword string `form:"password-confirm" validate:"required,eqfield=Password"`
		Submission      controller.FormSubmission
	}
)

func (c *resetPassword) Get(ctx echo.Context) error {
	fmt.Println("LLEGA0")
	page := controller.NewPage(ctx)
	page.Layout = "auth"
	page.Name = "reset-password"
	page.Title = "Reset password"
	page.Form = resetPasswordForm{}

	fmt.Println("LLEGA1")

	if form := ctx.Get(context.FormKey); form != nil {
		page.Form = form.(*resetPasswordForm)
	}

	fmt.Println("LLEGA2")

	return c.RenderPage(ctx, page)
}

func (c *resetPassword) Post(ctx echo.Context) error {
    fmt.Println("LLEGA-1")
	var form resetPasswordForm
	ctx.Set(context.FormKey, &form)

	// Parse the form values
	if err := ctx.Bind(&form); err != nil {
		return c.Fail(err, "unable to parse password reset form")
	}

	if err := form.Submission.Process(ctx, form); err != nil {
		return c.Fail(err, "unable to process form submission")
	}

	if form.Submission.HasErrors() {
		return c.Get(ctx)
	}

	// Hash the new password
	hash, err := c.Container.Auth.HashPassword(form.Password)
	if err != nil {
		return c.Fail(err, "unable to hash password")
	}

	// Get the requesting user
	usr := ctx.Get(context.UserKey).(*models.User)

	// // Update the user
	// _, err = usr.
	// 	Update().
	// 	SetPassword(hash).
	// 	Save(ctx.Request().Context())

	_, err = c.Container.Bun.NewUpdate().
		Model(usr).
		Set("password = ?", hash).
        Where("id = ?", usr.ID).
		Exec(ctx.Request().Context())

	if err != nil {
		return c.Fail(err, "unable to update password")
	}

	// Delete all password tokens for this user
	err = c.Container.Auth.DeletePasswordTokens(ctx, usr.ID)
	if err != nil {
		return c.Fail(err, "unable to delete password tokens")
	}

	msg.Success(ctx, "Your password has been updated.")
	return c.Redirect(ctx, "login")
}
