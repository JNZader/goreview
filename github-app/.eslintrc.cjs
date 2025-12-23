/** @type {import('eslint').Linter.Config} */
module.exports = {
  root: true,
  env: {
    node: true,
    es2022: true,
  },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
  ],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module',
    project: './tsconfig.json',
    tsconfigRootDir: __dirname,
  },
  plugins: ['@typescript-eslint'],
  rules: {
    // TypeScript specific
    '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    '@typescript-eslint/explicit-function-return-type': 'off',
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/no-non-null-assertion': 'warn',

    // General
    'no-console': 'warn',
    'prefer-const': 'error',
    'eqeqeq': ['error', 'always'],
    'no-unused-expressions': 'error',
  },
  ignorePatterns: [
    'dist/',
    'node_modules/',
    'coverage/',
    '*.config.js',
    '*.config.cjs',
    '*.test.ts',
  ],
};
