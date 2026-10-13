# Metago

We use [metago](https://github.com/guillemus/metago) to generate some boilerplate.
In [api/model.metago](../model.metago) you will find our custom templates, which are used as follows:

## `model` / `impl`

By annotating a type with `//mgo:gen model` and `//mgo:impl ...`:

```go
//mgo:gen model
//mgo:impl node searchnode activitylogger paginated
type MyType struct {}
```

you can automatically generate methods and types for the supported `impl` keywords,

- `node`: implements the GraphQL `Node` interface
- `searchnode`: marks the model as a search node
- `activitylogger`: marks the model as an activity log producing type
- `paginated`: generates `Connection` and `Edge` types for the model

## `order`

This is used to generate an `Order` type with `Field` and `Direction` for this model.
Used in many list operations to order the response. Must match the respective type in
the GraphQL schema, and you will use the values in your SQL queries to select which
column to order by.

```go
//mgo:gen order FIELD1 FIELD2 FIELD3
type MyType struct {}
```

## `enum`

Used to generate string enums. If marked with `//mgo:enum lowercase`, it will convert the
GraphQL values (generally SCREAMING_UPPER_CASE) to lowercase when marshalling, and similarly
will marshal the value to upper case going the other way.
