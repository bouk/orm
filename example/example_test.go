package example

import (
	"testing"

	"github.com/stretchr/testify/require"

	"bou.ke/orm/example/db"
)

var _ db.UserRelation = db.Users

func TestEmptyUser(t *testing.T) {
	user, err := db.Users.First(ctx, d)

	require.EqualError(t, err, "not found")
	require.Nil(t, user)
}

func TestCreateUser(t *testing.T) {
	u := db.Users.New()
	err := u.Save(ctx, d)
	require.NoError(t, err)
	require.NotZero(t, u.ID)
}

func TestCreatePostUnderUser(t *testing.T) {
	u := db.Users.New()
	err := u.Save(ctx, d)
	require.NoError(t, err)
	require.NotZero(t, u.ID)

	err = u.Posts().New().Save(ctx, d)
	require.NoError(t, err)

	u, err = db.Users.Last(ctx, d)
	require.NoError(t, err)
	c, err := u.Posts().Count(ctx, d)
	require.NoError(t, err)
	require.Equal(t, int64(1), c)
}
