package routes

import (
	"github.com/francoganga/pagoda_bun/models"
	"github.com/francoganga/pagoda_bun/pkg/context"
	"github.com/francoganga/pagoda_bun/pkg/controller"
	"github.com/francoganga/pagoda_bun/pkg/msg"
	"github.com/labstack/echo/v4"
)

type verifyEmail struct {
	controller.Controller
}

func (c *verifyEmail) Get(ctx echo.Context) error {
	var usr *models.User

	// Validate the token
	token := ctx.Param("token")
	email, err := c.Container.Auth.ValidateEmailVerificationToken(token)
	if err != nil {
		msg.Warning(ctx, "The link is either invalid or has expired.")
		return c.Redirect(ctx, "home")
	}

	// Check if it matches the authenticated user
	if u := ctx.Get(context.AuthenticatedUserKey); u != nil {
		authUser := u.(*models.User)

		if authUser.Email == email {
			usr = authUser
		}
	}

	if usr == nil {
		usr := new(models.User)

		err := c.Container.Bun.NewSelect().
			Where("email = ?", email).
			Scan(ctx.Request().Context())

		if err != nil {
			return c.Fail(err, "query failed loading email verification token user")
		}

		if !usr.Verified {

			updateUser := &models.User{
				ID:       usr.ID,
				Verified: true,
			}

			_, err := c.Container.Bun.NewUpdate().
				Model(updateUser).
                Column("verified").
				WherePK().
				Exec(ctx.Request().Context())

			if err != nil {
				return c.Fail(err, "failed to set user as verified")
			}
		}
	}

	msg.Success(ctx, "Your email has been successfully verified.")
	return c.Redirect(ctx, "home")
}
