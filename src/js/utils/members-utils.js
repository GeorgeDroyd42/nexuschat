let currentGuildMembers = [];


function updateMembersList(members, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    currentGuildMembers = members;
    renderMembersList();
}



function updateMemberStatus(userId, isOnline, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    const member = currentGuildMembers.find(m => m.user_id === userId);
    if (member) {
        member.is_online = isOnline;
        renderMembersList();
    }
}

function addMemberToList(member, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    const existingIndex = currentGuildMembers.findIndex(m => m.user_id === member.user_id);
    if (existingIndex === -1) {
        currentGuildMembers.push(member);
        
        renderMembersList();
    }
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
    
    const onlineUsers = currentGuildMembers.filter(member => member.is_online);
    const offlineUsers = currentGuildMembers.filter(member => !member.is_online);
    
    onlineUsers.sort((a, b) => a.username.localeCompare(b.username));
    offlineUsers.sort((a, b) => a.username.localeCompare(b.username));
    
    onlineUsers.forEach(member => {
        const memberElement = createMemberElement(member.user_id, member.username, member.profile_picture, true);
        membersList.appendChild(memberElement);
    });
    
if (onlineUsers.length > 0 && offlineUsers.length > 0) {
    const separator = document.createElement('div');
    separator.className = 'sidebar-separator';
    separator.style.margin = '8px 0';
    membersList.appendChild(separator);
}
    
    offlineUsers.forEach(member => {
        const memberElement = createMemberElement(member.user_id, member.username, member.profile_picture, false);
        membersList.appendChild(memberElement);
    });
}

function createMemberElement(userID, username, profilePicture, isOnline) {
    const memberElement = document.createElement('div');
    memberElement.className = `member-item ${isOnline ? 'online' : 'offline'}`;
    memberElement.setAttribute('data-user-id', userID);
    
    memberElement.innerHTML = `
        <div class="member-avatar-container">
            ${window.createUserAvatarHTML(username, profilePicture)}
            <div class="member-status ${isOnline ? 'online' : 'offline'}"></div>
        </div>
        <span class="member-name">${username || 'Unknown'}</span>
    `;
    
    const imgEl = memberElement.querySelector('img.member-avatar');
    if (imgEl) {
        imgEl.onerror = function() {
            this.outerHTML = window.AvatarUtils ? window.AvatarUtils.show404pfp(username) : `<span class="member-initial">${username ? username.charAt(0).toUpperCase() : '?'}</span>`;
        };
    }
        
    return memberElement;
}
window.updateMembersList = updateMembersList;
window.addMemberToList = addMemberToList;
window.removeMemberFromList = removeMemberFromList;
window.updateMemberStatus = updateMemberStatus;