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
