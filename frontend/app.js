const API_URL = "http://localhost:8080/api/v1";

// Helper: Show Toast
function showToast(message, isError = false) {
    const toastEl = document.getElementById('liveToast');
    const toastBody = document.getElementById('toastMessage');
    toastBody.innerText = message;
    if (isError) {
        toastEl.classList.add('text-bg-danger');
        toastEl.classList.remove('text-bg-success');
    } else {
        toastEl.classList.add('text-bg-success');
        toastEl.classList.remove('text-bg-danger');
    }
    const toast = new bootstrap.Toast(toastEl);
    toast.show();
}

// Helper: Fetch Wrapper
async function apiRequest(endpoint, method = "GET", body = null) {
    const options = {
        method,
        headers: { "Content-Type": "application/json" }
    };
    if (body) options.body = JSON.stringify(body);

    try {
        const res = await fetch(`${API_URL}${endpoint}`, options);
        if (!res.ok) {
            const errData = await res.json();
            throw new Error(errData.error || "Request failed");
        }
        // Patch often returns no content but status 200/204
        if (method === "PATCH" && res.status === 200) return {};

        return await res.json();
    } catch (err) {
        showToast(err.message, true);
        throw err;
    }
}

// --- Navigation Logic ---
document.querySelectorAll('.nav-link').forEach(link => {
    link.addEventListener('click', (e) => {
        e.preventDefault();
        // UI Tabs
        document.querySelectorAll('.nav-link').forEach(n => n.classList.remove('active'));
        e.target.classList.add('active');
        // Sections
        document.querySelectorAll('.section-card').forEach(s => s.classList.remove('active'));
        const targetId = e.target.getAttribute('data-section');
        document.getElementById(targetId).classList.add('active');
    });
});

// Set default date for search
document.getElementById('searchDate').valueAsDate = new Date();


// --- 1. PASSENGERS ---

document.getElementById('createPassengerForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fd = new FormData(e.target);
    const data = {
        user_id: parseInt(fd.get('user_id')),
        passport_no: fd.get('passport_no'),
        phone: fd.get('phone')
    };

    try {
        const res = await apiRequest('/passengers', 'POST', data);
        showToast(`Passenger created! ID: ${res.id}`);
        document.getElementById('currentPassengerId').value = res.id;
        e.target.reset();
    } catch (e) { console.error(e); }
});

document.getElementById('btnGetPassenger').addEventListener('click', async () => {
    const id = document.getElementById('searchPassengerId').value;
    if (!id) return;
    try {
        const res = await apiRequest(`/passengers/${id}`);
        const resultDiv = document.getElementById('passengerResult');
        resultDiv.classList.remove('d-none');
        resultDiv.innerText = `Найден: Passport ${res.passport_no}, Phone: ${res.phone}`;
    } catch (e) { console.error(e); }
});


// --- 2. FLIGHTS & SEARCH ---

document.getElementById('searchFlightForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const origin = document.getElementById('searchOrigin').value;
    const dest = document.getElementById('searchDest').value;
    const date = document.getElementById('searchDate').value;

    try {
        const flights = await apiRequest(`/flights?origin=${origin}&destination=${dest}&date=${date}`);
        renderFlights(flights);
    } catch (e) { console.error(e); }
});

function renderFlights(flights) {
    const container = document.getElementById('flightsList');
    container.innerHTML = '';

    if (!flights || flights.length === 0) {
        container.innerHTML = '<div class="col-12 text-center text-muted mt-3">Рейсов не найдено</div>';
        return;
    }

    flights.forEach(f => {
        const depTime = new Date(f.departure_time).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
        const arrTime = new Date(f.arrival_time).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});

        const col = document.createElement('div');
        col.className = 'col-md-6 mb-3';
        col.innerHTML = `
            <div class="card h-100 border-primary">
                <div class="card-body">
                    <div class="d-flex justify-content-between align-items-center mb-2">
                        <h5 class="card-title text-primary">${f.flight_no}</h5>
                        <span class="badge ${getStatusBadge(f.status)}">${f.status}</span>
                    </div>
                    <p class="card-text mb-1">
                        <strong>${f.origin}</strong> (${depTime}) &rarr; 
                        <strong>${f.destination}</strong> (${arrTime})
                    </p>
                    <small class="text-muted">Seats: ${f.total_seats}</small>
                    <div class="mt-3 text-end">
                        <button class="btn btn-sm btn-outline-primary" onclick="bookFlight(${f.id})">Забронировать ($150)</button>
                    </div>
                    <div class="mt-1 text-end text-muted" style="font-size: 0.8rem">
                        Ver: ${f.Version} (Use for Admin update)
                    </div>
                </div>
            </div>
        `;
        container.appendChild(col);
    });
}

