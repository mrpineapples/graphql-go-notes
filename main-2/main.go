package main

import (
	"context"
	"encoding/json"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
)

const schemaString = `
	schema {
		query: Query
	}

	type Query {
		greet: String!
	}
`

type rootResolver struct{}

// Greet needs to be exported. Unexported resolver methods will error.
func (*rootResolver) Greet() string {
	return "Hello, world!"
}

var schema = graphql.MustParseSchema(schemaString, &rootResolver{})

func main() {
	query := ` {
		greet
	}
	`

	ctx := context.Background()
	resp := schema.Exec(ctx, query, "", nil)
	json, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(json))
}
