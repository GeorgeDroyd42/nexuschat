function show404pfp(username) {
    const initial = username ? username.charAt(0).toUpperCase() : '?';
    return `<span class="member-initial">${initial}</span>`;
}

function show404guild(guildName) {
    const initial = guildName ? guildName.charAt(0).toUpperCase() : '?';
    return `<span class="guild-initial">${initial}</span>`;
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