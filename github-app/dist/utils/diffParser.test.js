import { describe, it, expect } from 'vitest';
import { parseDiff, detectLanguage } from './diffParser.js';
describe('parseDiff', () => {
    it('parses simple diff', () => {
        const diff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}`;
        const files = parseDiff(diff);
        expect(files).toHaveLength(1);
        expect(files[0]?.path).toBe('main.go');
        expect(files[0]?.status).toBe('modified');
        expect(files[0]?.language).toBe('go');
        expect(files[0]?.additions).toBe(1);
    });
    it('detects new files', () => {
        const diff = `diff --git a/new.ts b/new.ts
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/new.ts
@@ -0,0 +1 @@
+export const x = 1;`;
        const files = parseDiff(diff);
        expect(files[0]?.status).toBe('added');
    });
    it('detects deleted files', () => {
        const diff = `diff --git a/old.js b/old.js
deleted file mode 100644
index 1234567..0000000
--- a/old.js
+++ /dev/null
@@ -1 +0,0 @@
-const x = 1;`;
        const files = parseDiff(diff);
        expect(files[0]?.status).toBe('deleted');
    });
    it('handles multiple files', () => {
        const diff = `diff --git a/file1.ts b/file1.ts
index 1234567..abcdefg 100644
--- a/file1.ts
+++ b/file1.ts
@@ -1 +1,2 @@
 const a = 1;
+const b = 2;
diff --git a/file2.ts b/file2.ts
index 1234567..abcdefg 100644
--- a/file2.ts
+++ b/file2.ts
@@ -1 +1,2 @@
 const c = 3;
+const d = 4;`;
        const files = parseDiff(diff);
        expect(files).toHaveLength(2);
        expect(files[0]?.path).toBe('file1.ts');
        expect(files[1]?.path).toBe('file2.ts');
    });
    it('counts additions and deletions', () => {
        const diff = `diff --git a/test.js b/test.js
index 1234567..abcdefg 100644
--- a/test.js
+++ b/test.js
@@ -1,3 +1,3 @@
 const a = 1;
-const b = 2;
+const b = 3;
 const c = 4;`;
        const files = parseDiff(diff);
        expect(files[0]?.additions).toBe(1);
        expect(files[0]?.deletions).toBe(1);
    });
});
describe('detectLanguage', () => {
    const cases = [
        ['main.go', 'go'],
        ['index.ts', 'typescript'],
        ['app.tsx', 'typescript'],
        ['script.js', 'javascript'],
        ['main.py', 'python'],
        ['lib.rs', 'rust'],
        ['App.java', 'java'],
        ['config.yaml', 'yaml'],
        ['data.json', 'json'],
        ['style.css', 'css'],
        ['README.md', 'markdown'],
        ['unknown.xyz', 'unknown'],
    ];
    it.each(cases)('detectLanguage(%s) = %s', (path, expected) => {
        expect(detectLanguage(path)).toBe(expected);
    });
});
//# sourceMappingURL=diffParser.test.js.map