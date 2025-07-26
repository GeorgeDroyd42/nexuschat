class GuildAPI {
    static create(formData) {
        return BaseAPI.post('/api/guild/create', formData, true);
    }

    static fetchUserGuilds() {
        return BaseAPI.get('/api/user/guilds');
    }

    static getChannels(guildId) {
        return BaseAPI.get(`/api/channels/get?guild_id=${guildId}`);
    }

static getMembers(guildId, page = 1) {
    return BaseAPI.get(`/api/guild/${guildId}/members?page=${page}`);
}

    static leave(guildId) {
        return BaseAPI.post(`/api/guild/leave/${guildId}`);
    }

    static join(guildId) {
        return BaseAPI.post(`/api/guild/join/${guildId}`);
    }

    static getPage(guildId) {
        return BaseAPI.get(`/v/${guildId}`);
    }

    static getInfo(guildId) {
        return BaseAPI.get(`/api/guild/${guildId}/info`);
    }

    static async showGuildInfo(guildId) {
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
                            setTimeout(checkElements, 50);
                        }
                    };
                    checkElements();
                });

                window.modalManager.openModal('guild-settings-modal');
            }
        } catch (error) {
            console.error('Error loading guild info:', error);
        }
    }
}

window.GuildAPI = GuildAPI;