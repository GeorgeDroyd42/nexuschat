const StatusCore = {
    updateMemberStatus(userId, isOnline, guildId) {
        if (getCurrentGuildId() !== guildId) return;
        
        const member = currentGuildMembers.find(m => m.user_id === userId);
        if (member) {
            member.is_online = isOnline;
            renderMembersList();
        }
    }
};

window.StatusCore = StatusCore;