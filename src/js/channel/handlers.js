const ChannelHandlers = {
    async getChannelInfo(channelId) {
        try {
            return await ChannelAPI.getInfo(channelId);
        } catch (error) {
            console.error('Error fetching channel info:', error);
            return null;
        }
    },

    async handleCreateChannel(e, channelManager) {
        e.preventDefault();
        const guildIdInput = $('channel-guild-id');
        const guildId = guildIdInput.value || getCurrentGuildId();
    
        await handleFormSubmission({
            formElement: channelManager.form,
            apiFunction: ChannelAPI.create,
            errorContainerId: 'channel-error-container',
            validateForm: () => $('channel-name').value.trim() !== '',
            operationName: 'channel creation',
            onSuccess: async (result) => {
                window.ChannelUI.closeModal(channelManager);
                window.ChannelUI.clearForm();
                window.ChannelUI.hideError();
                await window.ChannelUI.loadChannels(guildId, channelManager);
                
                const channelsData = await GuildAPI.getChannels(guildId);
                if (channelsData.channels && channelsData.channels.length > 0) {
                    const newestChannel = channelsData.channels[channelsData.channels.length - 1];
                    await window.ChannelHandlers.handleChannelSelect(newestChannel, channelManager);
                }
            }
        });
    },

    async handleChannelSelect(channel, channelManager) {
        const guildId = channel.guild_id || channelManager.currentGuildId;
        const targetPath = `/v/${guildId}/${channel.channel_id}`;
        
        if (isCurrentPath(targetPath)) {
            return;
        }
        
        if (guildId !== channelManager.currentGuildId) {
            await window.GuildNavigation.switchToGuild(guildId);
            return;
        }
        
        try {
            const html = await ChannelAPI.getPage(guildId, channel.channel_id);
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, 'text/html');
            const newContent = doc.querySelector('.main-content .container');
            
            if (newContent) {
                document.querySelector('.main-content .container').innerHTML = newContent.innerHTML;
                
                API.utils.processTimestamps(document.querySelector('.main-content .container'));
                if (window.MessageAPI) {
                    window.MessageAPI.init();
                    if (channel.channel_id) {
                        window.MessageManager.loadChannelMessages(channel.channel_id);
                    }
                }
                
                const messageInput = document.getElementById('message-input');
                if (messageInput && window.getResponsiveChannelPlaceholder) {
                    messageInput.placeholder = window.getResponsiveChannelPlaceholder(channel.name);
                }
            }
            
            history.pushState({channelId: channel.channel_id, guildId: guildId}, '', targetPath);
            document.title = doc.title;
            
            channelManager.focusedChannel = channel.channel_id;
            window.setActiveChannel(channel.channel_id);
            if (window.innerWidth > 768) {
                document.getElementById('members-sidebar').classList.add('visible');
                document.querySelector('.main-content').classList.add('with-members');
            }
            
        } catch (error) {
            NavigationUtils.redirectToChannel(guildId, channel.channel_id);
        }
    }
};

window.ChannelHandlers = ChannelHandlers;