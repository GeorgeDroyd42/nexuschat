const GuildMembers = {
    async loadGuildMembers(guildID) {
        try {
            const data = await API.guild.getMembers(guildID);
            
            if (data.error) {
                console.error('Error loading members:', data.error);
                return;
            }
            
            updateMembersList(data.members, guildID);
        } catch (error) {
            console.error('Error loading guild members:', error);
        }
    },

    async getUsernameByID(userID) {
        try {
            const data = await UserAPI.getUserProfile(userID);
            return data.username || 'Unknown';
        } catch {
            return 'Unknown';
        }
    },

    async getUserProfilePicture(userID) {
        try {
            const data = await UserAPI.getUserProfile(userID);
            return data.profile_picture || '';
        } catch {
            return '';
        }
    },

    setupMembersSidebar(guildId) {
        if (window.innerWidth > 768) {
            document.getElementById('members-sidebar').classList.add('visible');
            document.querySelector('.main-content').classList.add('with-members');
        }
        this.loadGuildMembers(guildId);
    }
};

window.GuildMembers = GuildMembers;