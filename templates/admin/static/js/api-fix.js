// Fix for Admin Panel API Authentication
// This fixes the issue where templates were using localStorage tokens instead of cookies

document.addEventListener('DOMContentLoaded', function() {
    // Override all fetch calls to use cookies instead of localStorage tokens
    const originalFetch = window.fetch;
    window.fetch = function(url, options = {}) {
        // If this is an API call to our server, ensure cookies are included
        if (url.startsWith('/api/')) {
            options.credentials = 'include';
            
            // Remove Authorization header if it's trying to use localStorage
            if (options.headers && options.headers['Authorization']) {
                delete options.headers['Authorization'];
            }
        }
        
        return originalFetch(url, options);
    };

    // Fix logout function if it exists
    if (typeof window.logout === 'undefined') {
        window.logout = function() {
            if (confirm('Are you sure you want to logout?')) {
                fetch('/api/auth/logout', {
                    method: 'POST',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                })
                .then(() => {
                    localStorage.clear();
                    window.location.href = '/admin/login';
                })
                .catch(() => {
                    localStorage.clear();
                    window.location.href = '/admin/login';
                });
            }
        };
    }

    // Fix maintenance function
    window.runMaintenance = function() {
        if (confirm('Are you sure you want to run system maintenance?')) {
            fetch('/api/admin/maintenance', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Maintenance completed successfully');
                    location.reload();
                } else {
                    alert('Maintenance failed: ' + (data.message || 'Unknown error'));
                }
            })
            .catch(error => {
                alert('Error: ' + error.message);
            });
        }
    };

    // Fix user management functions
    window.createUser = function() {
        const userData = {
            name: document.getElementById('userName').value,
            email: document.getElementById('userEmail').value,
            password: document.getElementById('userPassword').value
        };

        fetch('/api/users', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        })
        .then(response => response.json())
        .then(data => {
            if (data.id) {
                alert('User created successfully');
                location.reload();
            } else {
                alert('Error: ' + (data.message || 'Failed to create user'));
            }
        })
        .catch(error => {
            alert('Error: ' + error.message);
        });
    };

    window.updateUser = function(userId) {
        const userData = {
            name: document.getElementById('editUserName').value,
            email: document.getElementById('editUserEmail').value
        };

        fetch(`/api/users/${userId}`, {
            method: 'PUT',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        })
        .then(response => response.json())
        .then(data => {
            if (data.id) {
                alert('User updated successfully');
                location.reload();
            } else {
                alert('Error: ' + (data.message || 'Failed to update user'));
            }
        })
        .catch(error => {
            alert('Error: ' + error.message);
        });
    };

    window.deleteUser = function(userId) {
        if (confirm('Are you sure you want to delete this user?')) {
            fetch(`/api/users/${userId}`, {
                method: 'DELETE',
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('User deleted successfully');
                    location.reload();
                } else {
                    alert('Error: ' + (data.message || 'Failed to delete user'));
                }
            })
            .catch(error => {
                alert('Error: ' + error.message);
            });
        }
    };

    // Fix deleted users functions
    window.restoreUser = function(userId) {
        if (confirm('Are you sure you want to restore this user?')) {
            fetch(`/api/admin/users/${userId}/restore`, {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                }
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('User restored successfully');
                    location.reload();
                } else {
                    alert('Error: ' + (data.message || 'Failed to restore user'));
                }
            })
            .catch(error => {
                alert('Error: ' + error.message);
            });
        }
    };

    window.permanentDeleteUser = function(userId) {
        if (confirm('Are you sure you want to PERMANENTLY delete this user? This action cannot be undone!')) {
            fetch(`/api/admin/users/${userId}/permanent-delete`, {
                method: 'DELETE',
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('User permanently deleted');
                    location.reload();
                } else {
                    alert('Error: ' + (data.message || 'Failed to delete user'));
                }
            })
            .catch(error => {
                alert('Error: ' + error.message);
            });
        }
    };

    // Fix export logs function
    window.exportLogs = function() {
        fetch('/api/admin/logs?export=true', {
            credentials: 'include'
        })
        .then(response => response.blob())
        .then(blob => {
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.style.display = 'none';
            a.href = url;
            a.download = 'activity_logs.csv';
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
        })
        .catch(error => {
            alert('Error exporting logs: ' + error.message);
        });
    };

    console.log('Admin Panel API Fix loaded - Now using cookie authentication');
}); 