function getStatusBadge(status) {
    const map = {
        'SCHEDULED': 'bg-success',
        'DELAYED': 'bg-warning text-dark',
        'CANCELLED': 'bg-danger',
        'BOARDING': 'bg-info text-dark',
        'DEPARTED': 'bg-secondary'
    };
    return map[status] || 'bg-light text-dark';
}

// --- 3. BOOKINGS ---

window.bookFlight = async (flightId) => {
    const passengerId = document.getElementById('currentPassengerId').value;
    if (!passengerId) {
        showToast("Введите ID пассажира в верхнем меню", true);
        return;
    }

    if (!confirm(`Бронируем рейс ID ${flightId} для пассажира ${passengerId}?`)) return;

    try {
        const res = await apiRequest('/bookings', 'POST', {
            flight_id: parseInt(flightId),
            passenger_id: parseInt(passengerId)
        });
        showToast(`Билет успешно куплен! ID билета: ${res.id}`);
        // Switch to booking tab and show ticket
        document.querySelector('[data-section="bookings"]').click();
        document.getElementById('ticketIdInput').value = res.id;
        document.getElementById('btnGetTicket').click();
    } catch (e) { console.error(e); }
};

document.getElementById('btnGetTicket').addEventListener('click', async () => {
    const id = document.getElementById('ticketIdInput').value;
    if(!id) return;

    try {
        const data = await apiRequest(`/bookings/${id}`);
        const t = data.ticket;
        const f = data.flight || {};

        document.getElementById('t-id').innerText = t.id;
        document.getElementById('t-status').innerText = t.status;
        document.getElementById('t-price').innerText = t.price;

        if (f.flight_no) {
            document.getElementById('t-flight').innerText = `${f.flight_no}`;
            document.getElementById('t-route').innerText = `${f.origin} -> ${f.destination}`;
        } else {
            document.getElementById('t-flight').innerText = `Flight ID: ${t.flight_id}`;
            document.getElementById('t-route').innerText = '-';
        }

        const card = document.getElementById('ticketDetailsCard');
        card.classList.remove('d-none');

        // Setup cancel button
        const btnCancel = document.getElementById('btnCancelTicket');
        btnCancel.onclick = async () => {
            if(!confirm('Отменить этот билет?')) return;
            try {
                await apiRequest(`/bookings/${t.id}/cancel`, 'POST');
                showToast('Билет отменен');
                document.getElementById('btnGetTicket').click(); // Refresh
            } catch(e) { console.error(e); }
        }

    } catch (e) { console.error(e); }
});

// --- 4. ADMIN & OPS ---

document.getElementById('createFlightForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fd = new FormData(e.target);

    // ISO 8601 formatting needed for Go time.Time
    const dep = new Date(fd.get('departure_time')).toISOString();
    const arr = new Date(fd.get('arrival_time')).toISOString();

    const data = {
        flight_no: fd.get('flight_no'),
        origin: fd.get('origin'),
        destination: fd.get('destination'),
        departure_time: dep,
        arrival_time: arr
    };

    try {
        await apiRequest('/flights', 'POST', data);
        showToast('Рейс создан успешно');
        e.target.reset();
    } catch(e) { console.error(e); }
});

document.getElementById('createGateForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fd = new FormData(e.target);
    const data = {
        terminal_id: parseInt(fd.get('terminal_id')),
        code: fd.get('code')
    };
    try {
        await apiRequest('/gates', 'POST', data);
        showToast('Гейт создан');
        e.target.reset();
    } catch(e) { console.error(e); }
});

document.getElementById('assignGateForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fd = new FormData(e.target);
    const data = {
        flight_id: parseInt(fd.get('flight_id')),
        gate_id: parseInt(fd.get('gate_id'))
    };
    try {
        await apiRequest('/gates/assign', 'POST', data);
        showToast('Гейт назначен на рейс');
        e.target.reset();
    } catch(e) { console.error(e); }
});

document.getElementById('updateStatusForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fd = new FormData(e.target);
    const id = fd.get('id');
    const data = {
        status: fd.get('status'),
        version: parseInt(fd.get('version'))
    };

    try {
        await apiRequest(`/flights/${id}/status`, 'PATCH', data);
        showToast('Статус рейса обновлен');
    } catch(e) { console.error(e); }
});