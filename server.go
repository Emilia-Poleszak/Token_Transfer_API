package main

import (
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Emilia-Poleszak/Token_Transfer_API/graph"
	"github.com/Emilia-Poleszak/Token_Transfer_API/models"
	"github.com/Emilia-Poleszak/Token_Transfer_API/db"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	var DB *gorm.DB = db.ConnectDB()
	if err := DB.AutoMigrate(&models.Wallet{}); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	var count int64
	if err := DB.Model(&models.Wallet{}).Where("address = ?", "0x0000000000000000000000000000000000000000").Count(&count).Error; err != nil {
		log.Fatalf("Count query failed: %v", err)
	}
	
	// Create a default wallet if database is empty
	if count == 0 {
		DB.Create(&models.Wallet{Address: "0x0000000000000000000000000000000000000000", Balance: 1000000})
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
