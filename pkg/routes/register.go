package routes

import (
	"errors"
	"fmt"

	"github.com/francoganga/pagoda_bun/models"
	"github.com/francoganga/pagoda_bun/pkg/context"
	"github.com/francoganga/pagoda_bun/pkg/controller"
	"github.com/francoganga/pagoda_bun/pkg/msg"
	"github.com/jackc/pgconn"

	"github.com/labstack/echo/v4"
)

type (
	register struct {
		controller.Controller
	}

	registerForm struct {
		Name            string `form:"name" validate:"required"`
		Email           string `form:"email" validate:"required,email"`
		Password        string `form:"password" validate:"required"`
		ConfirmPassword string `form:"password-confirm" validate:"required,eqfield=Password"`
		Submission      controller.FormSubmission
	}
)

func (c *register) Get(ctx echo.Context) error {
	page := controller.NewPage(ctx)
	page.Layout = "auth"
	page.Name = "register"
	page.Title = "Register"
	page.Form = registerForm{}

	if form := ctx.Get(context.FormKey); form != nil {
		page.Form = form.(*registerForm)
	}

	return c.RenderPage(ctx, page)
}

func (c *register) Post(ctx echo.Context) error {
	var form registerForm
	ctx.Set(context.FormKey, &form)

	// Parse the form values
	if err := ctx.Bind(&form); err != nil {
		return c.Fail(err, "unable to parse register form")
	}

	if err := form.Submission.Process(ctx, form); err != nil {
		return c.Fail(err, "unable to process form submission")
	}

	if form.Submission.HasErrors() {
		return c.Get(ctx)
	}

	// Hash the password
	pwHash, err := c.Container.Auth.HashPassword(form.Password)
	if err != nil {
		return c.Fail(err, "unable to hash password")
	}

	u := &models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: pwHash,
	}

	_, err = c.Container.Bun.NewInsert().
		Model(u).
		Exec(ctx.Request().Context())

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				msg.Warning(ctx, "A user with this email address already exists. Please log in.")
				return c.Redirect(ctx, "login")
			}
		}

		return c.Fail(err, "could not insert")
	}

	ctx.Logger().Infof("user created: %s", u.Name)

	// Log the user in
	err = c.Container.Auth.Login(ctx, u.ID)
	if err != nil {
		ctx.Logger().Errorf("unable to log in: %v", err)
		msg.Info(ctx, "Your account has been created.")
		return c.Redirect(ctx, "login")
	}

	msg.Success(ctx, "Your account has been created. You are now logged in.")

	// Send the verification email
	c.sendVerificationEmail(ctx, u)

	return c.Redirect(ctx, "home")
}

func (c *register) sendVerificationEmail(ctx echo.Context, usr *models.User) {
	// Generate a token
	token, err := c.Container.Auth.GenerateEmailVerificationToken(usr.Email)
	if err != nil {
		ctx.Logger().Errorf("unable to generate email verification token: %v", err)
		return
	}

	// Send the email
	url := ctx.Echo().Reverse("verify_email", token)
	err = c.Container.Mail.
		Compose().
		To(usr.Email).
		Subject("Confirm your email address").
		Body(fmt.Sprintf("Click here to confirm your email address: %s", url)).
		Send(ctx)

	if err != nil {
		ctx.Logger().Errorf("unable to send email verification link: %v", err)
		return
	}

	msg.Info(ctx, "An email was sent to you to verify your email address.")
}
