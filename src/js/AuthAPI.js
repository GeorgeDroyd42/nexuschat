const AuthAPI = {

async login(formData) {
    return await BaseAPI.request('/api/auth/login', {
        method: 'POST',
        body: formData,
        isFormData: true,
        throwOnError: false
    });
},

async register(formData) {
    return await BaseAPI.request('/api/auth/register', {
        method: 'POST',
        body: formData,
        isFormData: true,
        throwOnError: false
    });
},
    async logout() {
        sessionStorage.setItem('userInitiatedLogout', 'true');
        return await BaseAPI.post('/api/auth/logout');
    },

    async refreshSession() {
        return await BaseAPI.request('/api/auth/refresh', {
            method: 'POST',
            requiresAuth: false,
            headers: { 'credentials': 'include' }
        });
    }
};