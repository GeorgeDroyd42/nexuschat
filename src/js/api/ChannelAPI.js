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
            createBtn.addEventListener('click', (e) => this.handleCreateChannel(e));
        }
    }

    async handleCreateChannel(e) {
        e.preventDefault();
            const guildIdInput = $('channel-guild-id');
            const guildId = guildIdInput.value || getCurrentGuildId();
        
        await handleFormSubmission({
            formElement: this.form,
            apiFunction: API.channel.create,
            errorContainerId: 'channel-error-container',
            validateForm: () => $('channel-name').value.trim() !== '',
            operationName: 'channel creation',
            onSuccess: async (result) => {
                this.closeModal();
                this.clearForm();
                window.ChannelUI.hideError();
                await window.ChannelUI.loadChannels(guildId, this);
                
                const channelsData = await GuildAPI.getChannels(guildId);
                if (channelsData.channels && channelsData.channels.length > 0) {
                    const newestChannel = channelsData.channels[channelsData.channels.length - 1];
                    await window.ChannelHandlers.handleChannelSelect(newestChannel, this);
                }
            }
        });
    }


    updateChannelsList(channels) {
        this.updateDesktopChannels(channels);
        this.updateMobileChannels(channels);
    }

    updateDesktopChannels(channels) {
        if (!this.channelsList) return;
        
        this.channelsList.innerHTML = '';
        
        if (channels && channels.length > 0) {
            channels.forEach(channel => {
                const channelElement = window.ChannelUI.createChannelElement(channel, this);
                this.channelsList.appendChild(channelElement);
            });
        } else {
            this.channelsList.innerHTML = '<div class="no-channels">No channels yet</div>';
        }
    }

    updateMobileChannels(channels) {
        const guildElement = document.querySelector(`[data-guild-id="${this.currentGuildId}"]`);
        if (!guildElement) return;
        
        const channelsContainer = guildElement.querySelector('.guild-channels');
        if (!channelsContainer) return;
        
        channelsContainer.innerHTML = '';
        if (channels && channels.length > 0) {
            channels.forEach(channel => {
                const channelElement = document.createElement('div');
                channelElement.className = 'guild-channel-item';
                channelElement.textContent = window.truncateChannelName ? window.truncateChannelName(channel.name) : channel.name;
                channelElement.setAttribute('data-channel-id', channel.channel_id);
                channelElement.addEventListener('click', () => window.ChannelHandlers.handleChannelSelect({...channel, guild_id: this.currentGuildId}, this));
                channelsContainer.appendChild(channelElement);
            });
        }
    }
    updateMobileChannelsForGuild(guildId, channels, channelsContainer) {
            channelsContainer.innerHTML = '';
            if (channels && channels.length > 0) {
                channels.forEach(channel => {
                    const channelElement = document.createElement('div');
                    channelElement.className = 'guild-channel-item';
                    channelElement.textContent = window.truncateChannelName ? window.truncateChannelName(channel.name) : channel.name;
                    channelElement.setAttribute('data-channel-id', channel.channel_id);
                    channelElement.addEventListener('click', () => window.ChannelHandlers.handleChannelSelect({...channel, guild_id: guildId}, this));
                    channelsContainer.appendChild(channelElement);
                });
            }
        }

closeModal() {
    if (this.modal) {
        window.modalManager.closeModal(this.modal.id);
    }
}

    clearForm() {
        $('channel-name').value = '';
        $('channel-description').value = '';
    }


async getChannelInfo(channelId) {
    try {
        return await API.channel.getInfo(channelId);
    } catch (error) {
        console.error('Error fetching channel info:', error);
        return null;
    }
}
async checkGuildOwnershipAndShowGears(guildId) {
    await window.PermissionManager.updateGuildUI(guildId);
    return await window.PermissionManager.hasPermission(guildId, 'canManageChannels');
}


}


window.channelManager = new ChannelManager();