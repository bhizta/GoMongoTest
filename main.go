package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func main() {
    // MongoDB connection
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    var err error
    client, err = mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    // Ping the primary
    if err = client.Ping(ctx, nil); err != nil {
        log.Fatal("MongoDB ping error:", err)
    }
    fmt.Println("Successfully connected and pinged MongoDB!")

    // Handle routes
    http.HandleFunc("/idk", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "idk.html")})
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/projects", projectsHandler)
    http.HandleFunc("/add-project", addProjectHandler)

    fmt.Println("Server started at http://localhost:8080")
    http.ListenAndServe(":8080", nil)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "<h1>Welcome to MongoDB test</h1>")
    fmt.Fprintf(w, "<p>Hello, I'm Abhista.</p>")
    fmt.Fprintf(w, "<p>I'm a beginner Go developer.</p>")
    fmt.Fprintf(w, "<p>I'm trying to use mongoDB.</p>")
	fmt.Fprintf(w, `<p>Check out my <a href="/projects">projects</a>!</p>`)
	fmt.Fprintf(w, `<p>Try to add new document <a href="/idk">add</a>!</p>`)
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
    collection := client.Database("portofolio").Collection("projects")
    cursor, err := collection.Find(context.Background(), bson.M{})
    if err != nil {
        fmt.Println("Error fetching cursor:", err)
        http.Error(w, "Error fetching projects", http.StatusInternalServerError)
        return
    }

    var projects []bson.M
    if err = cursor.All(context.Background(), &projects); err != nil {
        fmt.Println("Error parsing projects:", err)
        http.Error(w, "Error parsing projects", http.StatusInternalServerError)
        return
    }

    fmt.Println("Projects fetched and parsed successfully:", projects)
    fmt.Fprintf(w, "<h1>My Projects</h1>")
    for _, project := range projects {
        fmt.Println("Project:", project)  // Debug statement
        fmt.Fprintf(w, "<p>%s: %s</p>", project["name"], project["description"])
    }

    fmt.Fprintf(w, `<p><a href="/">back to home</a>!</p>`)
}

func addProjectHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse form data
    name := r.FormValue("name")
    description := r.FormValue("description")

    // Find the highest current ID
    collection := client.Database("portofolio").Collection("projects")
    var lastProject bson.M
    err := collection.FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})).Decode(&lastProject)
    if err != nil && err != mongo.ErrNoDocuments {
        http.Error(w, "Error finding last project", http.StatusInternalServerError)
        return
    }

     // Determine the new ID
     var newID int32
     if lastProject != nil {
         newID = lastProject["_id"].(int32) + 1
     } else {
         newID = 1 // If there are no documents, start with ID 1
     }

    // Insert new document into MongoDB
    newProject := bson.M{"_id": newID, "name": name, "description": description}
    _, err = collection.InsertOne(context.Background(), newProject)
    if err != nil {
        http.Error(w, "Error adding project", http.StatusInternalServerError)
        return
    }

    // Redirect back to the projects page
    http.Redirect(w, r, "/projects", http.StatusSeeOther)
}
