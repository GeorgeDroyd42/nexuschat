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
                window.modalManager.closeModal('channel-modal');
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
        
        await window.GuildNavigation.navigateToGuild(guildId, channel.channel_id);
    }
};

window.ChannelHandlers = ChannelHandlers;