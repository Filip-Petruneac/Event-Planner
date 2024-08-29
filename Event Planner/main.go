package main

import (
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

var db *gorm.DB
var templates *template.Template

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

// Load templates
func LoadTemplates() {
    templates = template.Must(template.ParseFiles("templates/index.html", "templates/event.html"))
}

// API Handlers

// CreateEvent handles POST requests to create a new event
func CreateEvent(w http.ResponseWriter, r *http.Request) {
    var event Event

    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    if event.Date.IsZero() {
        http.Error(w, "Invalid date", http.StatusBadRequest)
        return
    }

    if err := db.Create(&event).Error; err != nil {
        http.Error(w, "Failed to create event", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(event)
}

// GetEvents handles GET requests to retrieve events
func GetEvents(w http.ResponseWriter, r *http.Request) {
    filter := r.URL.Query().Get("filter")
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
    eventIDStr := params["id"]

    eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid event ID", http.StatusBadRequest)
        return
    }

    var event Event
    if err := db.Preload("RSVPs").First(&event, eventID).Error; err != nil {
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

// Page Handlers

// Render the main event list page
func RenderIndex(w http.ResponseWriter, r *http.Request) {
    var events []Event
    db.Find(&events)
    err := templates.ExecuteTemplate(w, "index.html", events)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func RenderEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    eventIDStr := params["id"]

    eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid event ID", http.StatusBadRequest)
        return
    }

    var event Event
    if err := db.Preload("RSVPs").First(&event, eventID).Error; err != nil {
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }

    err = templates.ExecuteTemplate(w, "event.html", event)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// Set up the routes
func InitializeRoutes() *mux.Router {
    router := mux.NewRouter()

    // API Routes
    router.HandleFunc("/events", CreateEvent).Methods("POST")
    router.HandleFunc("/events", GetEvents).Methods("GET")
    router.HandleFunc("/events/{id}", GetEvent).Methods("GET")
    router.HandleFunc("/events/{id}", UpdateEvent).Methods("PUT")
    router.HandleFunc("/events/{id}", DeleteEvent).Methods("DELETE")
    router.HandleFunc("/events/{id}/rsvp", RSVPEvent).Methods("POST")

    // Frontend Routes
    router.HandleFunc("/", RenderIndex).Methods("GET")
    router.HandleFunc("/events/{id}", RenderEvent).Methods("GET")

    // Static files
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

    return router
}

// Main function
func main() {
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }

    InitDB()
    LoadTemplates()
    router := InitializeRoutes()

    c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Content-Type"},
    })

    handler := c.Handler(router)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8001"
    }
    log.Printf("Server is running on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}
