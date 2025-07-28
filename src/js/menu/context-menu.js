class ContextMenu {
constructor() {
    this.menu = null;
this.currentChannelId = null;
this.currentChannelName = null;
this.currentGuildId = null;
this.currentMessageId = null;
this.currentMessageContent = null;
    this.isOpen = false;
    this.createMenu();
    this.bindEvents();
}

    createMenu() {
        this.menu = document.createElement('div');
        this.menu.className = 'context-menu';
        document.body.appendChild(this.menu);
    }

    

async show(x, y, contextType = 'guild') {
    try {
        let url = `/api/context/${contextType}?`;
        if (this.currentGuildId) url += `guild_id=${this.currentGuildId}&`;
        if (this.currentChannelId) url += `channel_id=${this.currentChannelId}&`;
        if (this.currentMessageId) url += `message_id=${this.currentMessageId}&`;
        
        const response = await fetch(url);
        const data = await response.json();
        
        if (data.buttons && data.buttons.length > 0) {
            this.isOpen = true;
            document.querySelectorAll('.guild-tooltip').forEach(tooltip => tooltip.remove());
            this.menu.innerHTML = '';
            this.renderButtons(data.buttons);
            this.menu.style.left = x + 'px';
            this.menu.style.top = y + 'px';
            this.menu.style.display = 'block';
        }
    } catch (error) {
        console.error('Error fetching context menu:', error);
    }
}


    hide() {
        this.menu.style.display = 'none';
        this.isOpen = false;
    }

    renderButtons(buttons) {
        buttons.forEach(button => {
            if (button.type === 'separator') {
                const separatorElement = document.createElement('div');
                separatorElement.className = 'context-menu-separator';
                this.menu.appendChild(separatorElement);
            } else {
                const btnElement = document.createElement('div');
                btnElement.className = 'context-menu-item';
                btnElement.textContent = button.text;
                btnElement.style.color = button.color;
                btnElement.addEventListener('click', () => {
                    this.handleAction(button.action);
                    this.hide();
                });
                this.menu.appendChild(btnElement);
            }
        });
    }

    handleAction(action) {
        switch(action) {
case 'invite':
    this.generateInviteCode();
    break;
            case 'copy_guild_id':
                navigator.clipboard.writeText(this.currentGuildId);
                break;
            case 'copy_channel_id':
                navigator.clipboard.writeText(this.currentChannelId);
                break;
            case 'copy_message_id':
                navigator.clipboard.writeText(this.currentMessageId);
                break;
            case 'copy_message_content':
                if (this.currentMessageContent) {
                    navigator.clipboard.writeText(this.currentMessageContent);
                }
                break;
            case 'delete_message':
                if (window.MessageManager) window.MessageManager.deleteMessage(this.currentMessageId);
                break;
            case 'delete_channel':
                window.ChannelAPI.delete(this.currentChannelId);
                break;
            case 'channel_settings':
                window.ChannelUI.showChannelInfo(this.currentChannelId, this.currentChannelName, window.channelManager);
                break;
            case 'leave_guild':
                window.GuildAPI.leave(this.currentGuildId);
                break;
            case 'guild_settings':
                if (window.GuildAPI && window.GuildAPI.showGuildInfo) {
                    window.GuildAPI.showGuildInfo(this.currentGuildId);
                }
                break;
                }
            }

    async generateInviteCode() {
        try {
            const response = await fetch(`/api/invite/generate/${this.currentGuildId}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                }
            });
            
            const data = await response.json();
            
            if (data.success) {
                window.modalManager.openModal('confirm-modal');
                const linkElement = document.getElementById('invite-link-text');
                if (linkElement) linkElement.value = data.invite_url;
            } else {
                console.error('Failed to generate invite code:', data.error);
            }
        } catch (error) {
            console.error('Error generating invite code:', error);
        }
    }

    bindEvents() {
        document.addEventListener('click', () => this.hide());
document.addEventListener('contextmenu', (e) => {
        e.preventDefault();
        
        const sidebarElement = e.target.closest('.sidebar');
        const channelsSidebarElement = e.target.closest('.channels-sidebar');
        const messageElement = e.target.closest('.message');
        
        if (sidebarElement || channelsSidebarElement || messageElement) {
        
        let contextType = 'sidebar';
        
if (messageElement) {
    contextType = 'message';
    this.currentMessageId = messageElement.dataset.messageId;
    this.currentGuildId = getCurrentGuildId();
    
    const messageContentEl = messageElement.querySelector('.message-content');
    this.currentMessageContent = messageContentEl ? messageContentEl.textContent.trim() : '';

} else if (channelsSidebarElement) {
    const channelElement = e.target.closest('[data-channel-id]');
    if (channelElement) {
        contextType = 'channel';
        this.currentChannelId = channelElement.dataset.channelId;
const channelNameSpan = channelElement.querySelector('.channel-name');
this.currentChannelName = channelNameSpan ? channelNameSpan.textContent.replace('#', '') : channelElement.textContent.trim();
        this.currentGuildId = getCurrentGuildId();
    } else {
        contextType = 'channels-sidebar';
        this.currentGuildId = getCurrentGuildId();
    }
        } else if (sidebarElement) {
            const guildChannelsElement = e.target.closest('.guild-channels');
            const channelElement = e.target.closest('[data-channel-id]');
            const guildElement = e.target.closest('[data-guild-id]');
            
if (guildChannelsElement && channelElement) {
    contextType = 'channel';
    this.currentChannelId = channelElement.dataset.channelId;
const channelNameSpan = channelElement.querySelector('.channel-name');
this.currentChannelName = channelNameSpan ? channelNameSpan.textContent.replace('#', '') : channelElement.textContent.trim();
    this.currentGuildId = guildElement ? guildElement.dataset.guildId : getCurrentGuildId();

} else if (guildChannelsElement) {
    contextType = 'channels-sidebar';
    this.currentGuildId = guildElement ? guildElement.dataset.guildId : getCurrentGuildId();
} else if (guildElement) {
    contextType = 'guild';
    this.currentGuildId = guildElement.dataset.guildId;
} else {
    contextType = 'guild';
    this.currentGuildId = getCurrentGuildId();
}
        }
        
        this.show(e.clientX, e.clientY, contextType);
    }
});
    }
}

window.ctxMenu = new ContextMenu();

