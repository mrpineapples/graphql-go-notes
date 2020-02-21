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
		greetPerson(person: String!): String!
		greetPersonTimeOfDay(person: String!, timeOfDay: TimeOfDay!): String!
	}

	enum TimeOfDay {
		MORNING
		AFTERNOON
		EVENING
	}
`

var timesOfDay = map[string]string{
	"MORNING":   "Good morning",
	"AFTERNOON": "Good afternoon",
	"EVENING":   "Good evening",
}

type rootResolver struct{}

// Fields must be exported.
type personTimeOfDayArgs struct {
	Person    string
	TimeOfDay string
}

// Always export method for resolvers
func (*rootResolver) Greet() string {
	return "Hello, world!"
}

func (*rootResolver) GreetPerson(args struct{ Person string }) string {
	return fmt.Sprintf("Hello, %s!", args.Person)
}

func (*rootResolver) GreetPersonTimeOfDay(ctx context.Context, args personTimeOfDayArgs) string {
	timeOfDay, ok := timesOfDay[args.TimeOfDay]
	if !ok {
		timeOfDay = "Go to bed"
	}

	return fmt.Sprintf("%s, %s!", timeOfDay, args.Person)
}

var schema = graphql.MustParseSchema(schemaString, &rootResolver{})

func main() {
	ctx := context.Background()

	// Struct to make it easier to test queries
	type clientQuery struct {
		OpName    string // name of method being used
		Query     string
		Variables map[string]interface{}
	}

	q1 := clientQuery{
		OpName: "Greet",
		// Passing in OpName into schema.Exec means that query needs to be in longer form.
		Query: `query Greet { 
			greet 
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
		OpName: "GreetPerson", // method name
		// Queries with variables need to be in longer form.
		Query: `query GreetPerson($person: String!) {
			greetPerson(person: $person)
		}`,
		Variables: map[string]interface{}{
			"person": "Michael",
		},
	}
	resp2 := schema.Exec(ctx, q2.Query, q2.OpName, q2.Variables)
	json2, err := json.MarshalIndent(resp2, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json2))

	q3 := clientQuery{
		OpName: "GreetPersonTimeOfDay",
		Query: `query GreetPersonTimeOfDay($person: String!, $timeOfDay: TimeOfDay!) {
			greetPersonTimeOfDay(person: $person, timeOfDay: $timeOfDay)
		}`,
		Variables: map[string]interface{}{
			"person":    "Michael",
			"timeOfDay": "AFTERNOON",
		},
	}
	resp3 := schema.Exec(ctx, q3.Query, q3.OpName, q3.Variables)
	json3, err := json.MarshalIndent(resp3, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json3))
}
