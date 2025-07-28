const StatusUI = {
    refreshGuildSettingsMembers() {
        if (window.guildMenuAPI && window.guildMenuAPI.currentGuildId === getCurrentGuildId()) {
            const settingsMembersList = document.querySelector('.members-list-container');
            const searchInput = document.querySelector('.members-search-input');
            if (settingsMembersList && searchInput) {
                window.guildMenuAPI.loadGuildMembers(settingsMembersList, searchInput);
            }
        }
    },
    
    refreshMainMemberList() {
        if (typeof getCurrentGuildId === 'function' && window.GuildMembers) {
            const currentGuildId = getCurrentGuildId();
            if (currentGuildId) {
                window.GuildMembers.loadGuildMembers(currentGuildId);
            }
        }
    }
};

window.StatusUI = StatusUI;