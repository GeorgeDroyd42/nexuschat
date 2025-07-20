const MessageAPI = {
    currentChannelId: null,
    messages: [],
    messageCache: new Map(),
    isLoadingMessages: false,
    hasMoreMessages: true,

    async getChannelMessages(channelId, beforeMessageId = null) {
        let url = `/api/channels/${channelId}/messages?limit=25`;
        if (beforeMessageId) {
            url += `&before=${beforeMessageId}`;
        }
        
        return await BaseAPI.get(url);
    },

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

    init() {
        this.setupMessageInput();
        this.setupScrollListener();
    },

    setupMessageInput() {
        const messageInput = $('message-input');
        const sendBtn = $('send-message-btn');
        const charCount = $('char-count');


        if (!messageInput) {
            console.error('message-input element not found!');
            return;
        }

        messageInput.addEventListener('input', (e) => {
            const length = e.target.value.length;
            charCount.textContent = `${length}/2000`;
            
            if (length >= 1950) {
                charCount.style.color = 'var(--error-color)';
            } else if (length >= 1800) {
                charCount.style.color = 'var(--warning-color)';
            } else {
                charCount.style.color = 'var(--text-muted)';
            }
            
            sendBtn.disabled = length === 0 || length > 2000;
            
            e.target.style.height = 'auto';
            e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px';
            
            if (this.currentChannelId && window.TypingIndicator) {
                window.TypingIndicator.handleInputChange(e.target.value, this.currentChannelId);
            }
        });

        messageInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.handleSendMessage();
            }
        });

        sendBtn.addEventListener('click', () => this.handleSendMessage());
    },

    async handleSendMessage() {
        const messageInput = $('message-input');
        const content = messageInput.value.trim();
        
        if (!content || content.length > 2000) {
            return;
        }
        
        if (!this.currentChannelId) {
            this.currentChannelId = getCurrentChannelId();
        }
        
        if (!this.currentChannelId) {
            return;
        }
        
        try {
            await this.sendMessage(this.currentChannelId, content);
            
            if (window.TypingIndicator) {
                window.TypingIndicator.forceStopTyping(this.currentChannelId);
            }
            
            messageInput.value = '';
            messageInput.style.height = 'auto';
            $('char-count').textContent = '0/2000';
            $('char-count').style.color = 'var(--text-muted)';
            $('send-message-btn').disabled = true;
        } catch (error) {
            console.error('Failed to send message:', error);
        }
    },

async loadChannelMessages(channelId) {
    if (this.isLoadingMessages) return;
    
    this.currentChannelId = channelId;
    
    if (this.messageCache.has(channelId)) {
        const cached = this.messageCache.get(channelId);
        this.messages = cached.messages;
        this.hasMoreMessages = cached.hasMore;
        this.renderMessages();
        return;
    }
    
    this.messages = [];
    this.hasMoreMessages = true;
    this.isLoadingMessages = true;
    
    try {
        const response = await this.getChannelMessages(channelId);
        if (response.success) {
            this.messages = (response.messages || []).reverse();
            this.hasMoreMessages = response.has_more;
            
            this.messageCache.set(channelId, {
                messages: [...this.messages],
                hasMore: this.hasMoreMessages
            });
            
            this.renderMessages();
        }
    } catch (error) {
        console.error('Failed to load messages:', error);
    } finally {
        this.isLoadingMessages = false;
    }
},

    renderMessages() {
        const messagesList = $('messages-list');
        if (!messagesList) return;
        
        messagesList.innerHTML = '';
        
        this.messages.forEach(message => {
            const messageEl = this.createMessageElement(message);
            messagesList.appendChild(messageEl);
        });
        
        messagesList.scrollTop = messagesList.scrollHeight;
    },

createMessageElement(message) {
    const messageEl = document.createElement('div');
    messageEl.className = message.is_webhook ? 'message webhook-message' : 'message';
    messageEl.dataset.messageId = message.message_id;
messageEl.dataset.userId = message.user_id;
    
    // Create header
    const headerEl = document.createElement('div');
    headerEl.className = 'message-header';
    
    // Create avatar - use same function for all avatars
    let avatarEl = window.createUserAvatarElement(message.username, message.profile_picture);
    
    // Create message info
    const infoEl = document.createElement('div');
    infoEl.className = 'message-info';
    
    const usernameEl = document.createElement('span');
    usernameEl.className = 'message-username';
    
    if (message.is_webhook) {
        usernameEl.innerHTML = `${message.username} <span class="bot-tag">BOT</span>`;
    } else {
        usernameEl.textContent = message.username;
    }
    
    const timeEl = document.createElement('span');
    timeEl.className = 'message-time';
    timeEl.textContent = formatTimestamp(message.created_at, 'time');
    
    infoEl.appendChild(usernameEl);
    infoEl.appendChild(timeEl);
    
    headerEl.appendChild(avatarEl);
    headerEl.appendChild(infoEl);
    
    // Create content (safe from XSS)
    const contentEl = document.createElement('div');
    contentEl.className = 'message-content';
    contentEl.innerHTML = EmbedUtils.linkifyURLs(message.content);
    
messageEl.appendChild(headerEl);
messageEl.appendChild(contentEl);

const embedEl = EmbedUtils.createEmbedElement(message.content);
if (embedEl) {
    messageEl.appendChild(embedEl);
}

return messageEl;
},

    setupScrollListener() {
        const messagesList = $('messages-list');
        if (!messagesList) return;
        
        messagesList.addEventListener('scroll', () => {
            if (messagesList.scrollTop === 0 && this.hasMoreMessages && !this.isLoadingMessages) {
                this.loadMoreMessages();
            }
        });
    },

async loadMoreMessages() {
    if (!this.currentChannelId || !this.hasMoreMessages || this.isLoadingMessages) return;

    this.isLoadingMessages = true;
    const oldestMessageId = this.messages[0]?.message_id;

    try {
        const response = await this.getChannelMessages(this.currentChannelId, oldestMessageId);
        if (response.success && response.messages.length > 0) {
            const newMessages = response.messages.reverse();

            const messagesList = $('messages-list');
            const firstVisible = messagesList.firstElementChild;
            const previousScrollHeight = messagesList.scrollHeight;

            this.messages = [...newMessages, ...this.messages];

            // Render only new messages
            newMessages.forEach(msg => {
                const el = this.createMessageElement(msg);
                messagesList.insertBefore(el, firstVisible);
            });

            // Adjust scroll position to maintain view
            const newScrollHeight = messagesList.scrollHeight;
            messagesList.scrollTop += newScrollHeight - previousScrollHeight;

            this.hasMoreMessages = response.has_more;
        }
    } catch (error) {
        console.error('Failed to load more messages:', error);
    } finally {
        this.isLoadingMessages = false;
    }
},


    addNewMessage(message) {
        this.messages.push(message);
        const messageEl = this.createMessageElement(message);
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

// Debug initialization

document.addEventListener('DOMContentLoaded', () => {
    MessageAPI.init();
});

// Also try immediate initialization in case DOM is already ready
if (document.readyState === 'loading') {
} else {
    MessageAPI.init();
}

window.MessageAPI = MessageAPI;