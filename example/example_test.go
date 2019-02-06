package example

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"bou.ke/ctxdb"
	"bou.ke/orm/example/db"
)

func TestEmptyUser(t *testing.T) {
	user, err := db.Users().First(ctx)

	require.NoError(t, err)
	require.Nil(t, user)
}

func TestCreateUser(t *testing.T) {
	ctxdb.Tx(ctx, func(ctx context.Context) error {
		u := db.Users().New()
		err := u.Save(ctx)
		require.NoError(t, err)
		require.NotZero(t, u.ID)

		return ctxdb.Rollback
	})
}

func TestCreatePostUnderUser(t *testing.T) {
	u := db.Users().New()
	err := u.Save(ctx)
	require.NoError(t, err)
	require.NotZero(t, u.ID)

	err = u.Posts().New().Save(ctx)
	require.NoError(t, err)

	u, err = db.Users().Last(ctx)
	require.NoError(t, err)
	c, err := u.Posts().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), c)
}
