document.addEventListener('DOMContentLoaded', (event) => {
    const createEventBtn = document.getElementById('create-event-btn');
    const createEventModal = document.getElementById('create-event-modal');
    const cancelCreateEventBtn = document.getElementById('cancel-create-event');
    const newEventForm = document.getElementById('new-event-form');
    const newRSVPForm = document.getElementById('new-rsvp-form');

    if (createEventBtn && createEventModal) {
        createEventBtn.addEventListener('click', () => {
            console.log('Create event button clicked');
            createEventModal.classList.remove('hidden');
            createEventModal.classList.add('flex');
        });
    } else {
        console.error('Create event button or modal not found');
    }

    if (cancelCreateEventBtn && createEventModal) {
        cancelCreateEventBtn.addEventListener('click', () => {
            console.log('Cancel button clicked');
            createEventModal.classList.add('hidden');
            createEventModal.classList.remove('flex');
        });
    } else {
        console.error('Cancel button or modal not found');
    }

    if (newEventForm) {
        newEventForm.addEventListener('submit', createEvent);
    } else {
        console.error('New event form not found');
    }

    if (newRSVPForm) {
        newRSVPForm.addEventListener('submit', submitRSVP);
    } else {
        console.error('RSVP form not found');
    }
});

async function createEvent(e) {
    e.preventDefault();
    console.log('Create event function called');
    const formData = new FormData(e.target);
    const eventData = Object.fromEntries(formData.entries());

    try {
        const response = await fetch('/events', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(eventData),
        });

        if (response.ok) {
            console.log('Event created successfully');
            window.location.reload();
        } else {
            console.error('Failed to create event');
            alert('Failed to create event. Please try again.');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred. Please try again.');
    }
}

async function submitRSVP(e) {
    e.preventDefault();
    console.log('Submit RSVP function called');
    const formData = new FormData(e.target);
    const rsvpData = Object.fromEntries(formData.entries());
    const eventId = document.getElementById('event-id').value;

    try {
        const response = await fetch(`/events/${eventId}/rsvp`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(rsvpData),
        });

        if (response.ok) {
            const newRSVP = await response.json();
            updateRSVPList(newRSVP);
            e.target.reset();
        } else {
            console.error('Failed to submit RSVP');
            alert('Failed to submit RSVP. Please try again.');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred. Please try again.');
    }
}

function updateRSVPList(newRSVP) {
    const rsvpList = document.getElementById('rsvp-list');
    if (rsvpList) {
        const listItem = document.createElement('li');
        listItem.className = 'mb-2';
        listItem.innerHTML = `User ID: ${newRSVP.user_id} - Response: <span class="font-semibold ${newRSVP.response === 'accepted' ? 'text-green-600' : 'text-red-600'}">${newRSVP.response}</span>`;
        
        if (rsvpList.firstChild && rsvpList.firstChild.textContent === 'No RSVPs yet.') {
            rsvpList.innerHTML = '';
        }
        
        rsvpList.appendChild(listItem);
    } else {
        console.error('RSVP list not found');
    }
}