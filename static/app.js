const API_BASE = '/api';

// Load blocked numbers on page load
document.addEventListener('DOMContentLoaded', function() {
    loadBlockedNumbers();
});

// Handle form submission
document.getElementById('blockPhoneForm').addEventListener('submit', function(e) {
    e.preventDefault();
    addBlockedPhone();
});

async function loadBlockedNumbers() {
    const loading = document.getElementById('loading');
    const table = document.getElementById('blockedTable');
    const emptyState = document.getElementById('emptyState');
    
    loading.style.display = 'block';
    table.style.display = 'none';
    emptyState.style.display = 'none';

    try {
        const response = await fetch(`${API_BASE}/blocked-phones`);
        if (!response.ok) {
            throw new Error('Failed to load blocked numbers');
        }
        
        const blockedNumbers = await response.json();
        const tbody = document.getElementById('blockedTableBody');
        tbody.innerHTML = '';

        if (blockedNumbers && blockedNumbers.length > 0) {
            blockedNumbers.forEach(phone => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td class="phone-number">${phone.phone_number}</td>
                    <td class="reason">${phone.reason || 'No reason provided'}</td>
                    <td>${phone.blocked_by || 'System'}</td>
                    <td class="blocked-date">${new Date(phone.blocked_date).toLocaleString()}</td>
                    <td>
                        <button class="btn btn-danger" onclick="removeBlockedPhone(${phone.id})">
                            Remove
                        </button>
                    </td>
                `;
                tbody.appendChild(row);
            });
            table.style.display = 'table';
        } else {
            emptyState.style.display = 'block';
        }
    } catch (error) {
        console.error('Error loading blocked numbers:', error);
        showAlert('Failed to load blocked numbers. Please try again.', 'error');
        emptyState.style.display = 'block';
    } finally {
        loading.style.display = 'none';
    }
}

async function addBlockedPhone() {
    const phoneNumber = document.getElementById('phoneNumber').value.trim();
    const reason = document.getElementById('reason').value.trim();
    const blockedBy = document.getElementById('blockedBy').value.trim();

    if (!phoneNumber) {
        showAlert('Phone number is required', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/blocked-phones`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                phone_number: phoneNumber,
                reason: reason,
                blocked_by: blockedBy
            })
        });

        if (response.ok) {
            showAlert('Phone number successfully added to blocked list', 'success');
            document.getElementById('blockPhoneForm').reset();
            loadBlockedNumbers();
        } else {
            const errorText = await response.text();
            showAlert(errorText || 'Failed to block phone number', 'error');
        }
    } catch (error) {
        console.error('Error adding blocked phone:', error);
        showAlert('Failed to block phone number. Please try again.', 'error');
    }
}
async function checkPhoneNumber() {
    const phoneNumber = document.getElementById('checkPhoneNumber').value.trim();
    const resultDiv = document.getElementById('checkResult');

    if (!phoneNumber) {
        showAlert('Please enter a phone number to check', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/check-phone`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                phone_number: phoneNumber
            })
        });

        if (response.ok) {
            const result = await response.json();
            
            if (result.is_blocked) {
                resultDiv.innerHTML = `
                    <div class="alert alert-warning" style="margin-top: 15px;">
                        <h4>⚠️ WARNING - BLOCKED NUMBER</h4>
                        <p><strong>Phone Number:</strong> ${result.phone_number}</p>
                        <p><strong>Reason:</strong> ${result.reason}</p>
                        <p><strong>Blocked By:</strong> ${result.blocked_by}</p>
                        <p><strong>Date Blocked:</strong> ${result.blocked_date}</p>
                        <p><strong>⚠️ DO NOT CASH CHECK FOR THIS PHONE NUMBER</strong></p>
                    </div>
                `;
            } else {
                resultDiv.innerHTML = `
                    <div class="alert alert-success" style="margin-top: 15px;">
                        <h4>✅ Safe to Proceed</h4>
                        <p><strong>Phone Number:</strong> ${result.phone_number}</p>
                        <p>This phone number is not in the blocked list.</p>
                    </div>
                `;
            }
        } else {
            const errorText = await response.text();
            showAlert(errorText || 'Failed to check phone number', 'error');
            resultDiv.innerHTML = '';
        }
    } catch (error) {
        console.error('Error checking phone number:', error);
        showAlert('Failed to check phone number. Please try again.', 'error');
        resultDiv.innerHTML = '';
    }
}

function showAlert(message, type) {
    const alert = document.createElement('div');
    alert.className = `alert alert-${type}`;
    alert.innerHTML = message;

    const container = document.getElementById('alertContainer');
    if (container) {
        container.appendChild(alert);
    }

    setTimeout(() => {
        alert.remove();
    }, 5000);
}


// Format phone number as user types
document.getElementById('phoneNumber').addEventListener('input', function(e) {
    formatPhoneInput(e.target);
});

document.getElementById('checkPhoneNumber').addEventListener('input', function(e) {
    formatPhoneInput(e.target);
});

function formatPhoneInput(input) {
    let value = input.value.replace(/\D/g, '');
    
    if (value.length >= 6) {
        value = value.replace(/(\d{3})(\d{3})(\d{0,4})/, '($1) $2-$3');
    } else if (value.length >= 3) {
        value = value.replace(/(\d{3})(\d{0,3})/, '($1) $2');
    }
    
    input.value = value;
}

// Clear check result when input changes
document.getElementById('checkPhoneNumber').addEventListener('input', function() {
    document.getElementById('checkResult').innerHTML = '';
});


async function removeBlockedPhone(id) {
    if (!confirm('Are you sure you want to remove this phone number from the blocked list?')) return;

    try {
        const response = await fetch(`${API_BASE}/blocked-phones/${id}`, { method: 'DELETE' });

        if (response.ok) {
            showAlert('Phone number removed from blocked list', 'success');
            loadBlockedNumbers();
        } else {
            const errorText = await response.text();
            showAlert(errorText || 'Failed to remove phone number', 'error');
        }
    } catch (error) {
        console.error('Error removing blocked phone:', error);
        showAlert('Failed to remove phone number. Please try again.', 'error');
    }
}