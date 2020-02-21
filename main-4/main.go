package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/graph-gophers/graphql-go"
)

const schemaString = `
	schema {
		query: Query
	}

	type User {
		userID: ID!
		username: String!
		emoji: String!
		notes: [Note!]!
	}

	type Note {
		noteID: ID!
		data: String!
	}

	type Query {
		users: [User!]!
		user(userID: ID!): User!
		notes(userID: ID!): [Note!]!
		note(noteID: ID!): Note!
	}
`

// User contains infomation about our users.
type User struct {
	UserID   graphql.ID
	Username string
	Emoji    string
	Notes    []Note
}

// Note contains information on a user's notes.
type Note struct {
	NoteID graphql.ID
	Data   string
}

// Define mock data:
var users = []User{
	{
		UserID:   graphql.ID("u-001"),
		Username: "CR7",
		Emoji:    "🇵🇹",
		Notes: []Note{
			{NoteID: "n-001", Data: "Olá Mundo!"},
			{NoteID: "n-002", Data: "Olá novamente, mundo!"},
			{NoteID: "n-003", Data: "Olá, escuridão!"},
		},
	}, {
		UserID:   graphql.ID("u-002"),
		Username: "Sergio_Ramos",
		Emoji:    "🇪🇸",
		Notes: []Note{
			{NoteID: "n-004", Data: "!Hola Mundo!"},
			{NoteID: "n-005", Data: "¡Hola de nuevo mundo!"},
			{NoteID: "n-006", Data: "¡Hola oscuridad!"},
		},
	}, {
		UserID:   graphql.ID("u-003"),
		Username: "michaelmiranda_",
		Emoji:    "🇺🇸",
		Notes: []Note{
			{NoteID: "n-007", Data: "Hello, world!"},
			{NoteID: "n-008", Data: "Hello again, world!"},
			{NoteID: "n-009", Data: "Hello, darkness!"},
		},
	},
}

type rootResolver struct{}

func (r *rootResolver) Users() ([]User, error) {
	return users, nil
}

func (r *rootResolver) User(args struct{ UserID graphql.ID }) (User, error) {
	// Find user:
	for _, user := range users {
		if args.UserID == user.UserID {
			return user, nil
		}
	}

	// User not found:
	return User{}, nil
}

func (r *rootResolver) Notes(args struct{ UserID graphql.ID }) ([]Note, error) {
	// Find user to find notes:
	user, err := r.User(args)
	if reflect.ValueOf(user).IsZero() || err != nil {
		// Didn’t find user:
		return nil, err
	}

	// User found => return notes
	return user.Notes, nil
}

func (r *rootResolver) Note(args struct{ NoteID graphql.ID }) (Note, error) {
	// Find note:
	for _, user := range users {
		for _, note := range user.Notes {
			if args.NoteID == note.NoteID {
				// Found note
				return note, nil
			}
		}
	}

	// No note found
	return Note{}, nil
}

var (
	opts   = []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema = graphql.MustParseSchema(schemaString, &rootResolver{}, opts...)
)

func main() {
	ctx := context.Background()

	type clientQuery struct {
		OpName    string // method name
		Query     string
		Variables map[string]interface{}
	}

	q1 := clientQuery{
		OpName: "Users",
		Query: `query Users {
			users {
				userID
				username
				emoji
			}
		}`,
		Variables: nil,
	}
	resp1 := schema.Exec(ctx, q1.Query, q1.OpName, q1.Variables)
	json1, err := json.MarshalIndent(resp1, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json1))

	q2 := clientQuery{
		OpName: "User",
		Query: `query User($userID: ID!) {
			user(userID: $userID) {
				userID
				username
				emoji
			}
		}`,
		Variables: map[string]interface{}{
			"userID": "u-003",
		},
	}
	resp2 := schema.Exec(ctx, q2.Query, q2.OpName, q2.Variables)
	json2, err := json.MarshalIndent(resp2, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json2))

	q3 := clientQuery{
		OpName: "Notes",
		Query: `query Notes($userID: ID!) {
			notes(userID: $userID) {
				noteID
				data
			}
		}`,
		Variables: map[string]interface{}{
			"userID": "u-001",
		},
	}
	resp3 := schema.Exec(ctx, q3.Query, q3.OpName, q3.Variables)
	json3, err := json.MarshalIndent(resp3, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json3))

	q4 := clientQuery{
		OpName: "Note",
		Query: `query Note($noteID: ID!) {
			note(noteID: $noteID) {
				noteID
				data
			}
		}`,
		Variables: map[string]interface{}{
			"noteID": "n-008",
		},
	}
	resp4 := schema.Exec(ctx, q4.Query, q4.OpName, q4.Variables)
	json4, err := json.MarshalIndent(resp4, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json4))
}
