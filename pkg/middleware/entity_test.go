package middleware

import (
	"fmt"
	"testing"

	"github.com/francoganga/pagoda_bun/models"
	"github.com/francoganga/pagoda_bun/pkg/context"
	"github.com/francoganga/pagoda_bun/pkg/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUser(t *testing.T) {
	ctx, _ := tests.NewContext(c.Web, "/")
	ctx.SetParamNames("user")
	ctx.SetParamValues(fmt.Sprintf("%d", usr.ID))
	_ = tests.ExecuteMiddleware(ctx, LoadUser(c.Bun))
	ctxUsr, ok := ctx.Get(context.UserKey).(*models.User)
	require.True(t, ok)
	assert.Equal(t, usr.ID, ctxUsr.ID)
}
