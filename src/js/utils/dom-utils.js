function $(id) {
    return document.getElementById(id);
}

function $$(selector) {
    return document.querySelector(selector);
}

function hide(elementOrId) {
    const el = typeof elementOrId === 'string' ? $(elementOrId) : elementOrId;
    if (el) el.style.display = 'none';
}

function show(elementOrId) {
    const el = typeof elementOrId === 'string' ? $(elementOrId) : elementOrId;
    if (el) el.style.display = 'block';
}

function setText(selector, text) {
    const el = $$(selector);
    if (el) el.textContent = text;
}

function createUserAvatarHTML(username, profilePicture) {
    if (profilePicture && profilePicture.trim() !== '') {
        return `<img src="${profilePicture}" alt="${username}" class="member-avatar">`;
    } else {
        return window.AvatarUtils ? window.AvatarUtils.show404pfp(username) : `<span class="member-initial">${username ? username.charAt(0).toUpperCase() : '?'}</span>`;
    }
}

function createUserAvatarElement(username, profilePicture) {
    if (window.AvatarCache) {
        const cached = window.AvatarCache.get(username, profilePicture);
        if (cached) {
            return cached;
        }
    }
    
    const avatarEl = document.createElement('div');
    const imgEl = document.createElement('img');
    imgEl.alt = username;
    imgEl.className = 'member-avatar';
    
    avatarEl.appendChild(imgEl);
    
    if (window.AvatarUtils) {
        window.AvatarUtils.setupAvatarWithFallback(imgEl, username, profilePicture);
    } else {
        imgEl.src = profilePicture || '';
    }
    
    if (window.AvatarCache) {
        window.AvatarCache.set(username, profilePicture, avatarEl);
    }
    
    return avatarEl;
}

window.domUtils = { $, $$, hide, show, setText, createUserAvatarHTML, createUserAvatarElement };
window.createUserAvatarHTML = createUserAvatarHTML;
window.createUserAvatarElement = createUserAvatarElement;