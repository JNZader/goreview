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
export declare function parseDiff(diff: string): DiffFile[];
/**
 * Detect programming language from file extension.
 */
export declare function detectLanguage(path: string): string;
//# sourceMappingURL=diffParser.d.ts.map