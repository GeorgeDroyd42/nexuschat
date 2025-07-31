async function updateChannelsHeader(guildId = null) {
    const channelsHeader = document.getElementById('channels-header');
    if (!channelsHeader) return;
    
    if (guildId) {
        try {
            const guildInfo = await window.GuildAPI.getInfo(guildId);
            if (guildInfo && guildInfo.name) {
                channelsHeader.textContent = guildInfo.name;
            }
        } catch (error) {
            console.error('Error updating channels header:', error);
            channelsHeader.textContent = 'Channels';
        }
    } else {
        channelsHeader.textContent = 'Channels';
    }
}

function displayErrorMessage(message, containerId = 'error-container', type = 'error') {
    const container = document.getElementById(containerId);
    if (container) {
        container.textContent = message;
        container.className = `error-container show`;
        container.style.display = 'block';
        
        setTimeout(() => {
            container.style.display = 'none';
        }, 5000);
    } else {
        console.error(message);
    }
}

function updateUserInterface(userData) {
    if (window.profileManager) {
        window.profileManager.updateAllElements(userData);
    }
}

function toggleLoading(elementId, isLoading) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    if (isLoading) {
        element.setAttribute('disabled', 'disabled');
        element.dataset.originalText = element.innerHTML;
        element.innerHTML = '<span class="loading-spinner"></span> Loading...';
    } else {
        element.removeAttribute('disabled');
        element.innerHTML = element.dataset.originalText || element.innerHTML;
    }
}
function truncateChannelName(channelName) {
    if (!channelName) {
        return '';
    }
    
    const screenWidth = window.innerWidth;
    
    if (screenWidth <= 480) {
        return channelName.length > 8 ? channelName.substring(0, 8) + '...' : channelName;
    } else if (screenWidth <= 768) {
        return channelName.length > 12 ? channelName.substring(0, 12) + '...' : channelName;
    } else {
        return channelName.length > 18 ? channelName.substring(0, 18) + '...' : channelName;
    }
}

function getChannelPlaceholder(channelName) {
    const truncated = truncateChannelName(channelName);
    const isTruncated = truncated !== channelName;
    return `Message #${truncated}${isTruncated ? '...' : ''}`;
}

window.getChannelPlaceholder = getChannelPlaceholder;
window.truncateChannelName = truncateChannelName;

window.getResponsiveChannelPlaceholder = getChannelPlaceholder;

function updateAllChannelNames() {
    document.querySelectorAll('.channel-name').forEach(element => {
        const channelId = element.closest('[data-channel-id]')?.dataset.channelId;
        if (channelId && element.dataset.originalName) {
            const truncated = truncateChannelName(element.dataset.originalName);
            element.textContent = `#${truncated}`;
        }
    });
    
    document.querySelectorAll('.guild-channel-item').forEach(element => {
        if (element.dataset.originalName) {
            const truncated = truncateChannelName(element.dataset.originalName);
            element.textContent = truncated;
        }
    });
}

document.addEventListener('DOMContentLoaded', () => {
    // Initial call to set proper state on page load
    updateResponsiveLayout();
    
    // Set up resize listener after DOM is ready
    window.addEventListener('resize', () => {
        const messageInput = document.getElementById('message-input');
        const channelTitle = document.querySelector('.channel-title');
        if (messageInput && channelTitle) {
            const channelName = channelTitle.textContent.replace('#', '').trim();
            messageInput.placeholder = getChannelPlaceholder(channelName);
        }
        
        updateAllChannelNames();
        updateResponsiveLayout();
    });
});

function updateResponsiveLayout() {
    const isMobile = window.innerWidth <= 768;
    const isTablet = window.innerWidth <= 1000 && window.innerWidth > 768;
    
    // Update channels sidebar visibility
    const channelsSidebar = document.querySelector('.channels-sidebar');
    if (channelsSidebar) {
        if (isMobile) {
            channelsSidebar.style.display = 'none';
        } else {
            // Only show if it's not explicitly hidden (has channels)
            if (!channelsSidebar.classList.contains('hidden')) {
                channelsSidebar.style.display = 'block';
            }
        }
    }
    
    // Update members sidebar behavior
    const membersSidebar = document.getElementById('members-sidebar');
    const mainContent = document.querySelector('.main-content');
    if (membersSidebar && mainContent) {
        if (isMobile) {
            membersSidebar.classList.remove('visible');
            mainContent.classList.remove('with-members');
        } else {
            // On desktop, restore members sidebar visibility if we're in a guild
            const currentGuildId = getCurrentGuildId();
            if (currentGuildId) {
                membersSidebar.classList.add('visible');
                mainContent.classList.add('with-members');
            }
        }
    }
    
    // Update mobile navigation visibility
    const guildToggle = document.getElementById('guild-toggle');
    const membersToggle = document.getElementById('members-toggle');
    if (guildToggle) guildToggle.style.display = isMobile ? 'flex' : 'none';
    if (membersToggle) membersToggle.style.display = isMobile ? 'block' : 'none';
    
    // Update sidebar mobile state
    const sidebar = document.querySelector('.sidebar');
    if (sidebar && !isMobile) {
        sidebar.classList.remove('mobile-visible');
    }
    
    // Handle profile modal vs mobile page
    const mobileProfilePage = document.getElementById('mobile-profile-page');
    const userModal = document.getElementById('user-modal');
    if (mobileProfilePage && mobileProfilePage.style.display === 'block' && !isMobile) {
        mobileProfilePage.style.display = 'none';
        if (userModal) userModal.classList.add('active');
    }
    if (userModal && userModal.classList.contains('active') && isMobile) {
        userModal.classList.remove('active');
        if (mobileProfilePage) mobileProfilePage.style.display = 'block';
    }
}

function clearFormFields(fieldIds, resetToDefaults = {}) {
    fieldIds.forEach(fieldId => {
        const element = $(fieldId);
        if (element) {
            if (resetToDefaults[fieldId]) {
                if (element.tagName === 'IMG') {
                    element.src = resetToDefaults[fieldId];
                } else {
                    element.value = resetToDefaults[fieldId];
                }
            } else if (element.tagName === 'INPUT' || element.tagName === 'TEXTAREA') {
                element.value = '';
            }
        }
    });
}

window.clearFormFields = clearFormFields;

function setActiveElement(selector, activeSelector, activeId) {
    document.querySelectorAll(selector).forEach(el => el.classList.remove('active'));
    const activeElement = document.querySelector(activeSelector.replace('{id}', activeId));
    if (activeElement) activeElement.classList.add('active');
}

function setActiveGuild(guildId) {
    setActiveElement('.guild-pill', '[data-guild-id="{id}"]', guildId);
}

function setActiveChannel(channelId) {
    setActiveElement('.channel-item', '.channel-item[data-channel-id="{id}"]', channelId);
}

function showConfirmationDialog(message, callback) {
    document.getElementById('confirm-message').textContent = message;
    window.modalManager.openModal('confirm-modal');
    window.confirmCallback = callback;
}

window.showConfirmationDialog = showConfirmationDialog;


window.setActiveGuild = setActiveGuild;
window.setActiveChannel = setActiveChannel;
window.setActiveElement = setActiveElement;