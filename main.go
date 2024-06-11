package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Inquiry struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"inquiry_id,omitempty"`
	User_id     string             `bson:"user_id" json:"user_id"`
	Property_id string             `bson:"property_id" json:"property_id"`
	Message     string             `bson:"message" json:"message"`
	CreatedAt   time.Time          `bson:"Created_at" json:"Created_at"`
}

type Appointment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"appointment_id,omitempty"`
	UserID          string             `bson:"User_id,omitempty" json:"User_id,omitempty"`
	PropertyID      string             `bson:"Property_id,omitempty" json:"Property_id,omitempty"`
	ListingID       string             `bson:"Listing_id,omitempty" json:"Listing_id,omitempty"`
	AppointmentDate time.Time          `bson:"Appointment_date,omitempty" json:"Appointment_date,omitempty"`
	Status          string             `bson:"Status,omitempty" json:"Status,omitempty"` // scheduled, completed, cancelled
	CreatedAt       time.Time          `bson:"Created_at,omitempty" json:"Created_at,omitempty"`
}

// User represents the structure of a user document
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	Name      string             `bson:"name,omitempty" json:"name,omitempty"`
	Email     string             `bson:"email,omitempty" json:"email,omitempty"`
	Password  string             `bson:"password,omitempty" json:"password,omitempty"`
	Phone     string             `bson:"phone,omitempty" json:"phone,omitempty"`
	Role      string             `bson:"role,omitempty" json:"role,omitempty"` // client or admin
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
}

type Property struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"property_id,omitempty"`
	Title       string             `bson:"Title" json:"Title"`
	Developer   string             `bson:"Developer" json:"Developer"`
	Description string             `bson:"Description" json:"Description"`
	Coordinates [2]float64         `bson:"Coordinates" json:"Coordinates"`
	MinPrice    int                `bson:"MinPrice" json:"MinPrice"`
	MaxPrice    int                `bson:"MaxPrice" json:"MaxPrice"`
	Facilities  []string           `bson:"Facilities" json:"Facilities"`
	Images      []string           `bson:"Images" json:"Images"`
	Built       int                `bson:"Built" json:"Built"`
	CreatedAt   time.Time          `bson:"Created_at" json:"Created_at"`
}

// Initialize MongoDB client
func connectMongoDB() {
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

// Handler to get all properties
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

func getInquires(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("MVDB").Collection("inquiries")
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve Inquiries from MongoDB", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var inquiries []Inquiry
	for cur.Next(ctx) {
		var inquiry Inquiry
		if err := cur.Decode(&inquiry); err != nil {
			http.Error(w, "Failed to decode retrieved Inquiries", http.StatusInternalServerError)
			return
		}
		inquiries = append(inquiries, inquiry)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error iterating through cursor", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(inquiries)
}

// Handler to upload an image
func uploadImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// Parse the form data
	err := r.ParseMultipartForm(10 << 20) // Max file size: 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	// Get the file from form data
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Unable to get the file from form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		http.Error(w, "Failed to initialize Cloudinary", http.StatusInternalServerError)
		return
	}

	// Upload the file to Cloudinary
	uploadResult, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{})
	if err != nil {
		http.Error(w, "Failed to upload image to Cloudinary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the upload result for debugging
	fmt.Printf("Upload Result: %+v\n", uploadResult)

	// Check if the SecureURL is empty
	if uploadResult.SecureURL == "" {
		http.Error(w, "Empty SecureURL returned from Cloudinary", http.StatusInternalServerError)
		return
	}

	// Update the property with the image URL
	collection := client.Database("MVDB").Collection("properties")
	update := bson.M{
		"$push": bson.M{
			"Images": uploadResult.SecureURL,
		},
	}
	_, err = collection.UpdateByID(context.Background(), id, update)
	if err != nil {
		http.Error(w, "Failed to update property with image URL", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bson.M{"message": "Image uploaded successfully", "url": uploadResult.SecureURL})
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
	r.HandleFunc("/inquiries", getInquires).Methods("GET")
	r.HandleFunc("/properties/{id}/images", uploadImage).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default port to 8000 if PORT environment variable is not set
	}
	fmt.Println("Server is running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
