document.addEventListener('DOMContentLoaded', (event) => {
    const createEventBtn = document.getElementById('create-event-btn');
    const createEventModal = document.getElementById('create-event-modal');
    const cancelCreateEventBtn = document.getElementById('cancel-create-event');
    const newEventForm = document.getElementById('new-event-form');
    const newRSVPForm = document.getElementById('new-rsvp-form');

    if (createEventBtn && createEventModal) {
        createEventBtn.addEventListener('click', () => {
            createEventModal.classList.remove('hidden');
            createEventModal.classList.add('flex');
        });
    }

    if (cancelCreateEventBtn && createEventModal) {
        cancelCreateEventBtn.addEventListener('click', () => {
            createEventModal.classList.add('hidden');
            createEventModal.classList.remove('flex');
        });
    }

    if (newEventForm) {
        newEventForm.addEventListener('submit', createEvent);
    }

    if (newRSVPForm) {
        newRSVPForm.addEventListener('submit', submitRSVP);
    }
});

async function createEvent(e) {
    e.preventDefault();
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
    const listItem = document.createElement('li');
    listItem.className = 'mb-2';
    listItem.innerHTML = `User ID: ${newRSVP.user_id} - Response: <span class="font-semibold ${newRSVP.response === 'accepted' ? 'text-green-600' : 'text-red-600'}">${newRSVP.response}</span>`;
    
    if (rsvpList.firstChild.textContent === 'No RSVPs yet.') {
        rsvpList.innerHTML = '';
    }
    
    rsvpList.appendChild(listItem);
}