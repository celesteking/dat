package runner

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"

	"github.com/mgutz/dat/v1"
	"github.com/mgutz/dat/v1/postgres"
)

var conn *Connection
var db *sql.DB

func init() {
	dat.Dialect = postgres.New()
	db = realDb()
	conn = NewConnection(db, "postgres")
	dat.Strict = false
}

func createRealSession() *Session {
	sess, err := conn.NewSession()
	if err != nil {
		panic(err)
	}
	return sess
}

func createRealSessionWithFixtures() *Session {
	installFixtures()
	sess := createRealSession()
	return sess
}

func quoteColumn(column string) string {
	var buffer bytes.Buffer
	dat.Dialect.WriteIdentifier(&buffer, column)
	return buffer.String()
}

func quoteSQL(sqlFmt string, cols ...string) string {
	args := make([]interface{}, len(cols))

	for i := range cols {
		args[i] = quoteColumn(cols[i])
	}

	return fmt.Sprintf(sqlFmt, args...)
}

func realDb() *sql.DB {
	driver := os.Getenv("DAT_DRIVER")
	if driver == "" {
		logger.Fatal("env DAT_DRIVER is not set")
	}

	dsn := os.Getenv("DAT_DSN")
	if dsn == "" {
		logger.Fatal("env DAT_DSN is not set")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		logger.Fatal("Database error ", "err", err)
	}

	return db
}

type Person struct {
	ID        int64           `db:"id"`
	Amount    dat.NullFloat64 `db:"amount"`
	Doc       dat.NullString  `db:"doc"`
	Email     dat.NullString  `db:"email"`
	Foo       string          `db:"foo"`
	Image     []byte          `db:"image"`
	Key       dat.NullString  `db:"key"`
	Name      string          `db:"name"`
	CreatedAt dat.NullTime    `db:"created_at"`
}
type Post struct {
	ID        int          `db:"id"`
	UserID    int          `db:"user_id"`
	State     string       `db:"state"`
	Title     string       `db:"title"`
	DeletedAt dat.NullTime `db:"deleted_at"`
	CreatedAt dat.NullTime `db:"created_at"`
}

func installFixtures() {
	db := conn.DB
	createTablePeople := `
		CREATE TABLE people (
			id SERIAL PRIMARY KEY,
			amount decimal,
			doc hstore,
			email text,
			foo text default 'bar',
			image bytea,
			key text,
			name text NOT NULL,
			created_at timestamptz default now()
		)
	`
	createTablePosts := `
		CREATE TABLE posts (
			id SERIAL PRIMARY KEY,
			user_id int references people(id),
			state text,
			title text,
			deleted_at timestamptz,
			created_at timestamptz default now()
		)
	`
	createTableComments := `
		CREATE TABLE comments (
			id SERIAL PRIMARY KEY,
			user_id int references people(id),
			post_id int references posts(id),
			comment text not null,
			created_at timestamptz default now()
		)
	`

	sqlToRun := []string{
		"DROP TABLE IF EXISTS comments",
		"DROP TABLE IF EXISTS posts",
		"DROP TABLE IF EXISTS people",
		createTablePeople,
		createTablePosts,
		createTableComments,
		`
DO $$
BEGIN

	INSERT INTO people (id, name, email) VALUES
		(1, 'Mario', 'mario@acme.com'),
		(2, 'John', 'john@acme.com');

	INSERT INTO posts (id, user_id, title, state) VALUES
		(1, 1, 'Day 1', 'published'),
		(2, 1, 'Day 2', 'draft'),
		(3, 2, 'Apple', 'published'),
		(4, 2, 'Orange', 'draft');

	INSERT INTO comments (id, user_id, post_id, comment) VALUES
		(1, 1, 1, 'A very good day'),
		(2, 2, 3, 'Yum. Apple pie.');

	alter sequence people_id_seq RESTART with 100;
	alter sequence posts_id_seq RESTART with 100;
	alter sequence comments_id_seq RESTART with 100;
END $$
`,
	}

	for _, v := range sqlToRun {
		_, err := db.Exec(v)
		if err != nil {
			logger.Fatal("Failed to execute statement", "sql", v, "err", err)
		}
	}
}