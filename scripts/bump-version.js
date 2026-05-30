const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const versionFilePath = path.join(__dirname, '../version.txt');

try {
  if (fs.existsSync(versionFilePath)) {
    let content = fs.readFileSync(versionFilePath, 'utf8').trim();
    // Parse version number from content (e.g. "aianalyzer version 0.25" or "aianalyzer-25")
    const match = content.match(/(\d+\.\d+)/);
    if (match) {
      const currentVersion = parseFloat(match[1]);
      // Increment by 0.01
      const nextVersion = (currentVersion + 0.01).toFixed(2);
      const newContent = `aianalyzer version ${nextVersion}`;
      fs.writeFileSync(versionFilePath, newContent + '\n', 'utf8');
      console.log(`📈 Automatically bumped version in version.txt: ${content} -> ${newContent}`);
      
      // Stage the updated version.txt file
      execSync('git add version.txt', { cwd: path.join(__dirname, '..') });
    } else {
      console.error('❌ Could not parse version number from version.txt');
      process.exit(1);
    }
  } else {
    console.error('❌ version.txt not found');
    process.exit(1);
  }
} catch (error) {
  console.error('❌ Failed to bump version:', error.message);
  process.exit(1);
}
