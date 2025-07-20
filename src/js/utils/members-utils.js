let currentGuildMembers = [];
let onlineMembers = new Set();

function updateMembersList(members, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    currentGuildMembers = members;
    
    onlineMembers.clear();
    members.forEach(member => {
        if (member.is_online) {
            onlineMembers.add(member.user_id);
        }
    });
    
    sortAndRenderMembers();
}

function sortAndRenderMembers() {
    currentGuildMembers.sort((a, b) => {
        const aOnline = onlineMembers.has(a.user_id);
        const bOnline = onlineMembers.has(b.user_id);
        
        if (aOnline !== bOnline) {
            return bOnline - aOnline; // Online first
        }
        return a.username.localeCompare(b.username); // Then alphabetical
    });
    renderMembersList();
}

function updateMemberStatus(userId, isOnline, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    if (isOnline) {
        onlineMembers.add(userId);
    } else {
        onlineMembers.delete(userId);
    }
    sortAndRenderMembers();
}

function addMemberToList(member, guildId) {
    if (getCurrentGuildId() !== guildId) return;
    
    const existingIndex = currentGuildMembers.findIndex(m => m.user_id === member.user_id);
    if (existingIndex === -1) {
        currentGuildMembers.push(member);
        currentGuildMembers.sort((a, b) => a.username.localeCompare(b.username));
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
    
    const onlineUsers = currentGuildMembers.filter(member => onlineMembers.has(member.user_id));
    const offlineUsers = currentGuildMembers.filter(member => !onlineMembers.has(member.user_id));
    
    onlineUsers.sort((a, b) => a.username.localeCompare(b.username));
    offlineUsers.sort((a, b) => a.username.localeCompare(b.username));
    
    onlineUsers.forEach(member => {
        const memberElement = createMemberElement(member.user_id, member.username, member.profile_picture);
        membersList.appendChild(memberElement);
    });
    
if (onlineUsers.length > 0 && offlineUsers.length > 0) {
    const separator = document.createElement('div');
    separator.className = 'sidebar-separator';
    separator.style.margin = '8px 0';
    membersList.appendChild(separator);
}
    
    offlineUsers.forEach(member => {
        const memberElement = createMemberElement(member.user_id, member.username, member.profile_picture);
        membersList.appendChild(memberElement);
    });
}

function createMemberElement(userID, username, profilePicture) {
    const memberElement = document.createElement('div');
    const isOnline = onlineMembers.has(userID);
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