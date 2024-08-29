package main

import (
    "encoding/json"
    "html/template"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "github.com/sirupsen/logrus"
)

// DB struct to encapsulate the database connection
type DB struct {
    *gorm.DB
}

// Logger struct to encapsulate the logger
type Logger struct {
    *logrus.Logger
}

// App struct to encapsulate dependencies
type App struct {
    DB     *DB
    Logger *Logger
}

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
func InitDB() *DB {
    dbConn, err := gorm.Open(sqlite.Open("events.db"), &gorm.Config{})
    if err != nil {
        logrus.Fatal("Failed to connect to database", err)
    }
    db := &DB{DB: dbConn}
    db.AutoMigrate(&Event{}, &RSVP{})
    return db
}

// Load templates
func LoadTemplates() {
    templates = template.Must(template.ParseFiles("templates/index.html", "templates/event.html"))
}

// API Handlers

// CreateEvent handles POST requests to create a new event
func (app *App) CreateEvent(w http.ResponseWriter, r *http.Request) {
    var event Event

    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        app.Logger.Errorf("CreateEvent: Invalid input - %v", err)
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    if event.Date.IsZero() {
        app.Logger.Warn("CreateEvent: Invalid date")
        http.Error(w, "Invalid date", http.StatusBadRequest)
        return
    }

    if err := app.DB.Create(&event).Error; err != nil {
        app.Logger.Errorf("CreateEvent: Failed to create event - %v", err)
        http.Error(w, "Failed to create event", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(event)
}

// GetEvents handles GET requests to retrieve events
func (app *App) GetEvents(w http.ResponseWriter, r *http.Request) {
    filter := r.URL.Query().Get("filter")
    var events []Event

    switch filter {
    case "upcoming":
        app.DB.Preload("RSVPs").Where("date >= ?", time.Now()).Find(&events)
    case "past":
        app.DB.Preload("RSVPs").Where("date < ?", time.Now()).Find(&events)
    default:
        app.DB.Preload("RSVPs").Find(&events)
    }

    json.NewEncoder(w).Encode(events)
}

// GetEvent retrieves an event by its ID
func (app *App) GetEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    eventIDStr := params["id"]

    eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
    if err != nil {
        app.Logger.Errorf("GetEvent: Invalid event ID - %v", err)
        http.Error(w, "Invalid event ID", http.StatusBadRequest)
        return
    }

    var event Event
    if err := app.DB.Preload("RSVPs").First(&event, eventID).Error; err != nil {
        app.Logger.Errorf("GetEvent: Event not found - %v", err)
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(event)
}

// UpdateEvent updates an event by its ID
func (app *App) UpdateEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var event Event
    if err := app.DB.First(&event, params["id"]).Error; err != nil {
        app.Logger.Errorf("UpdateEvent: Event not found - %v", err)
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }
    var updatedEvent Event
    if err := json.NewDecoder(r.Body).Decode(&updatedEvent); err != nil {
        app.Logger.Errorf("UpdateEvent: Invalid input - %v", err)
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    event.Name = updatedEvent.Name
    event.Date = updatedEvent.Date
    event.Time = updatedEvent.Time
    event.Location = updatedEvent.Location
    event.Description = updatedEvent.Description
    app.DB.Save(&event)
    json.NewEncoder(w).Encode(event)
}

// DeleteEvent deletes an event by its ID
func (app *App) DeleteEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var event Event
    if err := app.DB.Delete(&event, params["id"]).Error; err != nil {
        app.Logger.Errorf("DeleteEvent: Event not found - %v", err)
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode("Event deleted")
}

// RSVPEvent handles RSVPs for events
func (app *App) RSVPEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var rsvp RSVP
    err := json.NewDecoder(r.Body).Decode(&rsvp)
    if err != nil || (rsvp.Response != "accepted" && rsvp.Response != "declined") {
        app.Logger.Errorf("RSVPEvent: Invalid input - %v", err)
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    eventID, err := strconv.ParseUint(params["id"], 10, 32)
    if err != nil {
        app.Logger.Errorf("RSVPEvent: Invalid event ID - %v", err)
        http.Error(w, "Invalid event ID", http.StatusBadRequest)
        return
    }
    rsvp.EventID = uint(eventID)

    app.DB.Create(&rsvp)
    json.NewEncoder(w).Encode(rsvp)
}

// Page Handlers

// RenderIndex renders the main event list page
func (app *App) RenderIndex(w http.ResponseWriter, r *http.Request) {
    var events []Event
    app.DB.Find(&events)
    err := templates.ExecuteTemplate(w, "index.html", events)
    if err != nil {
        app.Logger.Errorf("RenderIndex: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// RenderEvent renders a specific event page
func (app *App) RenderEvent(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    eventIDStr := params["id"]

    eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
    if err != nil {
        app.Logger.Errorf("RenderEvent: Invalid event ID - %v", err)
        http.Error(w, "Invalid event ID", http.StatusBadRequest)
        return
    }

    var event Event
    if err := app.DB.Preload("RSVPs").First(&event, eventID).Error; err != nil {
        app.Logger.Errorf("RenderEvent: Event not found - %v", err)
        http.Error(w, "Event not found", http.StatusNotFound)
        return
    }

    err = templates.ExecuteTemplate(w, "event.html", event)
    if err != nil {
        app.Logger.Errorf("RenderEvent: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// InitializeRoutes sets up the router with routes and handlers
func InitializeRoutes(app *App) *mux.Router {
    router := mux.NewRouter()

    // API Routes
    router.HandleFunc("/events", app.CreateEvent).Methods("POST")
    router.HandleFunc("/events", app.GetEvents).Methods("GET")
    router.HandleFunc("/events/{id}", app.GetEvent).Methods("GET")
    router.HandleFunc("/events/{id}", app.UpdateEvent).Methods("PUT")
    router.HandleFunc("/events/{id}", app.DeleteEvent).Methods("DELETE")
    router.HandleFunc("/events/{id}/rsvp", app.RSVPEvent).Methods("POST")

    // Frontend Routes
    router.HandleFunc("/", app.RenderIndex).Methods("GET")
    router.HandleFunc("/events/{id}", app.RenderEvent).Methods("GET")

    // Static files
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

    return router
}

// Main function
func main() {
    if err := godotenv.Load(); err != nil {
        logrus.Print("No .env file found")
    }

    // Initialize the logger
    logger := &Logger{Logger: logrus.New()}
    logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
    logger.SetLevel(logrus.InfoLevel)

    // Initialize the database
    db := InitDB()
    LoadTemplates()
    app := &App{DB: db, Logger: logger}
    router := InitializeRoutes(app)

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
    logger.Infof("Server is running on port %s", port)
    logger.Fatal(http.ListenAndServe(":"+port, handler))
}
