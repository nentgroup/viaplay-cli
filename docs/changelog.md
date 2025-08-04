<!-- /docs/changelog.md -->

# Changelog

This page contains the complete release history of viaplay-cli.

<div id="changelog-content">Loading changelog...</div>

<script>
  // Fetch and render the CHANGELOG.md content
  fetch('https://raw.githubusercontent.com/nentgroup/viaplay-cli/main/CHANGELOG.md')
    .then(response => response.text())
    .then(data => {
      // Remove the first line (title) as we already have a title for this page
      const content = data.split('\n').slice(1).join('\n');
      document.getElementById('changelog-content').innerHTML = marked.parse(content);
    })
    .catch(error => {
      document.getElementById('changelog-content').innerHTML = 'Failed to load changelog: ' + error;
    });
</script>
