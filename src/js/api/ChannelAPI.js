class ChannelAPI {
    static create(formData) {
        return BaseAPI.post('/api/channels/create', formData, true);
    }

    static getPage(guildId, channelId) {
        return BaseAPI.get(`/v/${guildId}/${channelId}`);
    }

    static getInfo(channelId) {
        return BaseAPI.get(`/api/channels/${channelId}/info`);
    }

    static delete(channelId) {
        return BaseAPI.post('/api/channels/delete', {channel_id: channelId});
    }

    static edit(channelId, name, description) {
        const formData = new FormData();
        formData.append('name', name);
        formData.append('description', description);
        return BaseAPI.put(`/api/channels/${channelId}/edit`, formData, true);
    }
}

window.ChannelAPI = ChannelAPI;