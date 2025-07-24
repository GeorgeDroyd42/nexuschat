async function fetchCurrentUser() {
    try {
        return await UserAPI.getCurrentUser();
    } catch (error) {
        console.error('Error fetching user data:', error);
        return null;
    }
}

function getCurrentChannelId() {
    const path = window.location.pathname;
    const parts = path.split('/');
    return parts[3];
}
function getCurrentGuildId() {
    const currentPath = window.location.pathname;
    if (currentPath.startsWith('/v/')) {
        const pathParts = currentPath.split('/v/')[1];
        const guildId = pathParts ? pathParts.split('/')[0] : null;
        
        if (guildId === 'main') {
            return null;
        }
        
        return guildId;
    }
    return null;
}
async function handleFormSubmission(options) {
    const {
        formElement,
        apiFunction,
        errorContainerId,
        onSuccess,
        operationName = 'form submission',
        validateForm = null
    } = options;

    if (validateForm && !validateForm()) {
        return false;
    }

    const formData = new FormData(formElement);
    
    try {
        const result = await apiFunction(formData);

        if (result.error) {
            displayErrorMessage(result.error, errorContainerId);
            return false;
        } else {
            if (onSuccess) {
                onSuccess(result);
            }
            return true;
        }
    } catch (error) {
        console.error(`Error during ${operationName}:`, error);
        return false;
    }
}
function isCurrentGuild(guildId) {
    const currentPath = window.location.pathname;
    return currentPath.startsWith(`/v/${guildId}`);
}

function isCurrentChannel(guildId, channelId) {
    const currentPath = window.location.pathname;
    return currentPath === `/v/${guildId}/${channelId}`;
}

function isCurrentPath(targetPath) {
    return window.location.pathname === targetPath;
}
window.handleFormSubmission = handleFormSubmission;