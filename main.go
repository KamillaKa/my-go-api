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

// Function to initialize mock data
func initializeMockData() {
	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Mock articles
	mockArticles := []Article{
		{Title: "The Rise of AI", Desc: "Exploring the advancements in Artificial Intelligence.", Content: "Artificial Intelligence (AI) has made significant strides in recent years, impacting industries and everyday life."},
		{Title: "Understanding Blockchain Technology", Desc: "A deep dive into how blockchain works.", Content: "Blockchain technology underpins cryptocurrencies like Bitcoin and has potential applications across various sectors."},
		{Title: "Climate Change: The Facts", Desc: "Discussing the realities of climate change and its impacts.", Content: "Climate change poses a significant threat to our planet, with rising temperatures and extreme weather events becoming more common."},
		{Title: "Health Benefits of Meditation", Desc: "The positive effects of meditation on mental health.", Content: "Meditation has been shown to reduce stress, enhance concentration, and promote emotional health."},
		{Title: "Top 10 Programming Languages in 2024", Desc: "A look at the most popular programming languages this year.", Content: "In 2024, languages like Python, JavaScript, and Go continue to dominate the programming landscape."},
		{Title: "Traveling the World on a Budget", Desc: "Tips and tricks for affordable travel.", Content: "Traveling doesn't have to break the bank. Here are some tips to make the most of your adventures without overspending."},
		{Title: "The Future of Electric Vehicles", Desc: "How electric vehicles are shaping the future of transportation.", Content: "Electric vehicles (EVs) are becoming more popular due to advancements in technology and growing environmental concerns."},
		{Title: "Healthy Eating: Myths vs. Facts", Desc: "Debunking common misconceptions about healthy eating.", Content: "Many myths surround nutrition. It's essential to distinguish between fact and fiction to maintain a balanced diet."},
		{Title: "The Impact of Social Media on Society", Desc: "Analyzing the effects of social media on communication and relationships.", Content: "Social media has transformed how we connect, but it also brings challenges related to mental health and privacy."},
		{Title: "Learning a New Language: Tips for Success", Desc: "Effective strategies for mastering a new language.", Content: "Learning a new language can be daunting, but with the right techniques, anyone can achieve fluency."},
	}

	// Check if the collection is empty before inserting mock data
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		// Insert mock articles into the collection
		_, err = collection.InsertMany(ctx, mockArticles)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Mock data inserted successfully!")
	} else {
		fmt.Println("Mock data already exists, skipping insertion.")
	}
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
	vars := mux.Vars(r)
	id := vars["id"]

	collection := client.Database("articles").Collection("go")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var article Article
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&article)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Article not found")
		return
	}
	json.NewEncoder(w).Encode(article)
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/articles", returnAllArticles).Methods("GET")
	router.HandleFunc("/article", createNewArticle).Methods("POST")
	router.HandleFunc("/article/{id}", returnSingleArticle).Methods("GET")
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

	// Initialize mock data
	initializeMockData()

	// Start the server
	handleRequests()
}
