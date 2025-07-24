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
    
    if (window.MessageAPI && window.MessageAPI.currentChannelId) {
        window.MessageAPI.loadChannelMessages(window.MessageAPI.currentChannelId);
    }
};

websocket.onclose = function(event) {
        console.log('WebSocket disconnected', event.code, event.reason);
        
        if (window.MessageAPI) {
            window.MessageAPI.messageCache.clear();
        }
        
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
                    handleUserStatusChanged(data);
                    break;
                case 'typing_update':
                    handleTypingUpdate(data);
                    break;                    
            }
}
function handleUserStatusChanged(data) {
    if (isCurrentGuild(data.guild_id)) {
        if (window.updateMemberStatus) {
            window.updateMemberStatus(data.user_id, data.is_online, data.guild_id);
        }
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
    if (data.member && data.guild_id) {
        addMemberToList(data.member, data.guild_id);
    }
}

function handleMemberLeft(data) {
    if (data.member && data.guild_id) {
        removeMemberFromList(data.member.user_id, data.guild_id);
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
    if (data.channel && data.guild_id && window.channelManager) {
        window.channelManager.loadChannels(data.guild_id);
    }
}

async function handleChannelDeleted(data) {
    if (data.channel && data.guild_id && window.channelManager) {
        if (isCurrentChannel(data.guild_id, data.channel.channel_id)) {
            const channelsData = await API.guild.getChannels(data.guild_id);
            if (channelsData.channels && channelsData.channels.length > 0) {
                window.channelManager.handleChannelSelect(channelsData.channels[0]);
            } else {
                window.location.href = `/v/${data.guild_id}`;
            }
        }
        window.channelManager.loadChannels(data.guild_id);
    }
}

function handleChannelUpdated(data) {
    if (data.channel && data.guild_id && window.channelManager) {
        if (isCurrentChannel(data.guild_id, data.channel.channel_id)) {
            const channelTitle = document.querySelector('.channel-title');
            if (channelTitle) {
                const nameElement = channelTitle.querySelector('h2');
                const descElement = channelTitle.querySelector('p');
                
                if (nameElement) {
                    nameElement.textContent = `# ${data.channel.name}`;
                }
                
                if (data.channel.description) {
                    if (descElement) {
                        descElement.textContent = data.channel.description;
                        descElement.style.display = 'block';
                    } else {
                        const newDesc = document.createElement('p');
                        newDesc.textContent = data.channel.description;
                        channelTitle.appendChild(newDesc);
                    }
                } else if (descElement) {
                    descElement.style.display = 'none';
                }
            }
        }
        window.channelManager.loadChannels(data.guild_id);
    }
}

function handleNewMessage(data) {
    if (window.MessageAPI) {
        // Only add message if it belongs to the currently viewed channel
        if (data.channel_id && data.channel_id === window.MessageAPI.currentChannelId) {
            window.MessageAPI.addNewMessage(data);
        }
    }
}

function handleMessageDeleted(data) {
    if (window.MessageAPI) {
        // Only remove message if it belongs to the currently viewed channel
        if (data.channel_id && data.channel_id === window.MessageAPI.currentChannelId) {
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
                    window.AvatarUtils.setupAvatarWithFallback(userAvatar, userData.username, userData.profile_picture);
                }
            });
        }   
        
        if (typeof loadUserGuilds === 'function') {
            loadUserGuilds();
        }
        
        if (typeof getCurrentGuildId === 'function' && typeof loadGuildMembers === 'function') {
            const currentGuildId = getCurrentGuildId();
            if (currentGuildId) {
                loadGuildMembers(currentGuildId);
            }
        }
        
        if (window.MessageAPI && window.MessageAPI.currentChannelId && typeof window.MessageAPI.loadChannelMessages === 'function') {
            window.MessageAPI.loadChannelMessages(window.MessageAPI.currentChannelId);
        }
        
        console.log(`Username updated: ${data.old_username} → ${data.new_username}`);
    }
}
function handleTypingUpdate(data) {
    if (window.TypingIndicator) {
        if (data.channel_id === window.MessageAPI?.currentChannelId) {
            const users = data.typing_users || [];
            const validUsers = users.filter(user => user && typeof user === 'string' && user.trim() !== '');
            window.TypingIndicator.updateTypingUsers(validUsers);
        }
    }
}

function sendMessage(data) {
    return window.wsQueue.sendMessage(data);
}

window.socket = websocket;
window.sendMessage = sendMessage;