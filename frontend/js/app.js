const apiUrl = 'http://localhost:8001/events'; 

document.addEventListener('DOMContentLoaded', () => {
    if (window.location.pathname === '/event.html') {
        const urlParams = new URLSearchParams(window.location.search);
        const eventId = urlParams.get('id');
        if (eventId) {
            loadEventDetails(eventId);
        } else {
            console.error('Event ID is missing in URL');
        }
    } else {
        loadEvents();
        document.getElementById('eventForm').addEventListener('submit', (e) => {
            e.preventDefault();
            createEvent();
        });    

        if (document.getElementById('rsvpForm')) {
            document.getElementById('rsvpForm').addEventListener('submit', (e) => {
                e.preventDefault();
                submitRSVP();
            });
        }
    }
});

// Function to format date and time
function formatDateTime(date, time) {
    const dateTimeString = `${date}T${time}`;
    const dateTime = new Date(dateTimeString);
    return dateTime.toISOString(); // Example format: '2024-08-28T14:30:00.000Z'
}

// Example of logging events to check structure
function loadEvents() {
    fetch(apiUrl)
        .then(response => response.json())
        .then(events => {
            console.log("Loaded events:", events); // Log the events
            const eventList = document.getElementById('eventList');
            eventList.innerHTML = '';
            events.forEach(event => {
                const li = document.createElement('li');
                li.textContent = `${event.name} - ${event.date} - ${event.location}`;
                
                // Add a click event listener to each list item
                li.addEventListener('click', () => {
                    console.log(`Redirecting to event.html?id=${event.id}`); // Log the event ID
                    window.location.href = `event.html?id=${event.id}`;
                });

                eventList.appendChild(li);
            });
        })
        .catch(error => {
            console.error('Error loading events:', error);
        });
}

// Function to create a new event
function createEvent() {
    const event = {
        name: document.getElementById('name').value,
        date: formatDateTime(document.getElementById('date').value, document.getElementById('time').value),
        location: document.getElementById('location').value,
        description: document.getElementById('description').value
    };

    console.log("Creating event with data:", event); 

    fetch(apiUrl, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(event)
    })
    .then(response => {
        console.log("Response status:", response.status);

        if (!response.ok) {
            return response.text().then(text => {
                console.error(`Error response: ${response.status} - ${text}`);
                throw new Error(`Network response was not ok: ${response.status} - ${text}`);
            });
        }
        return response.json();
    })
    .then(data => {
        console.log("Event creation successful:", data);
        hideCreateEventForm();
        loadEvents(); 
    })
    .catch(error => {
        console.error('Error creating event:', error);
        alert(`Failed to create event. Please try again. Error: ${error.message}`);
    });
}

function submitRSVP() {
    // Get event ID from URL
    const urlParams = new URLSearchParams(window.location.search);
    const eventId = urlParams.get('id');

    // Check if eventId is valid
    if (!eventId) {
        console.error('Event ID is missing in URL');
        return;
    }

    // Get the selected RSVP value
    const selectedRadio = document.querySelector('input[name="rsvp"]:checked');
    if (!selectedRadio) {
        console.error('No RSVP option selected');
        return;
    }
    const rsvpValue = selectedRadio.value;

    console.log(`Submitting RSVP for event ID: ${eventId}`); // Added logging for debugging

    // Send RSVP to server
    fetch(`http://localhost:8001/events/${eventId}/rsvp`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ response: rsvpValue })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(() => {
        loadEventDetails(eventId);
    })
    .catch(error => {
        console.error('Error submitting RSVP:', error);
    });
}

// Function to load event details
function loadEventDetails(eventId) {
    console.log(`Loading details for event ID: ${eventId}`);

    fetch(`http://localhost:8001/events/${eventId}`)
        .then(response => response.json())
        .then(event => {
            document.getElementById('eventName').textContent = event.name;
            document.getElementById('eventDate').textContent = `Date: ${event.date}`;
            document.getElementById('eventTime').textContent = `Time: ${event.time}`;
            document.getElementById('eventLocation').textContent = `Location: ${event.location}`;
            document.getElementById('eventDescription').textContent = `Description: ${event.description}`;

            const attendeesList = document.getElementById('attendeesList');
            attendeesList.innerHTML = '';
            event.rsvps.forEach(rsvp => {
                const li = document.createElement('li');
                li.textContent = `User ${rsvp.user_id} - ${rsvp.response}`;
                attendeesList.appendChild(li);
            });
        })
        .catch(error => console.error('Error fetching event details:', error));
}

// Function to show create event form
function showCreateEventForm() {
    document.getElementById('create-event-form').style.display = 'block';
}

// Function to hide create event form
function hideCreateEventForm() {
    document.getElementById('create-event-form').style.display = 'none';
}
