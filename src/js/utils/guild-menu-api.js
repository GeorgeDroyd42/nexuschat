class GuildMenuAPI extends ChannelMenuAPI {
    constructor() {
        super();
        this.currentGuild = null;
        this.currentGuildId = null;
    }

    renderButtons(modalId, guildData) {
        this.currentGuild = guildData;
        this.currentGuildId = guildData?.guild_id;
        
        const modal = document.getElementById(modalId);
        if (!modal) return;
        
        const content = modal.querySelector('.profile-menu-content');
        if (!content) return;
        
        this.createSidebarLayout(content);
        this.setupTabSwitching(content);
        this.populateTabContent();
        this.disableContextMenu(content);
    }

    disableContextMenu(content) {
        content.addEventListener('contextmenu', (e) => {
            e.preventDefault();
            return false;
        });
    }

    renderItem(panel, item) {
        if (item.type === 'guild-info') {
            this.renderGuildInfo(panel);
        } else if (item.type === 'members-list') {
            this.renderMembersList(panel);
        } else {
            super.renderItem(panel, item);
        }
    }

    async renderMembersList(panel) {
        const membersContainer = this.createElement('div', 'members-container');
        
        const searchContainer = this.createElement('div', 'members-search-container');
        const searchInput = this.createElement('input', 'members-search-input');
        searchInput.type = 'text';
        searchInput.placeholder = 'Search members...';
        
        const membersList = this.createElement('div', 'members-list-container');
        
        searchContainer.appendChild(searchInput);
        membersContainer.appendChild(searchContainer);
        membersContainer.appendChild(membersList);
        panel.appendChild(membersContainer);
        
        // Load and display members
        this.loadGuildMembers(membersList, searchInput);
    }

    async loadGuildMembers(membersList, searchInput) {
        if (!this.currentGuildId) return;
        
        try {
            const data = await API.guild.getMembers(this.currentGuildId);
            if (data && data.members) {
                this.displayMembers(data.members, membersList, searchInput);
            }
        } catch (error) {
            console.error('Error loading guild members:', error);
            membersList.innerHTML = '<p>Failed to load members</p>';
        }
    }

    displayMembers(members, membersList, searchInput) {
        const allMembers = members;
        
        const renderFilteredMembers = (filter = '') => {
            const filtered = allMembers.filter(member => 
                member.username.toLowerCase().includes(filter.toLowerCase())
            );
            
            membersList.innerHTML = '';
            
            if (filtered.length === 0) {
                membersList.innerHTML = '<p>No members found</p>';
                return;
            }
            
            filtered.forEach(member => {
                const memberElement = this.createElement('div', 'member-list-item');
                const isOnline = member.is_online;
                
                memberElement.innerHTML = `
                    <div class="member-avatar-container">
                        ${window.createUserAvatarHTML ? window.createUserAvatarHTML(member.username, member.profile_picture) : `<span>${member.username.charAt(0)}</span>`}
                        <div class="member-status ${isOnline ? 'online' : 'offline'}"></div>
                    </div>
                    <span class="member-name">${member.username}</span>
                    <span class="member-role">${member.role || 'Member'}</span>
                `;
                
                membersList.appendChild(memberElement);
            });
        };
        
        // Initial render
        renderFilteredMembers();
        
        // Search functionality
        searchInput.addEventListener('input', (e) => {
            renderFilteredMembers(e.target.value);
        });
    }

    renderGuildInfo(panel) {
        const guildContainer = document.createElement('div');
        guildContainer.className = 'profile-picture-container';
        
        const guildPic = document.createElement('img');
        guildPic.className = 'profile-picture-large';
        guildPic.alt = 'Guild Picture';
        
        guildContainer.appendChild(guildPic);
        panel.appendChild(guildContainer);
        
        if (window.AvatarUtils) {
            window.AvatarUtils.setupAvatarWithFallback(
                guildPic, 
                this.currentGuild?.name || 'Guild',
                this.currentGuild?.profile_picture_url
            );
        }

        const fields = [
            { label: 'GUILD NAME', id: 'guild-info-name', value: this.currentGuild?.name || 'Loading...' },
            { label: 'CREATED', id: 'guild-info-created', value: this.currentGuild?.created_at ? new Date(this.currentGuild.created_at).toLocaleDateString() : 'Loading...' },
            { label: 'GUILD ID', id: 'guild-info-id', value: this.currentGuild?.guild_id || 'Loading...', class: 'user-id' }
        ];

        this.renderInfoSection(panel, fields, 'guild-info-section');

        const descFields = [
            { label: 'DESCRIPTION', id: 'guild-info-description', value: this.currentGuild?.description || 'No description set', class: 'bio-display' }
        ];

        this.renderInfoSection(panel, descFields, 'guild-description-section');
    }
}

window.guildMenuAPI = new GuildMenuAPI();

guildMenuAPI
.addTab('overview', 'Overview')
.addTab('settings', 'Settings')
.addTab('members', 'Members')

.addToTab('overview', { type: 'guild-info' })
.addToTab('overview', { type: 'separator' })
.addToTab('overview', { type: 'button', text: 'Create Invite', action: async () => {
    console.log('Create invite for guild:', guildMenuAPI.currentGuildId);
}, color: '#4CAF50' })

.addToTab('settings', { type: 'textarea', label: 'GUILD NAME', id: 'guild-name-edit', value: '', placeholder: 'Enter guild name...', rows: 1 })
.addToTab('settings', { type: 'textarea', label: 'DESCRIPTION', id: 'guild-description-edit', value: '', placeholder: 'Describe your guild...', rows: 3 })
.addToTab('settings', { type: 'button', text: 'Save Changes', action: async () => {
    const newName = document.getElementById('guild-name-edit').value.trim();
    const newDescription = document.getElementById('guild-description-edit').value.trim();
    console.log('Save guild settings:', { name: newName, description: newDescription });
}, color: '#2196F3' })

.addToTab('members', { type: 'members-list' });