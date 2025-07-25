const ChannelUI = {
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
            
            channelManager.updateChannelsList(data.channels);
            channelManager.checkGuildOwnershipAndShowGears(channelManager.currentGuildId);
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
        channelManager.getChannelInfo(channelId).then(data => {
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
};


window.ChannelUI = ChannelUI;