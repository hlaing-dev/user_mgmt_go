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
}); 