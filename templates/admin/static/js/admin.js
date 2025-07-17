// Admin Panel JavaScript Functionality

(function() {
    'use strict';

    // Initialize admin panel when DOM is loaded
    document.addEventListener('DOMContentLoaded', function() {
        initializeSidebar();
        initializeTooltips();
        initializeActiveNavigation();
        initializeAutoRefresh();
        checkAuthenticationStatus();
    });

    // Sidebar functionality
    function initializeSidebar() {
        const sidebarCollapse = document.getElementById('sidebarCollapse');
        const sidebar = document.getElementById('sidebar');
        const content = document.getElementById('content');

        if (sidebarCollapse) {
            sidebarCollapse.addEventListener('click', function() {
                sidebar.classList.toggle('active');
                content.classList.toggle('active');
                
                // Save sidebar state to localStorage
                const isCollapsed = sidebar.classList.contains('active');
                localStorage.setItem('sidebarCollapsed', isCollapsed);
            });
        }

        // Restore sidebar state from localStorage
        const isCollapsed = localStorage.getItem('sidebarCollapsed') === 'true';
        if (isCollapsed) {
            sidebar.classList.add('active');
            content.classList.add('active');
        }
    }

    // Initialize Bootstrap tooltips
    function initializeTooltips() {
        const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
        tooltipTriggerList.map(function (tooltipTriggerEl) {
            return new bootstrap.Tooltip(tooltipTriggerEl);
        });
    }

    // Set active navigation item
    function initializeActiveNavigation() {
        const currentPath = window.location.pathname;
        const sidebarLinks = document.querySelectorAll('.sidebar-link');
        
        sidebarLinks.forEach(link => {
            link.classList.remove('active');
            
            const href = link.getAttribute('href');
            if (href && currentPath.startsWith(href)) {
                link.classList.add('active');
            }
        });
    }

    // Auto-refresh functionality for certain pages
    function initializeAutoRefresh() {
        const autoRefreshElements = document.querySelectorAll('[data-auto-refresh]');
        
        autoRefreshElements.forEach(element => {
            const interval = parseInt(element.getAttribute('data-auto-refresh')) * 1000;
            if (interval > 0) {
                setInterval(() => {
                    if (document.visibilityState === 'visible') {
                        window.location.reload();
                    }
                }, interval);
            }
        });
    }

    // Check authentication status
    function checkAuthenticationStatus() {
        const token = localStorage.getItem('token');
        if (!token && window.location.pathname !== '/admin/login') {
            window.location.href = '/admin/login';
            return;
        }

        // Verify token validity periodically
        if (token) {
            setInterval(verifyToken, 5 * 60 * 1000); // Check every 5 minutes
        }
    }

    // Verify token validity
    function verifyToken() {
        const token = localStorage.getItem('token');
        if (!token) return;

        fetch('/api/auth/profile', {
            headers: {
                'Authorization': 'Bearer ' + token
            }
        })
        .catch(() => {
            // Token is invalid, redirect to login
            localStorage.removeItem('token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/admin/login';
        });
    }

    // Global logout function
    window.logout = function() {
        const token = localStorage.getItem('token');
        
        if (token) {
            fetch('/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Authorization': 'Bearer ' + token
                }
            })
            .finally(() => {
                localStorage.removeItem('token');
                localStorage.removeItem('refresh_token');
                window.location.href = '/admin/login';
            });
        } else {
            localStorage.removeItem('token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/admin/login';
        }
    };

    // Global utility functions
    window.AdminUtils = {
        // Show loading state on buttons
        showButtonLoading: function(button, loadingText = 'Loading...') {
            if (!button.dataset.originalText) {
                button.dataset.originalText = button.innerHTML;
            }
            button.innerHTML = `<span class="spinner-border spinner-border-sm me-2"></span>${loadingText}`;
            button.disabled = true;
        },

        // Hide loading state on buttons
        hideButtonLoading: function(button) {
            if (button.dataset.originalText) {
                button.innerHTML = button.dataset.originalText;
                button.disabled = false;
            }
        },

        // Show notification
        showNotification: function(message, type = 'info', duration = 5000) {
            const alertHTML = `
                <div class="alert alert-${type} alert-dismissible fade show position-fixed" 
                     style="top: 20px; right: 20px; z-index: 9999; min-width: 300px;" role="alert">
                    ${message}
                    <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
                </div>
            `;
            
            document.body.insertAdjacentHTML('beforeend', alertHTML);
            
            // Auto-dismiss after duration
            if (duration > 0) {
                setTimeout(() => {
                    const alerts = document.querySelectorAll('.alert.position-fixed');
                    alerts.forEach(alert => {
                        if (alert.textContent.includes(message.replace(/<[^>]*>/g, ''))) {
                            alert.remove();
                        }
                    });
                }, duration);
            }
        },

        // Format numbers with commas
        formatNumber: function(num) {
            return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
        },

        // Format file sizes
        formatFileSize: function(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        },

        // Debounce function for search inputs
        debounce: function(func, wait) {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        },

        // Copy text to clipboard
        copyToClipboard: function(text) {
            navigator.clipboard.writeText(text).then(() => {
                this.showNotification('Copied to clipboard!', 'success', 2000);
            }).catch(() => {
                this.showNotification('Failed to copy to clipboard', 'danger', 3000);
            });
        },

        // Confirm dialog with custom styling
        confirmDialog: function(message, title = 'Confirm Action') {
            return new Promise((resolve) => {
                const modalHTML = `
                    <div class="modal fade" id="confirmModal" tabindex="-1">
                        <div class="modal-dialog">
                            <div class="modal-content">
                                <div class="modal-header">
                                    <h5 class="modal-title">${title}</h5>
                                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                                </div>
                                <div class="modal-body">
                                    ${message}
                                </div>
                                <div class="modal-footer">
                                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                                    <button type="button" class="btn btn-danger" id="confirmBtn">Confirm</button>
                                </div>
                            </div>
                        </div>
                    </div>
                `;
                
                document.body.insertAdjacentHTML('beforeend', modalHTML);
                const modal = new bootstrap.Modal(document.getElementById('confirmModal'));
                
                document.getElementById('confirmBtn').addEventListener('click', () => {
                    modal.hide();
                    resolve(true);
                });
                
                document.getElementById('confirmModal').addEventListener('hidden.bs.modal', () => {
                    document.getElementById('confirmModal').remove();
                    resolve(false);
                });
                
                modal.show();
            });
        }
    };

    // Initialize search functionality with debouncing
    document.addEventListener('input', function(e) {
        if (e.target.matches('[data-search]')) {
            const searchInput = e.target;
            const debouncedSearch = AdminUtils.debounce(() => {
                const searchTerm = searchInput.value.toLowerCase();
                const targetSelector = searchInput.getAttribute('data-search');
                const targets = document.querySelectorAll(targetSelector);
                
                targets.forEach(target => {
                    const text = target.textContent.toLowerCase();
                    const row = target.closest('tr') || target;
                    
                    if (text.includes(searchTerm)) {
                        row.style.display = '';
                    } else {
                        row.style.display = 'none';
                    }
                });
            }, 300);
            
            debouncedSearch();
        }
    });

    // Handle form submissions with loading states
    document.addEventListener('submit', function(e) {
        if (e.target.matches('.ajax-form')) {
            e.preventDefault();
            
            const form = e.target;
            const submitBtn = form.querySelector('button[type="submit"]');
            const formData = new FormData(form);
            
            AdminUtils.showButtonLoading(submitBtn);
            
            fetch(form.action || window.location.pathname, {
                method: form.method || 'POST',
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                },
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    AdminUtils.showNotification('Operation completed successfully!', 'success');
                    if (data.redirect) {
                        window.location.href = data.redirect;
                    }
                } else {
                    AdminUtils.showNotification(data.message || 'Operation failed', 'danger');
                }
            })
            .catch(error => {
                AdminUtils.showNotification('An error occurred: ' + error.message, 'danger');
            })
            .finally(() => {
                AdminUtils.hideButtonLoading(submitBtn);
            });
        }
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K for quick search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            const searchInput = document.querySelector('input[type="search"], input[placeholder*="search" i]');
            if (searchInput) {
                searchInput.focus();
            }
        }
        
        // Escape to close modals
        if (e.key === 'Escape') {
            const openModal = document.querySelector('.modal.show');
            if (openModal) {
                bootstrap.Modal.getInstance(openModal).hide();
            }
        }
    });

    // Auto-save form data in localStorage
    document.addEventListener('input', function(e) {
        if (e.target.matches('[data-auto-save]')) {
            const input = e.target;
            const key = `admin_form_${input.name || input.id}`;
            localStorage.setItem(key, input.value);
        }
    });

    // Restore auto-saved form data
    document.querySelectorAll('[data-auto-save]').forEach(input => {
        const key = `admin_form_${input.name || input.id}`;
        const savedValue = localStorage.getItem(key);
        if (savedValue) {
            input.value = savedValue;
        }
    });

})(); 