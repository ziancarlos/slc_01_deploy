package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/julienschmidt/httprouter"
)

type Product struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}

type WebResponse struct {
	Data any `json:"data"`
}

func main() {
	db, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close(context.Background())

	router := httprouter.New()

	router.GET("/products", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		rows, err := db.Query(context.Background(), "SELECT id, name, description, price FROM products")

		if err != nil {
			panic(err)
		}
		defer rows.Close()

		products := []Product{}
		for rows.Next() {
			product := Product{}
			rows.Scan(&product.Id, &product.Name, &product.Description, &product.Price)
			products = append(products, product)
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		encoder := json.NewEncoder(writer)
		encoder.Encode(WebResponse{
			Data: products,
		})
	})

	router.GET("/products/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		rows, err := db.Query(context.Background(), "SELECT id, name, description, price FROM products WHERE id = $1", id)

		if err != nil {
			panic(err)
		}
		defer rows.Close()

		product := Product{}
		if rows.Next() {
			rows.Scan(&product.Id, &product.Name, &product.Description, &product.Price)
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		encoder := json.NewEncoder(writer)
		encoder.Encode(WebResponse{
			Data: product,
		})
	})

	router.POST("/products", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		product := Product{}
		decoder := json.NewDecoder(request.Body)
		err := decoder.Decode(&product)
		fmt.Println(err)
		if err != nil {
			panic(err)
		}

		_, err = db.Exec(context.Background(), "INSERT INTO  products (name, description, price) VALUES($1, $2, $3)", product.Name, product.Description, product.Price)
		if err != nil {
			panic(err)
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		encoder := json.NewEncoder(writer)
		encoder.Encode(WebResponse{
			Data: "Successfuly created new products",
		})
	})

	router.PUT("/products/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		product := Product{}
		decoder := json.NewDecoder(request.Body)
		decoder.Decode(&product)

		_, err := db.Exec(context.Background(), "UPDATE products SET name=$1, price=$2, description=$3 WHERE id=$4", product.Name, product.Price, product.Description, id)
		if err != nil {
			panic(err)
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		encoder := json.NewEncoder(writer)
		encoder.Encode(WebResponse{
			Data: "Successfuly updated products",
		})
	})

	router.DELETE("/products/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")

		_, err := db.Exec(context.Background(), "DELETE FROM products WHERE id=$1", id)
		if err != nil {
			panic(err)
		}

		writer.Header().Add("Content-Type", "application/json")
		writer.WriteHeader(200)
		encoder := json.NewEncoder(writer)
		encoder.Encode(WebResponse{
			Data: "Successfuly deleted products",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	addr := "localhost:" + port
	server := http.Server{
		Addr:    addr,
		Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
