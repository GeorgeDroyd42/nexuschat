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