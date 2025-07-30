class TypingIndicator {
    constructor() {
            this.typingUsers = [];
            this.typingTimeout = null;
            this.currentChannelId = null;
            this.isTyping = false;
            this.lastInputLength = 0;
        }

setupTypingIndicator() {
        this.typingUsers = [];
        this.isTyping = false;
        clearTimeout(this.typingTimeout);
    }

updateTypingUsers(users) {
    this.typingUsers = users || [];
    this.renderTypingIndicator();
}
    renderTypingIndicator() {
        const indicator = document.getElementById('typing-indicator');
        if (!indicator) return;

        if (this.typingUsers.length === 0) {
            indicator.style.display = 'none';
            return;
        }

        const message = this.formatTypingMessage(this.typingUsers);
        indicator.innerHTML = `
            <div class="typing-dots">
                <span></span>
                <span></span>
                <span></span>
            </div>
            <span class="typing-text"></span>
        `;
        indicator.querySelector('.typing-text').textContent = message;
        indicator.style.display = 'flex';
    }

    formatTypingMessage(users) {
        if (users.length === 0) return '';
        if (users.length === 1) return `${users[0]} is typing...`;
        if (users.length === 2) return `${users[0]} and ${users[1]} are typing...`;
        if (users.length === 3) return `${users[0]}, ${users[1]}, and ${users[2]} are typing...`;
        return 'Several people are typing...';
    }

    startTyping(channelId) {
        if (this.currentChannelId !== channelId) {
            if (this.isTyping && this.currentChannelId) {
                this.stopTyping(this.currentChannelId, true);
            }
            this.currentChannelId = channelId;
            this.isTyping = false;
        }

        if (!this.isTyping) {
            this.isTyping = true;
            if (window.sendMessage) {
                window.sendMessage({
                    type: 'typing_start',
                    channel_id: channelId
                });
            }
        }

        clearTimeout(this.typingTimeout);
        this.typingTimeout = setTimeout(() => {
            if (this.currentChannelId === channelId) {
                this.stopTyping(channelId, true);
            }
        }, 2500);
    }

    stopTyping(channelId, force = false) {
        clearTimeout(this.typingTimeout);
        
        const shouldSendStop = force || (this.isTyping && channelId === this.currentChannelId);
        if (shouldSendStop && window.sendMessage) {
            window.sendMessage({
                type: 'typing_stop',
                channel_id: channelId
            });
        }
        
        if (channelId === this.currentChannelId || force) {
            this.isTyping = false;
        }
    }
    
    forceStopTyping(channelId) {
        this.stopTyping(channelId, true);
    }

    handleInputChange(inputValue, channelId) {
        const currentLength = inputValue.length;
        const isBackspace = currentLength < this.lastInputLength;
        this.lastInputLength = currentLength;
        
        if (currentLength === 0) {
            this.stopTyping(channelId);
            return;
        }
        
        if (!isBackspace && currentLength > 0) {
            this.startTyping(channelId);
        }
    }

    filterTypingUsers(users, currentUserId) {
        if (!users || !Array.isArray(users)) return [];
        return users.filter(user => user !== currentUserId);
    }
}

window.TypingIndicator = new TypingIndicator();