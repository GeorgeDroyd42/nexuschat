class MessageAPI {
    static getChannelMessages(channelId, beforeMessageId = null) {
        let url = `/api/channels/${channelId}/messages?limit=25`;
        if (beforeMessageId) {
            url += `&before=${beforeMessageId}`;
        }
        
        return BaseAPI.get(url);
    }
}

window.MessageAPI = MessageAPI;