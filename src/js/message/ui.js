const MessageUI = {
    createMessageElement(message) {
        const messageEl = document.createElement('div');
        messageEl.className = message.is_webhook ? 'message webhook-message' : 'message';
        messageEl.dataset.messageId = message.message_id;
        messageEl.dataset.userId = message.user_id;
        
        // Create header
        const headerEl = document.createElement('div');
        headerEl.className = 'message-header';
        
        // Create avatar - use same function for all avatars
        let avatarEl = window.AvatarUtils.createSecureAvatar(message.username, message.profile_picture);
        
        // Create message info
        const infoEl = document.createElement('div');
        infoEl.className = 'message-info';
        
        const usernameEl = document.createElement('span');
        usernameEl.className = 'message-username';
        
        if (message.is_webhook) {
            usernameEl.innerHTML = `<span class="bot-tag">BOT</span>`;
            usernameEl.prepend(message.username + ' ');
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
    setupMessageInput() {
        const messageInput = $('message-input');
        const sendBtn = $('send-message-btn');

        if (!messageInput) {
            console.error('message-input element not found!');
            return;
        }

        CharCountAPI.add('message-input', {
            maxLength: 2000,
            warningThreshold: 1800,
            errorThreshold: 1950
        });
                
        messageInput.addEventListener('input', (e) => {
            const length = e.target.value.length;
            sendBtn.disabled = length === 0 || length > 2000;
            
            e.target.style.height = 'auto';
            e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px';
            
            if (MessageManager.currentChannelId && window.TypingIndicator) {
                window.TypingIndicator.handleInputChange(e.target.value, MessageManager.currentChannelId);
            }
        });

        messageInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                MessageHandlers.handleSendMessage();
            }
        });

        sendBtn.addEventListener('click', () => MessageHandlers.handleSendMessage());
    },   
    
    setupScrollListener() {
        const messagesList = $('messages-list');
        if (!messagesList) return;
        
        messagesList.addEventListener('scroll', () => {
            if (messagesList.scrollTop === 0 && MessageManager.hasMoreMessages && !MessageManager.isLoadingMessages) {
                MessageLoading.loadMoreMessages();
            }
        });
    },   
    
    
    renderMessages() {
            const messagesList = $('messages-list');
            if (!messagesList) return;
            
            messagesList.innerHTML = '';
            
            MessageManager.messages.forEach(message => {
                const messageEl = MessageUI.createMessageElement(message);
                messagesList.appendChild(messageEl);
            });
            
            messagesList.scrollTop = messagesList.scrollHeight;
        },
        
    init() {
        this.setupMessageInput();
        this.setupScrollListener();
    }
};

window.MessageUI = MessageUI;