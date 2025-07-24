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

function createAvatarHTML(username, profilePicture) {
    if (profilePicture && profilePicture.trim() !== '') {
        const fallback = show404pfp(username);
        return `<img src="${profilePicture}" alt="${username}" class="member-avatar" onerror="this.outerHTML='${fallback.replace(/'/g, '&apos;')}'">`;
    } else {
        return show404pfp(username);
    }
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
    createAvatarHTML,
    setupAvatarWithFallback,
    show404pfp,
    show404guild
};