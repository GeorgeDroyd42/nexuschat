function createSecureAvatar(username, profilePicture, className = 'member-avatar') {
    const container = document.createElement('div');
    container.className = 'avatar-container';
    
    if (profilePicture && profilePicture.trim() !== '') {
        const img = document.createElement('img');
        img.className = className;
        img.src = profilePicture;
        img.alt = '';
        
        img.onerror = function() {
            const initial = document.createElement('span');
            initial.className = className.includes('guild') ? 'guild-initial' : className.replace('avatar', 'initial');
            initial.textContent = username ? username.charAt(0).toUpperCase() : '?';
            this.replaceWith(initial);
        };
        
        container.appendChild(img);
    } else {
        const initial = document.createElement('span');
        initial.className = className.includes('guild') ? 'guild-initial' : className.replace('avatar', 'initial');
        initial.textContent = username ? username.charAt(0).toUpperCase() : '?';
        container.appendChild(initial);
    }
    
    return container;
}

window.AvatarUtils = { createSecureAvatar };