package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"  // Import strconv for string conversion
    "time"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

var db *gorm.DB

// Event model
// Event model
type Event struct {
    ID          uint       `gorm:"primaryKey"`
    Name        string     `json:"name"`
    Date        time.Time  `json:"date"`
    Time        string     `json:"time"`
    Location    string     `json:"location"`
    Description string     `json:"description"`
    RSVPs       []RSVP     `json:"rsvps"`
}

// RSVP model
type RSVP struct {
    ID        uint   `gorm:"primaryKey"`
    EventID   uint   `json:"event_id"`
    UserID    uint   `json:"user_id"`
    Response  string `json:"response"`
}

// Initialize the database
func InitDB() {
    var err error
    db, err = gorm.Open(sqlite.Open("events.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database", err)
    }
    db.AutoMigrate(&Event{}, &RSVP{})
}

// API Handlers

// CreateEvent handles POST requests to create a new event
func CreateEvent(w http.ResponseWriter, r *http.Request) {
    var event Event

    // Decode the JSON request body
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Validate the parsed date
    if event.Date.IsZero() {
        http.Error(w, "Invalid date", http.StatusBadRequest)
        return
    }

    // Save the event to the database
    if err := db.Create(&event).Error; err != nil {
        http.Error(w, "Failed to create event", http.StatusInternalServerError)
        return
    }

    // Respond with the created event
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(event)
}

func GetEvents(w http.ResponseWriter, r *http.Request) {
    filter := r.URL.Query().Get("filter") // Filter query parameter
    var events []Event

    switch filter {
    case "upcoming":
        db.Preload("RSVPs").Where("date >= ?", time.Now()).Find(&events)
    case "past":
        db.Preload("RSVPs").Where("date < ?", time.Now()).Find(&events)
    default:
        db.Preload("RSVPs").Find(&events)
    }

    json.NewEncoder(w).Encode(events)
}

// GetEvent retrieves an event by its ID
func GetEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    eventID := params["id"]

    log.Printf("Extracted event ID: %s", eventID) // Debugging log

    if eventID == "" {
        http.Error(w, "Event ID is required", http.StatusBadRequest)
        return
    }

    var event Event
    if err := db.Preload("RSVPs").First(&event, eventID).Error; err != nil {
        log.Printf("Error fetching event with ID %s: %v", eventID, err)
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(event)
}



func UpdateEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var event Event
    if err := db.First(&event, params["id"]).Error; err != nil {
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }
    var updatedEvent Event
    if err := json.NewDecoder(r.Body).Decode(&updatedEvent); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    // Update event fields
    event.Name = updatedEvent.Name
    event.Date = updatedEvent.Date
    event.Time = updatedEvent.Time
    event.Location = updatedEvent.Location
    event.Description = updatedEvent.Description
    db.Save(&event)
    json.NewEncoder(w).Encode(event)
}

func DeleteEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var event Event
    if err := db.Delete(&event, params["id"]).Error; err != nil {
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode("Event deleted")
}

func RSVPEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var rsvp RSVP
    err := json.NewDecoder(r.Body).Decode(&rsvp)
    if err != nil || (rsvp.Response != "accepted" && rsvp.Response != "declined") {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }
    
    eventID, err := strconv.ParseUint(params["id"], 10, 32)
    if err != nil {
        http.Error(w, "Invalid event ID", http.StatusBadRequest)
        return
    }
    rsvp.EventID = uint(eventID)
    
    db.Create(&rsvp)
    json.NewEncoder(w).Encode(rsvp)
}

// Set up the routes
func InitializeRoutes() *mux.Router {
    router := mux.NewRouter()

    router.HandleFunc("/events", CreateEvent).Methods("POST")
    router.HandleFunc("/events", GetEvents).Methods("GET")
    router.HandleFunc("/events/{id}", GetEvent).Methods("GET")
    router.HandleFunc("/events/{id}", UpdateEvent).Methods("PUT")
    router.HandleFunc("/events/{id}", DeleteEvent).Methods("DELETE")
    router.HandleFunc("/events/{id}/rsvp", RSVPEvent).Methods("POST")

    return router
}

// Main function
func main() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }

    // Initialize the database
    InitDB()

    // Set up routes
    router := InitializeRoutes()

    // Create a new CORS handler
    c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"}, // Allow all origins
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Content-Type"},
    })

    // Wrap the router with CORS middleware
    handler := c.Handler(router)

    // Start the server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8001"
    }
    log.Printf("Server is running on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}
