let websocket = null;
let processedMessageIds = new Set();

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/ws`;
    
    websocket = new WebSocket(wsUrl);
    
websocket.onopen = function() {
    console.log('WebSocket connected');
    
    window.wsQueue.setWebSocket(websocket);
    
    // Send a status update message to trigger server-side online broadcasting
    window.wsQueue.sendMessage({
        type: 'status_update',
        status: 'online'
    });
    
    window.wsQueue.flushQueue();
    
    if (window.MessageManager && window.MessageManager.currentChannelId) {
        window.MessageManager.loadChannelMessages(window.MessageManager.currentChannelId);
    }
};

websocket.onclose = function(event) {
        console.log('WebSocket disconnected', event.code, event.reason);
        
        window.MessageManager.messageCache.clear();
        
        if (event.reason && event.reason !== 'unauthed') {
            displayErrorMessage(event.reason);
            setTimeout(() => {
                NavigationUtils.redirectToLogin();
            }, 2000);
            return;
        }
        
        setTimeout(connectWebSocket, 3000);
    };
    
    websocket.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            handleMessage(data);
        } catch (error) {
            console.error('Error parsing websocket message:', error);
        }
    };
}


function handleMessage(data) {
    switch(data.type) {
        case 'user_banned':
            redirectToLogin(data.message);
            break;
            
        case 'guild_created':
            handleGuildCreated(data);
            break;
        case 'guild_removed':
            handleGuildRemoved(data);
            break;
            
        case 'member_joined':
            handleMemberJoined(data);
            break;            
        case 'member_left':
                    handleMemberLeft(data);
                    break;
        case 'channel_created':
            handleChannelCreated(data);
            break;                    
        case 'channel_deleted':
            handleChannelDeleted(data);
            break;
        case 'channel_updated':
            handleChannelUpdated(data);
            break;
        case 'session_terminated':
            if (!sessionStorage.getItem('userInitiatedLogout')) {
                redirectToLogin(data.message);
            } else {
                sessionStorage.removeItem('userInitiatedLogout');
                NavigationUtils.redirectToLogin();
            }
            break;
        case 'all_sessions_terminated':
            redirectToLogin('All your sessions were terminated by an administrator');
            break;
        case 'new_message':
                    handleNewMessage(data);
                    break;
                case 'message_deleted':
                    handleMessageDeleted(data);
                    break;
                case 'username_changed':
                    handleUsernameChanged(data);
                    break;
                case 'user_status_changed':
                    window.StatusSockets.handleUserStatusChanged(data);
                    break;
                case 'typing_update':
                    handleTypingUpdate(data);
                    break;
                case 'webhook_created':
                    handleWebhookCreated(data);
                    break;
                case 'webhook_deleted':
                    handleWebhookDeleted(data);
                    break;                    
            }
}

function redirectToLogin(message) {
    NavigationUtils.redirectToLogin(message);
}

function handleGuildCreated(data) {
    if (data.guild && window.GuildUI.refreshGuildList) {
        window.GuildUI.refreshGuildList();
    }
}

function initWebSocket() {
    connectWebSocket();
}
function handleMemberJoined(data) {
    if (data.user_id && data.guild_id) {
        const member = {
            user_id: data.user_id,
            username: data.username,
            profile_picture: data.profile_picture
        };
        addMemberToList(member, data.guild_id);
    }
}

function handleMemberLeft(data) {
    if (data.user_id && data.guild_id) {
        removeMemberFromList(data.user_id, data.guild_id);
    }
}

function handleGuildRemoved(data) {
    if (data.guild_id) {
        const guildList = $('guild-list');
        if (guildList) {
            const guildElement = guildList.querySelector(`[data-guild-id="${data.guild_id}"]`);
            if (guildElement) {
                if (guildElement.currentTooltip) {
                    guildElement.currentTooltip.remove();
                    guildElement.currentTooltip = null;
                }
                guildElement.remove();
            }
        }
        
        if (isCurrentGuild(data.guild_id)) {
            NavigationUtils.redirectToMain();
        }
    }
}

function handleChannelCreated(data) {
    if (data.channel_id && data.guild_id && window.channelManager) {
        window.ChannelUI.loadChannels(data.guild_id, window.channelManager);
    }
}

async function handleChannelDeleted(data) {
    if (data.channel_id && data.guild_id && window.channelManager) {
        if (isCurrentChannel(data.guild_id, data.channel_id)) {
            const channelsData = await GuildAPI.getChannels(data.guild_id);
            if (channelsData.channels && channelsData.channels.length > 0) {
                await window.ChannelHandlers.handleChannelSelect(channelsData.channels[0], window.channelManager);
            } else {
                await window.GuildNavigation.navigateToGuild(data.guild_id);
            }
        }
        window.ChannelUI.loadChannels(data.guild_id, window.channelManager);
    }
}

function handleChannelUpdated(data) {
    if (data.channel_id && data.guild_id && window.channelManager) {
        if (isCurrentChannel(data.guild_id, data.channel_id)) {
            const channelTitle = document.querySelector('.channel-title');
            if (channelTitle) {
                const nameElement = channelTitle.querySelector('h2');
                const descElement = channelTitle.querySelector('.channel-description');
                
                if (nameElement) {
                    nameElement.textContent = `# ${data.name}`;
                }
                
                if (data.description) {
                    if (descElement) {
                        descElement.textContent = data.description;
                        descElement.style.display = 'block';
                    } else {
                        const newDesc = document.createElement('span');
                        newDesc.className = 'channel-description';
                        newDesc.textContent = data.description;
                        channelTitle.appendChild(newDesc);
                    }
                } else if (descElement) {
                    descElement.style.display = 'none';
                }
            }
        }
        window.ChannelUI.loadChannels(data.guild_id, window.channelManager);
    }
}

