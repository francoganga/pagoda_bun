package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/francoganga/pagoda_bun/models"
	"github.com/francoganga/pagoda_bun/pkg/context"
	"github.com/uptrace/bun"

	"github.com/labstack/echo/v4"
)

// LoadUser loads the user based on the ID provided as a path parameter
func LoadUser(bun *bun.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := strconv.Atoi(c.Param("user"))
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound)
			}

			u := new(models.User)

			err = bun.NewSelect().
				Model(u).
				Where("id = ?", userID).
				Scan(c.Request().Context())

			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusNotFound)
			}

			if err == nil {
				c.Set(context.UserKey, u)
				return next(c)
			}

			return echo.NewHTTPError(
				http.StatusInternalServerError,
				fmt.Sprintf("error querying user: %v", err),
			)
		}
	}
}
