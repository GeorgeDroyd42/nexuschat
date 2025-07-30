let currentGuildMembers = [];


function updateMembersList(members, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    currentGuildMembers = members;
    renderMembersList();
}



function addMemberToList(member, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    currentGuildMembers.push(member);
    renderMembersList();
}

function removeMemberFromList(userId, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    currentGuildMembers = currentGuildMembers.filter(m => m.user_id !== userId);
    renderMembersList();
}

function renderMembersList() {
    const membersList = $('members-list');
    if (!membersList) return;
    
    membersList.innerHTML = '';
    
    const sortedMembers = [...currentGuildMembers].sort((a, b) => {
        if (a.is_online !== b.is_online) {
            return a.is_online ? -1 : 1;
        }
        return a.username.toLowerCase().localeCompare(b.username.toLowerCase());
    });
    
    let separatorAdded = false;
    let foundOfflineUser = false;
    
    sortedMembers.forEach(member => {
        if (!member.is_online && !foundOfflineUser && !separatorAdded) {
            foundOfflineUser = true;
            const hasOnlineUsers = sortedMembers.some(m => m.is_online);
            if (hasOnlineUsers) {
                const separator = document.createElement('div');
                separator.className = 'sidebar-separator';
                separator.style.margin = '8px 0';
                membersList.appendChild(separator);
                separatorAdded = true;
            }
        }
        
        const memberElement = createMemberElement(member.user_id, member.username, member.profile_picture, member.is_online);
        membersList.appendChild(memberElement);
    });
}

function createMemberElement(userID, username, profilePicture, isOnline) {
    const memberElement = document.createElement('div');
    memberElement.className = `member-item ${isOnline ? 'online' : 'offline'}`;
    memberElement.setAttribute('data-user-id', userID);
    
    const avatarContainer = document.createElement('div');
    avatarContainer.className = 'member-avatar-container';
    
    const avatar = window.AvatarUtils.createSecureAvatar(username, profilePicture);
    const status = document.createElement('div');
    status.className = `member-status ${isOnline ? 'online' : 'offline'}`;
    
    avatarContainer.appendChild(avatar);
    avatarContainer.appendChild(status);
    
    const nameSpan = document.createElement('span');
    nameSpan.className = 'member-name';
    nameSpan.textContent = username || 'Unknown';
    
    memberElement.appendChild(avatarContainer);
    memberElement.appendChild(nameSpan);
        
    return memberElement;
}
window.updateMembersList = updateMembersList;
window.addMemberToList = addMemberToList;
window.removeMemberFromList = removeMemberFromList;
