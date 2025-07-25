const ChannelUI = {
    async checkGuildOwnershipAndShowGears(guildId) {
        await window.PermissionManager.updateGuildUI(guildId);
        return await window.PermissionManager.hasPermission(guildId, 'canManageChannels');
    },

    clearForm() {
        $('channel-name').value = '';
        $('channel-description').value = '';
    },

    closeModal(channelManager) {
        if (channelManager.modal) {
            window.modalManager.closeModal(channelManager.modal.id);
        }
    },

    hideError() {
        const errorContainer = $('channel-error-container');
        if (errorContainer) {
            errorContainer.style.display = 'none';
        }
    },
    async loadChannels(guildId, channelManager) {
        channelManager.currentGuildId = guildId;
        try {
            const data = await GuildAPI.getChannels(guildId);
            
            if (data.error) {
                console.error('Error loading channels:', data.error);
                return;
            }
            
            window.ChannelUI.updateChannelsList(data.channels, channelManager);
            window.ChannelUI.checkGuildOwnershipAndShowGears(channelManager.currentGuildId);
            window.ChannelUI.showChannelsSidebar();
            
            if (channelManager.focusedChannel) {
                setTimeout(() => {
                    window.setActiveChannel(channelManager.focusedChannel);
                }, 50);
            }
        } catch (error) {
            console.error('Error loading guild channels:', error);
        }
    }, 
createChannelElement(channel, channelManager) {
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
            window.ChannelHandlers.handleChannelSelect(channel, channelManager);
        });
        
        const gearBtn = channelElement.querySelector('.channel-settings-btn');
        if (gearBtn) {
            gearBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                window.ChannelUI.showChannelInfo(channel.channel_id, channel.name, channelManager);
            });
        }
        
        return channelElement;
    },
         showChannelsSidebar() {
        if (window.innerWidth > 768) {
            const channelsSidebar = document.querySelector('.channels-sidebar');
            if (channelsSidebar) {
                channelsSidebar.style.display = 'block';
            }
        }
    },
showChannelInfo(channelId, channelName, channelManager) {
        window.ChannelHandlers.getChannelInfo(channelId).then(data => {
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
    },
updateChannelsList(channels, channelManager) {
        this.updateDesktopChannels(channels, channelManager);
        this.updateMobileChannels(channels, channelManager);
    },

    updateDesktopChannels(channels, channelManager) {
        if (!channelManager.channelsList) return;
        
        channelManager.channelsList.innerHTML = '';
        
        if (channels && channels.length > 0) {
            channels.forEach(channel => {
                const channelElement = window.ChannelUI.createChannelElement(channel, channelManager);
                channelManager.channelsList.appendChild(channelElement);
            });
        } else {
            channelManager.channelsList.innerHTML = '<div class="no-channels">No channels yet</div>';
        }
    },

    updateMobileChannels(channels, channelManager) {
        const guildElement = document.querySelector(`[data-guild-id="${channelManager.currentGuildId}"]`);
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
                channelElement.addEventListener('click', () => window.ChannelHandlers.handleChannelSelect({...channel, guild_id: channelManager.currentGuildId}, channelManager));
                channelsContainer.appendChild(channelElement);
            });
        }
    },

    updateMobileChannelsForGuild(guildId, channels, channelManager) {
        const guildElement = document.querySelector(`[data-guild-id="${guildId}"]`);
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
                channelElement.addEventListener('click', () => window.ChannelHandlers.handleChannelSelect({...channel, guild_id: guildId}, channelManager));
                channelsContainer.appendChild(channelElement);
            });
        }
    }        
};


window.ChannelUI = ChannelUI;