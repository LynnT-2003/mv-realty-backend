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

type Inquiry struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"inquiry_id,omitempty"`
	User_id     string             `bson:"user_id" json:"user_id"`
	Property_id string             `bson:"property_id" json:"property_id"`
	Message     string             `bson:"message" json:"message"`
	CreatedAt   time.Time          `bson:"Created_at" json:"Created_at"`
}

type Appointment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"appointment_id,omitempty"`
	UserID          string             `bson:"User_id" json:"User_id"`
	PropertyID      string             `bson:"Property_id" json:"Property_id"`
	ListingID       string             `bson:"Listing_id" json:"Listing_id"`
	AppointmentDate time.Time          `bson:"Appointment_date" json:"Appointment_date"`
	Status          string             `bson:"Status" json:"Status"` // scheduled, completed, cancelled
	CreatedAt       time.Time          `bson:"Created_at" json:"Created_at"`
}

// User represents the structure of a user document
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"user_id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Phone     string             `bson:"phone" json:"phone"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
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

type Listing struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"listing_id,omitempty"`
	PropertyID      string             `bson:"property_id" json:"property_id"`
	Description     string             `bson:"description" json:"description"`
	Price           float64            `bson:"price" json:"price"`
	MinimumContract string             `bson:"minimum_contract" json:"minimum_contract"`
	Floor           int                `bson:"floor" json:"floor"`
	Size            float64            `bson:"size" json:"size"` // size in square meters
	Bedroom         int                `bson:"bedroom" json:"bedroom"`
	Bathroom        int                `bson:"bathroom" json:"bathroom"`
	Furniture       string             `bson:"furniture" json:"furniture"`               // fully-fitted or fully furnished
	Status          string             `bson:"status" json:"status"`                     // ready to move in or finishing in 2026
	ListingType     string             `bson:"listing_type" json:"listing_type"`         // sale or rent
	FacingDirection string             `bson:"facing_direction" json:"facing_direction"` // N, S, E, W, NE, NW, SE, SW
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	Photos          []string           `bson:"photos" json:"photos"`                 // URLs of photos
	ListingStatus   string             `bson:"listing_status" json:"listing_status"` // active or inactive
}

var client *mongo.Client

func connectMongoDB() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set")
	}

	// Create a new context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize the MongoDB client
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	// Ping the MongoDB server to ensure connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB:", err)
	}

	// Print success messages
	fmt.Println("Connected to MongoDB!")
	log.Println("MongoDB Client Initialized:", client)
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

func getAppointments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("MVDB").Collection("appointments")
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve Appointments from MongoDB", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var appointments []Appointment
	for cur.Next(ctx) {
		var appointment Appointment
		if err := cur.Decode(&appointment); err != nil {
			http.Error(w, "Failed to decode retrieved Appointments", http.StatusInternalServerError)
			return
		}
		appointments = append(appointments, appointment)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error iterating through cursor", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(appointments)
}
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("getUsers called") // Log the start of the function
	log.Println("MongoDB client status in getUsers:", client)

	if client == nil {
		log.Println("MongoDB client is not initialized")
		http.Error(w, "MongoDB client is not initialized", http.StatusInternalServerError)
		return
	}

	collection := client.Database("MVDB").Collection("users")
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Failed to retrieve Users from MongoDB:", err)
		http.Error(w, "Failed to retrieve Users from MongoDB", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var users []User
	for cur.Next(ctx) {
		var user User
		if err := cur.Decode(&user); err != nil {
			log.Println("Failed to decode retrieved Users:", err)
			http.Error(w, "Failed to decode retrieved Users", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	if err := cur.Err(); err != nil {
		log.Println("Error iterating through cursor:", err)
		http.Error(w, "Error iterating through cursor", http.StatusInternalServerError)
		return
	}
	log.Println("Successfully retrieved users")
	json.NewEncoder(w).Encode(users)
}


// func getUsers(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	log.Println("getUsers called") // Log the start of the function

// 	if client == nil {
// 		log.Println("MongoDB client is not initialized")
// 		http.Error(w, "MongoDB client is not initialized", http.StatusInternalServerError)
// 		return
// 	}

// 	collection := client.Database("MVDB").Collection("users")
// 	cur, err := collection.Find(ctx, bson.M{})
// 	if err != nil {
// 		log.Println("Failed to retrieve Users from MongoDB:", err)
// 		http.Error(w, "Failed to retrieve Users from MongoDB", http.StatusInternalServerError)
// 		return
// 	}
// 	defer cur.Close(ctx)

// 	var users []User
// 	for cur.Next(ctx) {
// 		var user User
// 		if err := cur.Decode(&user); err != nil {
// 			log.Println("Failed to decode retrieved Users:", err)
// 			http.Error(w, "Failed to decode retrieved Users", http.StatusInternalServerError)
// 			return
// 		}
// 		users = append(users, user)
// 	}
// 	if err := cur.Err(); err != nil {
// 		log.Println("Error iterating through cursor:", err)
// 		http.Error(w, "Error iterating through cursor", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Println("Successfully retrieved users")
// 	json.NewEncoder(w).Encode(users)
// }


func getListings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("MVDB").Collection("listings")
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve Listings from MongoDB", http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var listings []Listing
	for cur.Next(ctx) {
		var listing Listing
		if err := cur.Decode(&listing); err != nil {
			http.Error(w, "Failed to decode retrieved Listings", http.StatusInternalServerError)
			return
		}
		listings = append(listings, listing)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error iterating through cursor", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(listings)
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

func createProperty(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body for POST
	var property Property
	err := json.NewDecoder(r.Body).Decode(&property)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusInternalServerError)
		return
	}

	// Set CreatedAt timestamp
	property.CreatedAt = time.Now()
	property.Images = []string{}

	// Insert property into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("MVDB").Collection("properties")
	result, err := collection.InsertOne(ctx, property)
	if err != nil {
		http.Error(w, "Failed to create Property", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bson.M{"property_id": result.InsertedID})
}

func createListing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body for POST
	var listing Listing
	err := json.NewDecoder(r.Body).Decode(&listing)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusInternalServerError)
		return
	}

	// Ctx, cancel
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Validation (check if propertyId exists in Properties Collection)
	propertiesCollection := client.Database("MVDB").Collection("properties")
	var property Property
	propertyID, err := primitive.ObjectIDFromHex(listing.PropertyID)
	if err != nil {
		http.Error(w, "Invalid PropertyID format", http.StatusBadRequest)
		return
	}
	err = propertiesCollection.FindOne(ctx, bson.M{"_id": propertyID}).Decode(&property)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "PropertyID does not exist", http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to check PropertyID", http.StatusInternalServerError)
		}
		return
	}

	// Set CreatedAt timestamp
	listing.CreatedAt = time.Now()
	listing.Photos = []string{}

	// Insert listing into MongoDB
	listingsCollection := client.Database("MVDB").Collection("listings")
	result, err := listingsCollection.InsertOne(ctx, listing)
	if err != nil {
		http.Error(w, "Failed to create Listing", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bson.M{"listing_id": result.InsertedID})
}

func createInquiry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body for POST
	var inquiry Inquiry
	err := json.NewDecoder(r.Body).Decode(&inquiry)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusInternalServerError)
		return
	}

	// Ctx, cancel
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Validation Check
	// TODO: Complete any validation / verification

	// Set CreatedAt timestamp
	inquiry.CreatedAt = time.Now()

	// Insert inquiry into MongoDB
	inquiriesCollection := client.Database("MVDB").Collection("inquiries")
	result, err := inquiriesCollection.InsertOne(ctx, inquiry)
	if err != nil {
		http.Error(w, "Failed to create Inquiry", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bson.M{"inquiry_id": result.InsertedID})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body for POST
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusInternalServerError)
		return
	}

	// Set CreatedAt timestamp
	user.CreatedAt = time.Now()

	// Insert User into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("MVDB").Collection("users")
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Failed to create User", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bson.M{"user_id": result.InsertedID})
}

func checkUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email query parameter is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("MVDB").Collection("users")
	var user User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			json.NewEncoder(w).Encode(bson.M{"exists": false})
			return
		}
		http.Error(w, "Error checking user existence", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bson.M{"exists": true})
}

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file:", err)
    }

    // apiKey = os.Getenv("API_KEY")
    // if apiKey == "" {
    //     log.Fatal("API_KEY environment variable not set")
    // }

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
	r.HandleFunc("/appointments", getAppointments).Methods("GET")
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/check/user", checkUser).Methods("GET")
	r.HandleFunc("/listings", getListings).Methods("GET")

	r.HandleFunc("/add/property", createProperty).Methods("POST")
	r.HandleFunc("/add/listing", createListing).Methods("POST")
	r.HandleFunc("/add/inquiry", createInquiry).Methods("POST")
	r.HandleFunc("/add/user", createUser).Methods("POST")

	r.HandleFunc("/properties/{id}/images", uploadImage).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default port to 8000 if PORT environment variable is not set
	}
	fmt.Println("Server is running on port:", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
