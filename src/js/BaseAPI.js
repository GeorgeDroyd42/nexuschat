class BaseAPI {
    static async request(url, options = {}) {
        const requestId = Date.now().toString(36) + Math.random().toString(36).substr(2);
        const method = options.method || 'GET';
        
        console.log(`[${requestId}] ${method} ${url}`);
        
        try {
            const {
                body = null,
                headers = {},
                requiresAuth = true,
                isFormData = false,
                throwOnError = true
            } = options;

            const requestOptions = {
                method,
                credentials: 'include',
                headers: { ...headers }
            };

            if (requiresAuth) {
                const token = await fetchCSRFToken();
                requestOptions.headers['X-CSRF-Token'] = token;
            }

            if (!isFormData && body && method !== 'GET') {
                requestOptions.headers['Content-Type'] = 'application/json';
                requestOptions.body = typeof body === 'string' ? body : JSON.stringify(body);
            } else if (body) {
                requestOptions.body = body;
            }

            const response = await fetch(url, requestOptions);
            console.log(`[${requestId}] Response: ${response.status} ${response.statusText}`);

            const contentType = response.headers.get('content-type');
            const isJson = contentType && contentType.includes('application/json');

            if (!response.ok && throwOnError) {
                const errorMsg = `API request failed: ${response.status} ${response.statusText}`;
                console.error(`[${requestId}] ${errorMsg}`);
                throw new Error(errorMsg);
            }

            const result = isJson ? await response.json() : await response.text();
            console.log(`[${requestId}] Success`);
            return result;
            
        } catch (error) {
            console.error(`[${requestId}] Error:`, error.message);
            throw error;
        }
    }

    static async get(url, headers = {}) {
        return this.request(url, { method: 'GET', headers });
    }

    static async post(url, body = null, isFormData = false) {
        return this.request(url, { method: 'POST', body, isFormData });
    }

    static async put(url, body = null, isFormData = false) {
        return this.request(url, { method: 'PUT', body, isFormData });
    }

    static async delete(url, body = null) {
        return this.request(url, { method: 'DELETE', body });
    }
}

const API = {
    guild: {
        create: (formData) => BaseAPI.post('/api/guild/create', formData, true),
        fetchUserGuilds: () => BaseAPI.get('/api/user/guilds'),
        getChannels: (guildId) => BaseAPI.get(`/api/channels/get?guild_id=${guildId}`),
        getMembers: (guildId) => BaseAPI.get(`/api/guild/${guildId}/members?page=1&limit=200`),
        leave: (guildId) => BaseAPI.post(`/api/guild/leave/${guildId}`),
        join: (guildId) => BaseAPI.post(`/api/guild/join/${guildId}`),
        getPage: (guildId) => BaseAPI.get(`/v/${guildId}`),
        getInfo: (guildId) => BaseAPI.get(`/api/guild/${guildId}/info`),
        showGuildInfo: async (guildId) => {
            try {
                const data = await BaseAPI.get(`/api/guild/${guildId}/info`);
                if (data && window.guildMenuAPI) {
                    window.guildMenuAPI.renderButtons('guild-settings-modal', data);
                    
                    await new Promise(resolve => {
                        const checkElements = () => {
                            const nameElement = document.getElementById('guild-info-name');
                            const descElement = document.getElementById('guild-info-description');
                            const idElement = document.getElementById('guild-info-id');
                            
                            if (nameElement && descElement && idElement) {
                                nameElement.textContent = data.name;
                                descElement.textContent = data.description || 'No description set';
                                idElement.textContent = data.guild_id;
                                resolve();
                            } else {
                                setTimeout(checkElements, 10);
                            }
                        };
                        checkElements();
                    });
                    
                    window.modalManager.openModal('guild-settings-modal');
                }
            } catch (error) {
                console.error('Error showing guild info:', error);
            }
        }    },
    invite: {
        joinByInvite: (inviteCode) => BaseAPI.post(`/api/invite/join/${inviteCode}`),
    },

    channel: {
        create: (formData) => BaseAPI.post('/api/channels/create', formData, true),
        getPage: (guildId, channelId) => BaseAPI.get(`/v/${guildId}/${channelId}`),
        getInfo: (channelId) => BaseAPI.get(`/api/channels/${channelId}/info`),
        delete: (channelId) => BaseAPI.post('/api/channels/delete', {channel_id: channelId}),
        edit: (channelId, name, description) => {
            const formData = new FormData();
            formData.append('name', name);
            formData.append('description', description);
            return BaseAPI.put(`/api/channels/${channelId}/edit`, formData, true);
        }
    },

    auth: {
        login: (formData) => BaseAPI.request('/api/auth/login', { method: 'POST', body: formData, isFormData: true, throwOnError: false }),
        register: (formData) => BaseAPI.request('/api/auth/register', { method: 'POST', body: formData, isFormData: true, throwOnError: false }),
        logout: () => BaseAPI.post('/api/auth/logout'),
        refresh: () => BaseAPI.request('/api/auth/refresh', { method: 'POST', requiresAuth: false })
    },

    user: {
        getProfile: (userId) => BaseAPI.get(`/api/user/${userId}/profile`),
        getCurrentUser: () => BaseAPI.get('/api/user/me'),
        uploadProfilePicture: (formData) => BaseAPI.post('/api/user/profile-picture', formData, true),
        makeAdmin: (username) => BaseAPI.post(`/api/user/${username}/make-admin`),
        demoteAdmin: (username) => BaseAPI.post(`/api/user/${username}/demote-admin`)
    },

    utils: {
            processTimestamps: (container) => {
                container.querySelectorAll('p').forEach(paragraph => {
                    const text = paragraph.textContent;
                    const createdMatch = text.match(/Created:\s*(.+)/);
                    if (createdMatch) {
                        const timestamp = createdMatch[1].trim();
                        if (window.formatTimestamp) {
                            paragraph.innerHTML = paragraph.innerHTML.replace(timestamp, window.formatTimestamp(timestamp, 'date'));
                        }
                    }
                });
            }
        }
    };


window.BaseAPI = BaseAPI;
window.API = API;