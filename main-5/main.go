package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	graphql "github.com/graph-gophers/graphql-go"
)

type User struct {
	UserID   graphql.ID
	Username string
	Emoji    string
	Notes    []*Note
}

type Note struct {
	NoteID graphql.ID
	Data   string
}

type NoteInput struct {
	Data string
}

// Define mock data:
var users = []*User{
	{
		UserID:   graphql.ID("u-001"),
		Username: "CR7",
		Emoji:    "🇵🇹",
		Notes: []*Note{
			{NoteID: "n-001", Data: "Olá Mundo!"},
			{NoteID: "n-002", Data: "Olá novamente, mundo!"},
			{NoteID: "n-003", Data: "Olá, escuridão!"},
		},
	}, {
		UserID:   graphql.ID("u-002"),
		Username: "Sergio_Ramos",
		Emoji:    "🇪🇸",
		Notes: []*Note{
			{NoteID: "n-004", Data: "!Hola Mundo!"},
			{NoteID: "n-005", Data: "¡Hola de nuevo mundo!"},
			{NoteID: "n-006", Data: "¡Hola oscuridad!"},
		},
	}, {
		UserID:   graphql.ID("u-003"),
		Username: "michaelmiranda_",
		Emoji:    "🇺🇸",
		Notes: []*Note{
			{NoteID: "n-007", Data: "Hello, world!"},
			{NoteID: "n-008", Data: "Hello again, world!"},
			{NoteID: "n-009", Data: "Hello, darkness!"},
		},
	},
}

type rootResolver struct{}

func (r *rootResolver) Users() ([]*UserResolver, error) {
	var userRxs []*UserResolver
	for _, user := range users {
		userRxs = append(userRxs, &UserResolver{user})
	}
	return userRxs, nil
}

func (r *rootResolver) User(args struct{ UserID graphql.ID }) (*UserResolver, error) {
	// Find user:
	for _, user := range users {
		if args.UserID == user.UserID {
			return &UserResolver{user}, nil
		}
	}

	// No user found
	return nil, nil
}

func (r *rootResolver) Notes(args struct{ UserID graphql.ID }) ([]*NoteResolver, error) {
	// Find user to find notes:
	user, err := r.User(args)
	if user == nil || err != nil {
		return nil, err
	}

	return user.Notes(), nil
}

func (r *rootResolver) Note(args struct{ NoteID graphql.ID }) (*NoteResolver, error) {
	// Find note:
	for _, user := range users {
		for _, note := range user.Notes {
			if args.NoteID == note.NoteID {
				return &NoteResolver{note}, nil
			}
		}
	}

	return nil, nil
}

type CreateNoteArgs struct {
	UserID graphql.ID
	Note   NoteInput
}

func (r *rootResolver) CreateNote(args CreateNoteArgs) (*NoteResolver, error) {
	// Find user:
	var note *Note
	for _, user := range users {
		if args.UserID == user.UserID {
			// Create note
			note = &Note{NoteID: "n-010", Data: args.Note.Data}
			// Add note to user's notes
			user.Notes = append(user.Notes, note)
		}
	}

	return &NoteResolver{note}, nil
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

func (ur *UserResolver) Emoji() string {
	return ur.user.Emoji
}

func (ur *UserResolver) Notes() []*NoteResolver {
	var noteRxs []*NoteResolver
	for _, note := range ur.user.Notes {
		noteRxs = append(noteRxs, &NoteResolver{note})
	}
	return noteRxs
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

func main() {
	ctx := context.Background()

	// Read and parse schema:
	bs, err := ioutil.ReadFile("./schema.graphql")
	if err != nil {
		panic(err)
	}

	schemaString := string(bs)
	schema, err := graphql.ParseSchema(schemaString, &rootResolver{})
	if err != nil {
		panic(err)
	}

	// We can use a type alias for convenience.
	//
	// NOTE: It’s not recommended to use a true type because
	// you’ll need to implement MarshalJSON and UnmarshalJSON.
	type JSON = map[string]interface{}

	type clientQuery struct {
		OpName    string
		Query     string
		Variables JSON
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
		Variables: JSON{
			"userID": "u-001",
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
		Variables: JSON{
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
		Variables: JSON{
			"noteID": "n-001",
		},
	}
	resp4 := schema.Exec(ctx, q4.Query, q4.OpName, q4.Variables)
	json4, err := json.MarshalIndent(resp4, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json4))

	q5 := clientQuery{
		OpName: "CreateNote",
		Query: `mutation CreateNote($userID: ID!, $note: NoteInput!) {
			createNote(userID: $userID, note: $note) {
				noteID
				data
			}
		}`,
		Variables: JSON{
			"userID": "u-003",
			"note": JSON{
				"data": "We created a note!",
			},
		},
	}
	resp5 := schema.Exec(ctx, q5.Query, q5.OpName, q5.Variables)
	json5, err := json.MarshalIndent(resp5, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json5))

	q6 := clientQuery{
		OpName: "Users",
		Query: `query Users {
			users {
				userID
				username
				emoji
				notes {
					noteID
					data
				}
			}
		}`,
		Variables: nil,
	}
	resp6 := schema.Exec(ctx, q6.Query, q6.OpName, q6.Variables)
	json6, err := json.MarshalIndent(resp6, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json6))
}
