Go REST API with MongoDB
========================

This is a simple Go REST API that performs basic CRUD operations and includes features like filtering, sorting, and pagination using MongoDB as the database. The project covers the following steps:

*   Step 1: Basic CRUD operations in Go
    
*   Step 2: Integration with MongoDB
    
*   Step 3: Advanced features like filtering, sorting, and pagination
    

Prerequisites
-------------

To run this project, ensure you have the following installed:

*   Go (v1.16 or higher)
    
*   MongoDB (local or remote instance)
    
*   Gorilla Mux (for routing)
    
*   MongoDB Go Driver
    

Step 1: Basic CRUD Operations in Go
-----------------------------------

### Creating a Basic Go REST API

I created a basic Go web application that supports the following CRUD operations:

*   **GET**: Retrieve all or a specific article
    
*   **POST**: Create a new article
    
*   **DELETE**: Delete an article by ID

*   **PUT**: Update an article by ID
    

We use the gorilla/mux package for routing. Here's an overview of the basic API structure:

### Routes

*   **GET /articles**: Returns all articles
    
*   **GET /article/{id}**: Returns a single article by its ID
    
*   **POST /article**: Creates a new article
    
*   **DELETE /article/{id}**: Deletes an article by its ID

*   **PUT /article/{id}**: Update an article by ID
    

Example of the basic CRUD functions:

```go
func createNewArticle(w http.ResponseWriter, r *http.Request) {  
    // Implementation of article creation  
}  

func returnAllArticles(w http.ResponseWriter, r *http.Request) {  
    // Implementation of returning all articles      
}  

func returnSingleArticle(w http.ResponseWriter, r *http.Request) {  
    // Implementation of returning a single article by ID  
}  
``` 

Step 2: MongoDB Integration
---------------------------

In this step, I integrated MongoDB to persist the articles. I replaced the in-memory storage with MongoDB and made use of the MongoDB Go Driver (`go.mongodb.org/mongo-driver/mongo`) to interact with the database. I tried to use the Cloud Database but it didn't allow me to create anything more for free.

### Connecting to MongoDB

I set up a MongoDB client that connects to the database and interacts with a collection named `articles`. Here’s a snippet for establishing a MongoDB connection:

```go
var client *mongo.Client  

func main() {  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)  
    defer cancel()  

    client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))  
    if err != nil {  
        log.Fatal(err)  
    }  
    defer client.Disconnect(ctx)  

    handleRequests()  
}  
``` 

### MongoDB CRUD Operations

I updated the API to interact with MongoDB, using methods like `InsertOne`, `Find`, and `FindOne` for database operations. For example, creating a new article:

```go
func createNewArticle(w http.ResponseWriter, r *http.Request) {  
    var article Article  
    _ = json.NewDecoder(r.Body).Decode(&article)  
    
    collection := client.Database("mydatabase").Collection("articles")  
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
```

Step 3: Advanced Features - Filtering, Sorting, and Pagination
--------------------------------------------------------------

In this step, I implemented more advanced features like filtering, sorting, and pagination to enhance the API.

### Filtering

You can filter articles by title or description using query parameters:

`GET /articles?title=Go&desc=API`

This request will return all articles where the title contains "Go" and the description contains "API".

### Sorting

You can sort articles by a specific field (e.g., `Title`, ``Desc`). The sort order can be ascending (`order=1`) or descending (`order=-1`):

`GET /articles?sort=title&order=1`

### Pagination

To limit the number of results and paginate through large datasets, you can use the `page` and `limit` query parameters:

`GET /articles?page=2&limit=5`

This returns the second page of articles, with 5 articles per page.

### Example of Filtering, Sorting, and Pagination Implementation

```go
func returnAllArticles(w http.ResponseWriter, r *http.Request) {  
    collection := client.Database("mydatabase").Collection("articles")  
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)  
    defer cancel()  
    
    // Extract query parameters
    queryTitle := r.URL.Query().Get("title")   
    queryDesc := r.URL.Query().Get("desc")  
    sortField := r.URL.Query().Get("sort")   
    sortOrder := r.URL.Query().Get("order")   
    pageStr := r.URL.Query().Get("page")  
    limitStr := r.URL.Query().Get("limit")  
    
    // Set defaults for pagination
    page, _ := strconv.Atoi(pageStr)  
    limit, _ := strconv.Atoi(limitStr)  
    if page <= 0 {  
        page = 1  
    }  
    if limit <= 0 {  
        limit = 10  
    }  
    skip := (page - 1) * limit   
        
    // Build filter for MongoDB
    filter := bson.M{}  
    if queryTitle != "" {  
        filter["title"] = bson.M{"$regex": queryTitle, "$options": "i"}  
    }  
    if queryDesc != "" {  
        filter["desc"] = bson.M{"$regex": queryDesc, "$options": "i"}  
    }  
    
    // Sorting logic
    sortOptions := options.Find()  
    if sortField != "" {  
        order, _ := strconv.Atoi(sortOrder)  
        if order != 1 && order != -1 {  
            order = 1  
        }  
        sortOptions.SetSort(bson.D{{sortField, order}})  
    }  
    
    // Pagination logic
    sortOptions.SetSkip(int64(skip))  
    sortOptions.SetLimit(int64(limit))  
    
    // Fetch articles from MongoDB
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
    json.NewEncoder(w).Encode(articles)  
}
```

How to Run the Project
----------------------

1.  Clone the repository: ```git clone https://github.com/KamillaKa/go.git```
    
2.  Install dependencies using Go modules: `go mod tidy`
    
3.  Ensure MongoDB is running locally or update the MongoDB URI in the `main.go` file to point to your MongoDB instance.
    
4.  Run the Go applications: `go run main.go`
    
5.  Access the API at `http://localhost:10000.`
    

Testing Filtering, Sorting, and Pagination:
-------------

*   Filter by title and description: `GET /articles?title=Go&desc=API`

*   Sort by title in ascending order: `GET /articles?sort=title&order=1`

*   Retrieve the second page with 5 articles per page: `GET /articles?page=2&limit=5`
