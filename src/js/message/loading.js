const MessageLoading = {
    async loadMoreMessages() {
        if (!MessageManager.currentChannelId || !MessageManager.hasMoreMessages || MessageManager.isLoadingMessages) return;

        MessageManager.isLoadingMessages = true;
        const oldestMessageId = MessageManager.messages[0]?.message_id;

        try {
            const response = await MessageManager.getChannelMessages(MessageManager.currentChannelId, oldestMessageId);
            if (response.success && response.messages.length > 0) {
                const newMessages = response.messages.reverse();

                const messagesList = $('messages-list');
                const firstVisible = messagesList.firstElementChild;
                const previousScrollHeight = messagesList.scrollHeight;

                MessageManager.messages = [...newMessages, ...MessageManager.messages];

                newMessages.forEach(msg => {
                    const el = MessageUI.createMessageElement(msg);
                    messagesList.insertBefore(el, firstVisible);
                });

                const newScrollHeight = messagesList.scrollHeight;
                messagesList.scrollTop += newScrollHeight - previousScrollHeight;

                MessageManager.hasMoreMessages = response.has_more;
            }
        } catch (error) {
            console.error('Failed to load more messages:', error);
        } finally {
            MessageManager.isLoadingMessages = false;
        }
    }
};

window.MessageLoading = MessageLoading;