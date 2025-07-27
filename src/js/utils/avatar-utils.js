function createInitialFallback(name, className) {
    const initial = name ? name.charAt(0).toUpperCase() : '?';
    return `<span class="${className}">${initial}</span>`;
}

function show404pfp(username) {
    return createInitialFallback(username, 'member-initial');
}

function show404guild(guildName) {
    return createInitialFallback(guildName, 'guild-initial');
}

function setupAvatarWithFallback(imgElement, username, profilePicture) {
    if (profilePicture && profilePicture.trim() !== '') {
        imgElement.src = profilePicture;
        imgElement.onerror = () => {
            imgElement.outerHTML = show404pfp(username);
        };
    } else {
        imgElement.outerHTML = show404pfp(username);
    }
}

window.AvatarUtils = {
    setupAvatarWithFallback,
    show404pfp,
    show404guild
};