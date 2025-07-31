const MessageHandlers = {
    async handleSendMessage() {
        const messageInput = $('message-input');
        const content = messageInput.value.trim();
        
        if (!content || content.length > 2000) {
            return;
        }
        
        if (!MessageManager.currentChannelId) {
            MessageManager.currentChannelId = getCurrentChannelId();
        }
        
        if (!MessageManager.currentChannelId) {
            return;
        }
        
        try {
            await MessageManager.sendMessage(MessageManager.currentChannelId, content);
            
            if (window.TypingIndicator) {
                window.TypingIndicator.forceStopTyping(MessageManager.currentChannelId);
            }
            
            messageInput.value = '';
            messageInput.style.height = 'auto';
            
            CharCountAPI.update('message-input');
            
            $('send-message-btn').disabled = true;
        } catch (error) {
            console.error('Failed to send message:', error);
        }
    }
};

window.MessageHandlers = MessageHandlers;