class ChannelManager {
    constructor() {
        this.modal = null;
        this.form = null;
        this.channelsList = null;
        this.currentGuildId = null;
        this.focusedChannel = null;
        this.init();
    }

    init() {
        this.modal = $('channel-modal');
        this.form = $('create-channel-form');
        this.channelsList = $('channels-list');
        this.setupEventListeners();
    }

    setupEventListeners() {
        const createBtn = $('create-channel-button');
        if (createBtn) {
            createBtn.addEventListener('click', (e) => window.ChannelHandlers.handleCreateChannel(e, this));
        }
    }
}
window.channelManager = new ChannelManager();

// Self-initialize channel functionality when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const channelId = getCurrentChannelId();
    if (channelId) {
        window.channelManager.focusedChannel = channelId;
        MessageUI.init();
        MessageManager.loadChannelMessages(channelId);
        
        const channelTitle = document.querySelector('.channel-title');
        const messageInput = document.getElementById('message-input');
        if (channelTitle && messageInput && window.getResponsiveChannelPlaceholder) {
            const channelName = channelTitle.textContent.replace('#', '').trim();
            messageInput.placeholder = window.getResponsiveChannelPlaceholder(channelName);
        }
    }
});