package v2

import (
	"mime"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/n0m-d/DVAPI/graph"
	"github.com/vektah/gqlparser/v2/ast"

	ghandler "github.com/99designs/gqlgen/graphql/handler"
)

func registerGraphQL(r *gin.RouterGroup, notes graph.NoteService, userService graph.UserLister) {
	srv := ghandler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{NoteService: notes, UserService: userService},
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{}) // VULN(graphql-introspection): introspection enabled in production.
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	r.GET("/playground", gin.WrapH(playground.Handler("GraphQL Playground", "/api/v2/query")))
	r.Any("/query", graphqlQueryHandler(srv))
}

func graphqlQueryHandler(srv *ghandler.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		// gqlgen POST transport only accepts application/json.
		// Many clients (curl, some REST UIs) omit Content-Type.
		if c.Request.Method == http.MethodPost {
			mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
			if err != nil || mediaType == "" || mediaType == "text/plain" {
				c.Request.Header.Set("Content-Type", "application/json")
			}
		}
		srv.ServeHTTP(c.Writer, c.Request)
	}
}
