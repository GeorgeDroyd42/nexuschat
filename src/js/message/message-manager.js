const MessageManager = {
    currentChannelId: null,
    messages: [],
    messageCache: new Map(),
    isLoadingMessages: false,
    hasMoreMessages: true,

    

async sendMessage(channelId, content) {
    
    return new Promise((resolve, reject) => {
        if (window.sendMessage) {
            const success = window.sendMessage({
                type: 'message',
                channel_id: channelId,
                content: content
            });
            
            if (success) {
                resolve({ success: true });
            } else {
                console.error('Failed to send message - WebSocket not ready'); 
                reject(new Error('WebSocket not connected'));
            }
        } else {
            console.error('sendMessage function not available');
            reject(new Error('WebSocket function not available'));
        }
    });
},

async loadChannelMessages(channelId) {
    if (this.isLoadingMessages) return;
    
    this.currentChannelId = channelId;
    
    if (window.sendMessage) {
        window.sendMessage({
            type: 'request_typing_state',
            channel_id: channelId
        });
    }
    
    if (this.messageCache.has(channelId)) {
        const cached = this.messageCache.get(channelId);
        this.messages = cached.messages;
        this.hasMoreMessages = cached.hasMore;
        MessageUI.renderMessages();
        return;
    }
    
    this.messages = [];
    this.hasMoreMessages = true;
    this.isLoadingMessages = true;
    
    try {
        const response = await MessageAPI.getChannelMessages(channelId);
        if (response.success) {
            this.messages = (response.messages || []).reverse();
            this.hasMoreMessages = response.has_more;
            
            this.messageCache.set(channelId, {
                messages: [...this.messages],
                hasMore: this.hasMoreMessages
            });
            
            MessageUI.renderMessages();
        }
    } catch (error) {
        console.error('Failed to load messages:', error);
    } finally {
        this.isLoadingMessages = false;
    }
},


    addNewMessage(message) {
        this.messages.push(message);
        const messageEl = MessageUI.createMessageElement(message);
        const messagesList = $('messages-list');
        messagesList.appendChild(messageEl);
        messagesList.scrollTop = messagesList.scrollHeight;
    },

async deleteMessage(messageId) {
    if (window.sendMessage) {
        return window.sendMessage({
            type: 'delete_message',
            message_id: messageId
        });
    }
    return false;
}

};

window.MessageManager = MessageManager;