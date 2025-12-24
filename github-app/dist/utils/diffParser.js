/**
 * Parse a unified diff string into file objects.
 */
export function parseDiff(diff) {
    const files = [];
    const fileBlocks = diff.split(/^diff --git/m).slice(1);
    for (const block of fileBlocks) {
        const file = parseFileBlock('diff --git' + block);
        if (file) {
            files.push(file);
        }
    }
    return files;
}
/**
 * Determine file status from diff block content
 */
function determineFileStatus(block) {
    if (block.includes('new file mode'))
        return 'added';
    if (block.includes('deleted file mode'))
        return 'deleted';
    if (block.includes('rename from'))
        return 'renamed';
    return 'modified';
}
/**
 * Extract hunks content and count additions/deletions
 */
function extractHunks(lines) {
    let content = '';
    let additions = 0;
    let deletions = 0;
    let inHunk = false;
    for (const line of lines) {
        if (line.startsWith('@@')) {
            inHunk = true;
            content += line + '\n';
            continue;
        }
        if (!inHunk)
            continue;
        content += line + '\n';
        if (line.startsWith('+') && !line.startsWith('+++')) {
            additions++;
        }
        else if (line.startsWith('-') && !line.startsWith('---')) {
            deletions++;
        }
    }
    return { content: content.trim(), additions, deletions };
}
function parseFileBlock(block) {
    const lines = block.split('\n');
    const firstLine = lines[0];
    if (!firstLine)
        return null;
    const headerMatch = firstLine.match(/^diff --git a\/(.+) b\/(.+)$/);
    if (!headerMatch)
        return null;
    const oldPath = headerMatch[1] ?? '';
    const newPath = headerMatch[2] ?? '';
    const { content, additions, deletions } = extractHunks(lines);
    return {
        path: newPath || 'unknown',
        oldPath: oldPath !== newPath ? oldPath : undefined,
        status: determineFileStatus(block),
        language: detectLanguage(newPath || ''),
        content,
        isBinary: block.includes('Binary files'),
        additions,
        deletions,
    };
}
/**
 * Detect programming language from file extension.
 */
export function detectLanguage(path) {
    const ext = path.split('.').pop()?.toLowerCase() || '';
    const languageMap = {
        // JavaScript/TypeScript
        js: 'javascript',
        jsx: 'javascript',
        ts: 'typescript',
        tsx: 'typescript',
        mjs: 'javascript',
        cjs: 'javascript',
        // Go
        go: 'go',
        // Python
        py: 'python',
        pyi: 'python',
        // Rust
        rs: 'rust',
        // Java/Kotlin
        java: 'java',
        kt: 'kotlin',
        kts: 'kotlin',
        // C/C++
        c: 'c',
        h: 'c',
        cpp: 'cpp',
        cc: 'cpp',
        cxx: 'cpp',
        hpp: 'cpp',
        // Ruby
        rb: 'ruby',
        // PHP
        php: 'php',
        // Shell
        sh: 'shell',
        bash: 'shell',
        zsh: 'shell',
        // Config/Data
        json: 'json',
        yaml: 'yaml',
        yml: 'yaml',
        toml: 'toml',
        xml: 'xml',
        // Markup
        md: 'markdown',
        html: 'html',
        css: 'css',
        scss: 'scss',
    };
    return languageMap[ext] || 'unknown';
}
//# sourceMappingURL=diffParser.js.map