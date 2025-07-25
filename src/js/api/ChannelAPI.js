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
                this.hideError();
                await this.loadChannels(guildId);
                
                const channelsData = await GuildAPI.getChannels(guildId);
                if (channelsData.channels && channelsData.channels.length > 0) {
                    const newestChannel = channelsData.channels[channelsData.channels.length - 1];
                    this.handleChannelSelect(newestChannel);
                }
            }
        });
    }

async loadChannels(guildId) {
        this.currentGuildId = guildId;
        try {
            const data = await GuildAPI.getChannels(guildId);
            
            if (data.error) {
                console.error('Error loading channels:', data.error);
                return;
            }
            
            this.updateChannelsList(data.channels);
            this.checkGuildOwnershipAndShowGears(this.currentGuildId);
            this.showChannelsSidebar();
            
            if (this.focusedChannel) {
                setTimeout(() => {
                    window.setActiveChannel(this.focusedChannel);
                }, 50);
            }
        } catch (error) {
            console.error('Error loading guild channels:', error);
        }
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
                const channelElement = this.createChannelElement(channel);
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
                channelElement.addEventListener('click', () => this.handleChannelSelect({...channel, guild_id: this.currentGuildId}));
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
                    channelElement.addEventListener('click', () => this.handleChannelSelect({...channel, guild_id: guildId}));
                    channelsContainer.appendChild(channelElement);
                });
            }
        }
    createChannelElement(channel) {
        const channelElement = document.createElement('div');
        channelElement.className = 'channel-item';
        channelElement.setAttribute('data-channel-id', channel.channel_id);
        channelElement.innerHTML = `
            <span class="channel-name">#${window.truncateChannelName ? window.truncateChannelName(channel.name) : channel.name}</span>
            <button class="settings-btn channel-settings-btn" title="Channel Settings" data-channel-id="${channel.channel_id}" data-channel-name="${channel.name}" style="display: none;">⚙️</button>
        `;
                
        channelElement.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            this.handleChannelSelect(channel);
        });
        
const gearBtn = channelElement.querySelector('.channel-settings-btn');
if (gearBtn) {
    gearBtn.addEventListener('click', (e) => {
        e.preventDefault();
        e.stopPropagation();
        this.showChannelInfo(channel.channel_id, channel.name);
    });
}
        
        return channelElement;
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

    hideError() {
        const errorContainer = $('channel-error-container');
        if (errorContainer) {
            errorContainer.style.display = 'none';
        }
    }
showChannelsSidebar() {
        if (window.innerWidth > 768) {
            const channelsSidebar = document.querySelector('.channels-sidebar');
            if (channelsSidebar) {
                channelsSidebar.style.display = 'block';
            }
        }
    }

hideChannelsSidebar() {
        const channelsSidebar = document.querySelector('.channels-sidebar');
        if (channelsSidebar) {
            channelsSidebar.style.display = 'none';
        }
    }

async getChannelInfo(channelId) {
    try {
        const response = await fetch(`/api/channels/${channelId}/info`);
        return await response.json();
    } catch (error) {
        console.error('Error fetching channel info:', error);
        return null;
    }
}
async checkGuildOwnershipAndShowGears(guildId) {
    await window.PermissionManager.updateGuildUI(guildId);
    return await window.PermissionManager.hasPermission(guildId, 'canManageChannels');
}

showChannelInfo(channelId, channelName) {
    this.getChannelInfo(channelId).then(data => {
        if (data) {
            if (window.channelMenuAPI) {
                window.channelMenuAPI.renderButtons('channel-info-modal', channelId, channelName);
                
                setTimeout(() => {
                    const nameElement = document.getElementById('channel-info-name');
                    const descElement = document.getElementById('channel-info-description');
                    const idElement = document.getElementById('channel-info-id');
                    const descTextarea = document.getElementById('channel-desc-edit');
                    
                    if (nameElement) nameElement.textContent = `#${channelName}`;
                    if (descElement) descElement.textContent = data.description || 'No description provided';
                    if (idElement) idElement.textContent = channelId;
                    if (descTextarea) descTextarea.value = data.description || '';
                }, 200);
            }
            
            window.modalManager.openModal('channel-info-modal');
        }
    });
}

async handleChannelSelect(channel) {
    const guildId = channel.guild_id || this.currentGuildId;
    const targetPath = `/v/${guildId}/${channel.channel_id}`;
    
    if (isCurrentPath(targetPath)) {
        return;
    }
    
if (guildId !== this.currentGuildId) {
        await window.GuildNavigation.switchToGuild(guildId);
        return;
    }
    
    try {
        const html = await API.channel.getPage(guildId, channel.channel_id);
        const parser = new DOMParser();
        const doc = parser.parseFromString(html, 'text/html');
        const newContent = doc.querySelector('.main-content .container');
        
    if (newContent) {
        document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
        
        API.utils.processTimestamps(document.querySelector('.main-content .container'));
        if (window.MessageAPI) {
            window.MessageAPI.init();
            if (channel.channel_id) {
                window.MessageAPI.loadChannelMessages(channel.channel_id);
            }
        }
        
        const messageInput = document.getElementById('message-input');
        if (messageInput && window.getResponsiveChannelPlaceholder) {
            messageInput.placeholder = window.getResponsiveChannelPlaceholder(channel.name);
        }
        }
        
        history.pushState({channelId: channel.channel_id, guildId: guildId}, '', targetPath);
        document.title = doc.title;
        
        this.focusedChannel = channel.channel_id;
        window.setActiveChannel(channel.channel_id);
        if (window.innerWidth > 768) {
            document.getElementById('members-sidebar').classList.add('visible');
            document.querySelector('.main-content').classList.add('with-members');
        }
        
    } catch (error) {
        NavigationUtils.redirectToChannel(guildId, channelId);
    }
}

}


window.channelManager = new ChannelManager();