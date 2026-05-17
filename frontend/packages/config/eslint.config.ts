import type { Linter } from 'eslint';

const config: Linter.Config[] = [
  {
    files: ['**/*.ts', '**/*.tsx'],
    languageOptions: {
      parser: undefined, // @typescript-eslint/parser at runtime
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
      },
    },
    rules: {
      'no-unused-vars': 'off',
      'no-console': 'warn',
      'prefer-const': 'error',
      'no-var': 'error',
      eqeqeq: ['error', 'always'],
    },
  },
  {
    ignores: ['**/dist/**', '**/node_modules/**', '**/.turbo/**'],
  },
];

export default config;
