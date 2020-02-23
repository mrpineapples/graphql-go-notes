package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"

	graphql "github.com/graph-gophers/graphql-go"
	_ "github.com/lib/pq"
)

var (
	DB     *sql.DB
	Schema *graphql.Schema
)

type User struct {
	UserID   graphql.ID
	Username string
	Notes    []*Note
}

type Note struct {
	NoteID graphql.ID
	Data   string
}

type NoteInput struct {
	Data string
}

type RootResolver struct{}

func (rr *RootResolver) Users() ([]*UserResolver, error) {
	var userRxs []*UserResolver
	rows, err := DB.Query(`
		SELECT
			user_id,
			username
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.UserID, &user.Username)
		if err != nil {
			return nil, err
		}
		userRxs = append(userRxs, &UserResolver{user})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return userRxs, nil
}

func (rr *RootResolver) User(args struct{ UserID graphql.ID }) (*UserResolver, error) {
	user := &User{}
	err := DB.QueryRow(`
		SELECT
			user_id,
			username
		FROM users
		WHERE user_id = $1
	`, args.UserID).Scan(&user.UserID, &user.Username)
	if err != nil {
		return nil, err
	}

	return &UserResolver{user}, nil
}

func (rr *RootResolver) Notes(args struct{ UserID graphql.ID }) ([]*NoteResolver, error) {
	var noteRxs []*NoteResolver
	rows, err := DB.Query(`
		SELECT
			note_id,
			data
		FROM notes
		WHERE user_id = $1
	`, args.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		note := &Note{}
		err := rows.Scan(&note.NoteID, &note.Data)
		if err != nil {
			return nil, err
		}
		noteRxs = append(noteRxs, &NoteResolver{note})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return noteRxs, nil
}

func (rr *RootResolver) Note(args struct{ NoteID graphql.ID }) (*NoteResolver, error) {
	note := &Note{}
	err := DB.QueryRow(`
		SELECT
			note_id,
			data
		FROM notes
		WHERE note_id = $1
	`, args.NoteID).Scan(&note.NoteID, &note.Data)
	if err != nil {
		return nil, err
	}

	return &NoteResolver{note}, nil
}

type CreateNoteArgs struct {
	UserID graphql.ID
	Note   NoteInput
}

func (rr *RootResolver) CreateNote(args CreateNoteArgs) (*NoteResolver, error) {
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var noteID string
	err = tx.QueryRow(`
		INSERT INTO notes (
			user_id,
			data
		)
		VALUES ($1, $2)
		RETURNING note_id
	`, args.UserID, args.Note.Data).Scan(&noteID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// Returns note with the note id casted as a graphql.ID
	return rr.Note(struct{ NoteID graphql.ID }{graphql.ID(noteID)})
}

type UserResolver struct {
	user *User
}

func (ur *UserResolver) UserID() graphql.ID {
	return ur.user.UserID
}

func (ur *UserResolver) Username() string {
	return ur.user.Username
}

func (ur *UserResolver) Notes() ([]*NoteResolver, error) {
	rootRx := &RootResolver{}
	return rootRx.Notes(struct{ UserID graphql.ID }{ur.user.UserID})
}

type NoteResolver struct {
	note *Note
}

func (nr *NoteResolver) NoteID() graphql.ID {
	return nr.note.NoteID
}

func (nr *NoteResolver) Data() string {
	return nr.note.Data
}

func check(err error, desc string) {
	if err == nil {
		return
	}
	errStr := fmt.Sprintf("%s: %s", desc, err)
	panic(errStr)
}

func main() {
	// Connect to DB
	var err error
	DB, err = sql.Open("postgres", "host=localhost port=5432 user=postgres dbname=postgres password=example sslmode=disable")
	check(err, "sql.Open")
	err = DB.Ping()
	check(err, "sql.Ping")
	defer DB.Close()

	// Parse schema
	bs, err := ioutil.ReadFile("./schema.graphql")
	check(err, "ioutil.ReadFile")
	schemaString := string(bs)
	Schema, err = graphql.ParseSchema(schemaString, &RootResolver{})
	check(err, "graphql.ParseSchema")

	ctx := context.Background()

	type JSON = map[string]interface{}

	type ClientQuery struct {
		OpName    string
		Query     string
		Variables JSON
	}

	q1 := ClientQuery{
		OpName: "Users",
		Query: `query Users {
			users {
				userID
				username
			}
		}`,
		Variables: nil,
	}
	resp1 := Schema.Exec(ctx, q1.Query, q1.OpName, q1.Variables)
	json1, err := json.MarshalIndent(resp1, "", "\t")
	check(err, "json.MarshalIndent")
	fmt.Println(string(json1))

	q2 := ClientQuery{
		OpName: "User",
		Query: `query User($userID: ID!) {
			user(userID: $userID) {
				userID
				username
			}
		}`,
		Variables: JSON{
			"userID": "u-668431",
		},
	}
	resp2 := Schema.Exec(ctx, q2.Query, q2.OpName, q2.Variables)
	json2, err := json.MarshalIndent(resp2, "", "\t")
	check(err, "json.MarshalIndent")
	fmt.Println(string(json2))

	q3 := ClientQuery{
		OpName: "Notes",
		Query: `query Notes($userID: ID!) {
			notes(userID: $userID) {
				noteID
				data
			}
		}`,
		Variables: JSON{
			"userID": "u-668431",
		},
	}
	resp3 := Schema.Exec(ctx, q3.Query, q3.OpName, q3.Variables)
	json3, err := json.MarshalIndent(resp3, "", "\t")
	check(err, "json.MarshalIndent")
	fmt.Println(string(json3))

	q4 := ClientQuery{
		OpName: "Note",
		Query: `query Note($noteID: ID!) {
			note(noteID: $noteID) {
				noteID
				data
			}
		}`,
		Variables: JSON{
			"noteID": "n-c4127e",
		},
	}
	resp4 := Schema.Exec(ctx, q4.Query, q4.OpName, q4.Variables)
	json4, err := json.MarshalIndent(resp4, "", "\t")
	check(err, "json.MarshalIndent")
	fmt.Println(string(json4))

	q5 := ClientQuery{
		OpName: "CreateNote",
		Query: `mutation CreateNote($userID: ID!, $note: NoteInput!) {
			createNote(userID: $userID, note: $note) {
				noteID
				data
			}
		}`,
		Variables: JSON{
			"userID": "u-dfe108",
			"note": JSON{
				"data": "We created a note!",
			},
		},
	}
	resp5 := Schema.Exec(ctx, q5.Query, q5.OpName, q5.Variables)
	json5, err := json.MarshalIndent(resp5, "", "\t")
	check(err, "json.MarshalIndent")
	fmt.Println(string(json5))

	q6 := ClientQuery{
		OpName: "Users",
		Query: `query Users {
			users {
				userID
				username
				notes {
					noteID
					data
				}
			}
		}`,
		Variables: nil,
	}
	resp6 := Schema.Exec(ctx, q6.Query, q6.OpName, q6.Variables)
	json6, err := json.MarshalIndent(resp6, "", "\t")
	check(err, "json.MarshalIndent")
	fmt.Println(string(json6))
}
