const MessageAPI = {

    init() {
        MessageUI.setupMessageInput();
        MessageUI.setupScrollListener();
    },
    addNewMessage(message) {
        MessageManager.messages.push(message);
        const messageEl = MessageUI.createMessageElement(message);
        const messagesList = $('messages-list');
        messagesList.appendChild(messageEl);
        messagesList.scrollTop = messagesList.scrollHeight;
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