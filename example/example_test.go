package example

import (
	"testing"

	"github.com/stretchr/testify/require"

	"bou.ke/orm/example/db"
	"bou.ke/orm/rel"
)

var _ db.UserRelation = db.Users()

func TestEmptyUser(t *testing.T) {
	user, err := db.Users().First(ctx, d)

	require.EqualError(t, err, "not found")
	require.Nil(t, user)
}

func TestCreateUser(t *testing.T) {
	defer clear()
	createUser(t)
}

func TestWhereScope(t *testing.T) {
	db.Users().WhereEq("first_name", "Whatever")

	u := db.Users().Where(map[string]string{"first_name": "Bouke"}).New()
	require.Equal(t, "Bouke", u.FirstName)

	db.Users().Where("first_name IS NOT NULL")
	db.Users().Where("first_name = ?", "Bouke")
	db.Users().Where(rel.In{
		Left: rel.Field{"first_name"},
		Right: []rel.Expr{
			rel.BindParam{"Bouke"},
			rel.BindParam{"COOL"},
		},
	})

	u = db.Users().WhereEq("first_name", "Bouke").WhereEq("last_name", "Tables").New()
	require.Equal(t, "Bouke", u.FirstName)
	require.Equal(t, "Tables", u.LastName)
}

func TestCreatePostUnderUser(t *testing.T) {
	defer clear()
	u := createUser(t)

	err := u.Posts().New().Save(ctx, d)
	require.NoError(t, err)

	u, err = db.Users().Last(ctx, d)
	require.NoError(t, err)
	c, err := u.Posts().Count(ctx, d)
	require.NoError(t, err)
	require.Equal(t, int64(1), c)
}

func TestUpdateUser(t *testing.T) {
	defer clear()
	u := createUser(t)
	u.FirstName = "Bobby"
	u.LastName = "Tables"
	require.NoError(t, u.Save(ctx, d))

	id := u.ID
	u, err := db.Users().Find(ctx, d, id)
	require.NoError(t, err)
	require.Equal(t, u.FirstName, "Bobby")
	require.Equal(t, u.LastName, "Tables")
}

func TestCountUser(t *testing.T) {
	defer clear()

	createUser(t)
	createUser(t)
	createUser(t)

	count, err := db.Users().Count(ctx, d)
	require.NoError(t, err)
	require.EqualValues(t, 3, count)
}

func TestDeleteAll(t *testing.T) {
	defer clear()

	u := createUser(t)
	u.FirstName = "Bouke"
	require.NoError(t, u.Save(ctx, d))
	u = createUser(t)
	u.FirstName = "Bouke"
	require.NoError(t, u.Save(ctx, d))
	u = createUser(t)
	u.FirstName = "Not Bouke"
	require.NoError(t, u.Save(ctx, d))

	count, err := db.Users().Where("first_name = ?", "Bouke").DeleteAll(ctx, d)
	require.NoError(t, err)
	require.EqualValues(t, 2, count)

	count, err = db.Users().Count(ctx, d)
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
}

func TestUpdateAll(t *testing.T) {
	defer clear()

	u := createUser(t)
	u.FirstName = "Bouke"
	require.NoError(t, u.Save(ctx, d))
	u = createUser(t)
	u.FirstName = "Bouke"
	require.NoError(t, u.Save(ctx, d))
	u = createUser(t)
	u.FirstName = "Not Bouke"
	require.NoError(t, u.Save(ctx, d))

	count, err := db.Users().Where("first_name = ?", "Bouke").UpdateAll(ctx, d, "last_name = ?", "Neat")
	require.NoError(t, err)
	require.EqualValues(t, 2, count)

	count, err = db.Users().Where("last_name = ?", "Neat").Count(ctx, d)
	require.NoError(t, err)
	require.EqualValues(t, 2, count)
}

func TestFindBySQL(t *testing.T) {
	defer clear()

	u := createUser(t)
	u.FirstName = "Bouke"
	require.NoError(t, u.Save(ctx, d))
	u = createUser(t)
	u.FirstName = "Bouke"
	require.NoError(t, u.Save(ctx, d))
	u = createUser(t)
	u.FirstName = "Not Bouke"
	require.NoError(t, u.Save(ctx, d))

	users, err := db.Users().FindBySQL(ctx, d, `SELECT users.* FROM users WHERE first_name="Bouke"`)
	require.NoError(t, err)
	require.Len(t, users, 2)
	for _, u := range users {
		require.Equal(t, "Bouke", u.FirstName)
	}
}

func createUser(t *testing.T) *db.User {
	u := db.Users().New()
	err := u.Save(ctx, d)
	require.NoError(t, err)
	require.NotZero(t, u.ID)
	return u
}
