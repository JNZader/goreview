export interface DiffFile {
  path: string;
  oldPath?: string;
  status: 'added' | 'modified' | 'deleted' | 'renamed';
  language: string;
  content: string;
  isBinary: boolean;
  additions: number;
  deletions: number;
}

/**
 * Parse a unified diff string into file objects.
 */
export function parseDiff(diff: string): DiffFile[] {
  const files: DiffFile[] = [];
  const fileBlocks = diff.split(/^diff --git/m).slice(1);

  for (const block of fileBlocks) {
    const file = parseFileBlock('diff --git' + block);
    if (file) {
      files.push(file);
    }
  }

  return files;
}

function parseFileBlock(block: string): DiffFile | null {
  const lines = block.split('\n');

  // Parse header
  const firstLine = lines[0];
  if (!firstLine) return null;

  const headerMatch = firstLine.match(/^diff --git a\/(.+) b\/(.+)$/);
  if (!headerMatch) return null;

  const oldPath = headerMatch[1] ?? '';
  const newPath = headerMatch[2] ?? '';

  // Determine status
  let status: DiffFile['status'] = 'modified';
  if (block.includes('new file mode')) {
    status = 'added';
  } else if (block.includes('deleted file mode')) {
    status = 'deleted';
  } else if (block.includes('rename from')) {
    status = 'renamed';
  }

  // Check for binary
  const isBinary = block.includes('Binary files');

  // Extract content (everything after @@)
  let content = '';
  let additions = 0;
  let deletions = 0;

  let inHunk = false;
  for (const line of lines) {
    if (line.startsWith('@@')) {
      inHunk = true;
      content += line + '\n';
    } else if (inHunk) {
      content += line + '\n';

      if (line.startsWith('+') && !line.startsWith('+++')) {
        additions++;
      } else if (line.startsWith('-') && !line.startsWith('---')) {
        deletions++;
      }
    }
  }

  return {
    path: newPath || 'unknown',
    oldPath: oldPath !== newPath ? oldPath : undefined,
    status,
    language: detectLanguage(newPath || ''),
    content: content.trim(),
    isBinary,
    additions,
    deletions,
  };
}

/**
 * Detect programming language from file extension.
 */
export function detectLanguage(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() || '';

  const languageMap: Record<string, string> = {
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
