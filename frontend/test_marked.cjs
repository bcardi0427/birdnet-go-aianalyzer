const { marked } = require('marked');
const html = marked.parse('# Hello\n<table><tr><td>Test</td></tr></table>\n<h3>Heading</h3>', { async: false });
console.log(html);
