package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Property struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Title       string             `bson:"Title" json:"Title"`
	Developer   string             `bson:"Developer" json:"Developer"`
	Description string             `bson:"Description" json:"Description"`
	Coordinates [2]float64         `bson:"Coordinates" json:"Coordinates"`
	PriceRange  string             `bson:"Price-range" json:"Price-range"`
	Facilities  []string           `bson:"Facilities" json:"Facilities"`
	Built       int                `bson:"Built" json:"Built"`
	CreatedAt   time.Time          `bson:"Created_at" json:"Created_at"`
}

// Initialize MongoDB client
func connectMongoDB() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
}

func getProperties(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("MVDB").Collection("properties")
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve Properties from MongoDB", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var properties []Property
	for cur.Next(ctx) {
		var property Property
		if err := cur.Decode(&property); err != nil {
			http.Error(w, "Failed to decode retrieved Properties", http.StatusInternalServerError)
			return
		}
		properties = append(properties, property)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error iterating through cursor", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(properties)
}

func main() {
	connectMongoDB()
	r := mux.NewRouter()

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Allow requests from all origins
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-API-Key"}),
	)

	// Create a new handler with CORS middleware
	handler := cors(r)

	// Routes
	r.HandleFunc("/properties", getProperties).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default port to 8000 if PORT environment variable is not set
	}
	fmt.Println("Server is running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
