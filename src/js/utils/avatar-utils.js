function createSecureAvatar(username, profilePicture, className = 'avatar-circle-sm') {    
    if (profilePicture && profilePicture.trim() !== '') {
        const img = document.createElement('img');
        img.className = className;
        img.src = profilePicture;
        img.alt = '';
        
        img.onerror = function() {
            const initial = document.createElement('span');
            let initialClass;
if (className === 'avatar-circle') {
    initialClass = 'avatar-circle-initial';
} else if (className === 'avatar-circle-sm') {
    initialClass = 'avatar-circle-sm-initial';
} else if (className === 'avatar-circle-lg') {
    initialClass = 'avatar-circle-lg-initial';
} else {
    initialClass = className.replace('avatar-circle', 'avatar-circle-initial');
}
initial.className = initialClass;
            initial.textContent = username ? username.charAt(0).toUpperCase() : '?';
            this.replaceWith(initial);
        };
        
        return img;
    } else {
        const initial = document.createElement('span');
        let initialClass;
if (className === 'avatar-circle') {
    initialClass = 'avatar-circle-initial';
} else if (className === 'avatar-circle-sm') {
    initialClass = 'avatar-circle-sm-initial';
} else if (className === 'avatar-circle-lg') {
    initialClass = 'avatar-circle-lg-initial';
} else {
    initialClass = className.replace('avatar-circle', 'avatar-circle-initial');
}
initial.className = initialClass;
        initial.textContent = username ? username.charAt(0).toUpperCase() : '?';
        return initial;
    }
}

window.AvatarUtils = { createSecureAvatar };