function handleNewMessage(data) {
    if (window.MessageManager) {
        // Only add message if it belongs to the currently viewed channel
        if (data.channel_id && data.channel_id === window.MessageManager.currentChannelId) {
            MessageManager.addNewMessage(data);
        }
    }
}

function handleMessageDeleted(data) {
    if (window.MessageManager) {
        // Only remove message if it belongs to the currently viewed channel
        if (data.channel_id && data.channel_id === window.MessageManager.currentChannelId) {
            const messageElement = document.querySelector(`[data-message-id="${data.message_id}"]`);
            if (messageElement) {
                messageElement.remove();
                console.log('Message deleted from UI:', data.message_id);
            }
        }
    }
}

function handleUsernameChanged(data) {
    if (data.user_id && data.old_username && data.new_username) {
        if (window.profileManager && window.profileManager.currentUser && window.profileManager.currentUser.user_id === data.user_id) {
            window.profileManager.loadUserData().then(userData => {
                if (window.profileMenuAPI && userData) {
                    window.profileMenuAPI.currentUser = userData;
                    window.profileMenuAPI.populateTabContent();
                }
                
                const userAvatar = document.querySelector('#user-avatar');
                if (userAvatar && window.AvatarUtils) {
                    const newAvatar = window.AvatarUtils.createSecureAvatar(userData.username, userData.profile_picture);
                    const avatarContent = newAvatar.firstChild;
                    userAvatar.replaceWith(avatarContent);
                }
            });
        }   
        
        if (typeof loadUserGuilds === 'function') {
            loadUserGuilds();
        }
        
        window.StatusUI.refreshMainMemberList();
        
        window.StatusUI.refreshGuildSettingsMembers();
        
        if (window.MessageManager && window.MessageManager.currentChannelId && typeof window.MessageManager.loadChannelMessages === 'function') {
            window.MessageManager.loadChannelMessages(window.MessageManager.currentChannelId);
        }
        
        console.log(`Username updated: ${data.old_username} → ${data.new_username}`);
    }
}
function handleTypingUpdate(data) {
    if (window.TypingIndicator) {
        if (data.channel_id === window.MessageManager?.currentChannelId) {
            const users = data.typing_users || [];
            const validUsers = users.filter(user => user && typeof user === 'string' && user.trim() !== '');
            window.TypingIndicator.updateTypingUsers(validUsers);
        }
    }
}

function sendMessage(data) {
    return window.wsQueue.sendMessage(data);
}
document.addEventListener('DOMContentLoaded', () => {
    initWebSocket();
});

function handleWebhookCreated(data) {
    if (data.channel_id && window.location.pathname.includes('/webhooks/')) {
        // Refresh webhook list if currently viewing webhooks
        location.reload();
    }
}

function handleWebhookDeleted(data) {
    if (data.channel_id && window.location.pathname.includes('/webhooks/')) {
        // Refresh webhook list if currently viewing webhooks  
        location.reload();
    }
}
window.socket = websocket;
window.sendMessage = sendMessage;