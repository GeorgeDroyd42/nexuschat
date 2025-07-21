document.addEventListener('DOMContentLoaded', () => {
  
  loadUserGuilds();
});

function addSidebarButton(config) {
    const sidebar = document.querySelector('.sidebar');
    
    const button = document.createElement('button');
    button.id = config.id;
    button.className = 'sidebar-btn';
    button.textContent = config.text || '+';
    button.title = config.title || '';
    button.onclick = config.onclick || (() => {});
    
    sidebar.insertBefore(button, sidebar.firstChild);
    
    return button;
}
async function loadUserGuilds() {
	    try {
	        const data = await API.guild.fetchUserGuilds();
        
        if (data.guilds) {
            const guildList = $('guild-list');
            guildList.innerHTML = '';
            
            data.guilds.forEach(guild => {
                const guildElement = createGuildElement(guild);
                guildList.appendChild(guildElement);
            });
            
            // Restore scroll position (inside the if block)
            const savedScrollPosition = sessionStorage.getItem('guildListScrollPosition');
            if (savedScrollPosition) {
                setTimeout(() => {
                    guildList.scrollTop = parseInt(savedScrollPosition);
                }, 50);
            }
        }
    } catch (error) {
        console.error('Error loading guilds:', error);
    }
}