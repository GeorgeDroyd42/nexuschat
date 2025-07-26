const UserAPI = {

async getUserProfile(userID) {
    return await BaseAPI.get(`/api/user/${userID}/profile`);
},

async getCurrentUser() {
    return await BaseAPI.get('/api/user/me', { 'Accept': 'application/json' });
},

async getUsersList(page = 1, limit = 50) {
    return await BaseAPI.get(`/api/user/getusers?page=${page}&limit=${limit}`);
},

async makeUserAdmin(username) {
    return await BaseAPI.post(`/api/user/${username}/make-admin`);
},

async demoteUserAdmin(username) {
    return await BaseAPI.post(`/api/user/${username}/demote-admin`);
},
async banUser(userID) {
    return await BaseAPI.post(`/api/ban/${userID}`);
},

async unbanUser(userID) {
    return await BaseAPI.post(`/api/unban/${userID}`);
},

async updateUsername(newUsername) {
    return await BaseAPI.post('/api/user/update-username', {
        username: newUsername
    });
},

async updateBio(newBio) {
    return await BaseAPI.post('/api/user/update-bio', {
        bio: newBio
    });
}
};

