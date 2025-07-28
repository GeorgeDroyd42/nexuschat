const fs = require('fs');
const { execSync } = require('child_process');

const coreFiles = [
    'src/js/utils/dom-utils.js',
    'src/js/utils/api-utils.js',
    'src/js/utils/format-utils.js',
    'src/js/utils/ui-utils.js',
    'src/js/utils/session-utils.js',
    'src/js/utils/component-utils.js',
    'src/js/utils/members-utils.js',
    'src/js/utils.js',
    'src/js/csrf.js', 
    'src/js/api/AuthAPI.js',
    'src/js/api/UserAPI.js',
    'src/js/utils/navutils.js',
    'src/js/channel/channel-manager.js',
    'src/js/api/BaseAPI.js',
    'src/js/modal.js',
    'src/js/invite.js',
    'src/js/utils/profile-utils.js',
    'src/js/utils/avatar-utils.js', 
    'src/js/utils/modal-utils.js',
];

const guildFiles = [
    'src/js/status/core.js',
    'src/js/status/ui.js', 
    'src/js/status/sockets.js',
    'src/js/guild/guild-manager.js',
    'src/js/guild/members.js',
    'src/js/guild/navigation.js',
    'src/js/guild/ui.js',
    'src/js/guild/buttons.js',
    'src/js/channel/ui.js',
    'src/js/channel/handlers.js',
    'src/js/api/ChannelAPI.js',
    'src/js/menu/channel-menu-api.js',
    'src/js/message/message-manager.js',
    'src/js/message/handlers.js',
    'src/js/message/loading.js',
    'src/js/message/ui.js',
    'src/js/utils/embed-utils.js',
    'src/js/menu/profile-menu-api.js',
    'src/js/menu/guild-menu-api.js',
    'src/js/menu/menu-renderer.js',
    'src/js/utils/permission-manager.js',
    'src/js/api/MessageAPI.js',
    'src/js/api/GuildAPI.js',
    'src/js/sidebar.js',
    'src/js/utils/websocket-queue.js',
    'src/js/sockets.js',
    'src/js/menu/context-menu.js',
    'src/js/utils/typing-indicator.js'
];
const authFiles = [
    'src/js/auth/forms.js',
    'src/js/auth/ui.js'
];
const adminFiles = [
    'src/js/admin/search.js',
    'src/js/admin/ui.js',
    'src/js/admin/buttons.js',    
    'src/js/admin.js'];

function createBundle(files, outputName) {
    let bundledContent = '';
    files.forEach(file => {
        try {
            const content = fs.readFileSync(file, 'utf8');
            bundledContent += `${content}\n`;
        } catch (err) {
            console.error(`Error reading ${file}:`, err.message);
        }
    });
    
    const tempFile = `public/js/${outputName}.temp.js`;
    const minifiedFile = `public/js/${outputName}.minified.js`;
    const finalFile = `public/js/${outputName}`;
    
    fs.writeFileSync(tempFile, bundledContent);
    
    try {
        execSync(`npx terser ${tempFile} --compress --mangle -o ${minifiedFile}`, { stdio: 'inherit' });
        
        if (!fs.existsSync(minifiedFile)) {
            throw new Error('Minified file was not created');
        }
        
        execSync(`npx javascript-obfuscator ${minifiedFile} --output ${finalFile} --compact true --control-flow-flattening true --control-flow-flattening-threshold 0.5 --dead-code-injection true --dead-code-injection-threshold 0.3 --identifier-names-generator hexadecimal --string-array true --string-array-encoding base64 --string-array-threshold 0.5`, { stdio: 'inherit' });
        
        fs.unlinkSync(tempFile);
        fs.unlinkSync(minifiedFile);
        
        const originalSize = Math.round(bundledContent.length / 1024);
        const obfuscatedSize = Math.round(fs.readFileSync(finalFile).length / 1024);
        const savings = Math.round(((originalSize - obfuscatedSize) / originalSize) * 100);
        
        console.log(`${outputName}: ${originalSize}KB → ${obfuscatedSize}KB (${savings}% reduction, obfuscated)`);
    } catch (error) {
        console.log(`Obfuscation failed for ${outputName}, trying terser only...`);
        try {
            execSync(`npx terser ${tempFile} --compress --mangle -o ${finalFile}`, { stdio: 'inherit' });
            fs.unlinkSync(tempFile);
            console.log(`${outputName} created (terser minified only)`);
        } catch (terserError) {
            fs.renameSync(tempFile, finalFile);
            console.log(`${outputName} created (unminified fallback)`);
        }
    }
}

createBundle(coreFiles, 'core-bundle.js');
createBundle(guildFiles, 'guild-bundle.js');
createBundle(authFiles, 'auth-bundle.js');
createBundle(adminFiles, 'admin-bundle.js');