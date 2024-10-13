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
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title   string             `json:"Title" bson:"title"`
	Desc    string             `json:"desc" bson:"desc"`
	Content string             `json:"content" bson:"content"`
}

// Homepage Handler
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the HomePage!")
}

// Retrieve All Articles with Pagination, Filtering, and Sorting
func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query parameters for filtering, sorting, and pagination
	queryTitle := r.URL.Query().Get("title")
	queryDesc := r.URL.Query().Get("desc")
	sortField := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	skip := (page - 1) * limit

	filter := bson.M{}
	if queryTitle != "" {
		filter["title"] = bson.M{"$regex": queryTitle, "$options": "i"}
	}
	if queryDesc != "" {
		filter["desc"] = bson.M{"$regex": queryDesc, "$options": "i"}
	}

	sortOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit))
	if sortField != "" {
		order, _ := strconv.Atoi(sortOrder)
		if order != 1 && order != -1 {
			order = 1
		}
		sortOptions.SetSort(bson.D{{sortField, order}})
	}

	cursor, err := collection.Find(ctx, filter, sortOptions)
	if err != nil {
		http.Error(w, "Error fetching articles", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var articles []Article
	if err := cursor.All(ctx, &articles); err != nil {
		http.Error(w, "Error iterating articles", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(articles)
}

// Create a New Article
func createNewArticle(w http.ResponseWriter, r *http.Request) {
	var article Article
	if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	article.ID = primitive.NewObjectID() // Generate a new ObjectID
	_, err := collection.InsertOne(ctx, article)
	if err != nil {
		http.Error(w, "Error creating article", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(article)
}

// Retrieve a Single Article by ID
func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var article Article
	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&article)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(article)
}

// Update an Article by ID
func updateArticle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var article Article
	if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": article}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		http.Error(w, "Error updating article", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode("Article updated successfully")
}

// Delete an Article by ID
func deleteArticle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Error deleting article", http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Article deleted successfully")
}

// Handle HTTP Requests
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
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// MongoDB connection setup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbURI := os.Getenv("DB_URL")
	if dbURI == "" {
		log.Fatal("DB_URL not set in the environment")
	}

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	handleRequests()
}
