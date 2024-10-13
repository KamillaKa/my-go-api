package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Article struct {
	ID      string `json:"_id,omitempty" bson:"_id,omitempty"`
	Title   string `json:"Title" bson:"title"`
	Desc    string `json:"desc" bson:"desc"`
	Content string `json:"content" bson:"content"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

// Filtering, Sorting, and Pagination for retrieving all articles
func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Extracting query parameters for filtering, sorting, and pagination
	queryTitle := r.URL.Query().Get("title") // Filter by title
	queryDesc := r.URL.Query().Get("desc")   // Filter by description
	sortField := r.URL.Query().Get("sort")   // Sort by field (e.g., title, desc)
	sortOrder := r.URL.Query().Get("order")  // Sort order: 1 for ascending, -1 for descending
	pageStr := r.URL.Query().Get("page")     // Page number
	limitStr := r.URL.Query().Get("limit")   // Limit of items per page

	// Set defaults if pagination params are not provided
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10 // Default limit is 10
	}
	skip := (page - 1) * limit // Calculate the offset for pagination

	// Build MongoDB filter
	filter := bson.M{}
	if queryTitle != "" {
		filter["title"] = bson.M{"$regex": queryTitle, "$options": "i"} // Case-insensitive filtering
	}
	if queryDesc != "" {
		filter["desc"] = bson.M{"$regex": queryDesc, "$options": "i"}
	}

	// Sorting logic
	sortOptions := options.Find()
	if sortField != "" {
		order, _ := strconv.Atoi(sortOrder)
		if order != 1 && order != -1 {
			order = 1 // Default to ascending
		}
		sortOptions.SetSort(bson.D{{sortField, order}})
	}

	// Pagination logic
	sortOptions.SetSkip(int64(skip))
	sortOptions.SetLimit(int64(limit))

	// Execute query with filtering, sorting, and pagination
	var articles []Article
	cursor, err := collection.Find(ctx, filter, sortOptions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error fetching articles")
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var article Article
		cursor.Decode(&article)
		articles = append(articles, article)
	}
	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error iterating articles")
		return
	}

	// Send response with articles
	json.NewEncoder(w).Encode(articles)
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	var article Article
	_ = json.NewDecoder(r.Body).Decode(&article)

	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, article)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error creating article")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(article)
}

func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)                              // Get ID from the request URL
	id, err := primitive.ObjectIDFromHex(params["id"]) // Convert to ObjectID
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid ID format")
		return
	}

	var article Article
	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the article by its ObjectID
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&article)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Article not found")
		return
	}

	json.NewEncoder(w).Encode(article)
}

// DELETE an article by ID
func deleteArticle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Convert the ID from the URL string to ObjectID
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Println("Invalid ObjectID:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid ObjectID format"))
		return
	}

	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try deleting the article by its ObjectID
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Println("Error deleting article:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error deleting article"))
		return
	}

	// Check if the article was found and deleted
	if result.DeletedCount == 0 {
		log.Println("Article not found")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Article not found"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Article deleted successfully"))
}

func updateArticle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid ID format")
		return
	}

	var article Article
	_ = json.NewDecoder(r.Body).Decode(&article)

	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": article}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error updating article")
		return
	}

	json.NewEncoder(w).Encode("Article updated successfully")
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/articles", returnAllArticles).Methods("GET")
	router.HandleFunc("/article", createNewArticle).Methods("POST")
	router.HandleFunc("/article/{id}", returnSingleArticle).Methods("GET")
	router.HandleFunc("/article/{id}", updateArticle).Methods("PUT")
	router.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":10000", router))
}

func main() {
	// Load environment variables from the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// MongoDB connection setup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use DB_URL from .env
	dbURI := os.Getenv("DB_URL")
	if dbURI == "" {
		log.Fatal("DB_URL not set in the environment")
	}

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Start the server
	handleRequests()
}
