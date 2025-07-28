
// Store the original fetch
const originalFetch = window.fetch;

async function fetchCSRFToken() {
    try {
        const response = await originalFetch('/api/csrf-token');
        if (!response.ok) {
            throw new Error('Failed to fetsch CSRF token');
        }
        const data = await response.json();
        return data.csrf_token;
    } catch (error) {
        console.error('Error fetching CSRF token:', error);
        return null;
    }
}

async function fetchWithCSRF(url, options = {}) {
    const token = await fetchCSRFToken();
    
    if (!options.headers) {
        options.headers = {};
    }
    
    if (token) {
        options.headers['X-CSRF-Token'] = token;
    }
    
    return originalFetch(url, options);
}


window.fetch = async function(url, options = {}) {
    if (url.includes('/api/csrf-token')) {
        return originalFetch(url, options);
    }
    
    return fetchWithCSRF(url, options);
};

document.addEventListener('DOMContentLoaded', () => {
});
