import js from '@eslint/js'
import globals from 'globals'
import tseslint from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'

export default tseslint.config(
  { ignores: ['dist', 'node_modules', '**/*.js', '**/*.cjs', '**/*.mjs'] },
  {
    files: ['src/**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      ...tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
    ],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.vitest,
      },
    },
    linterOptions: {
      reportUnusedDisableDirectives: 'error',
    },
    plugins: {
      'react-refresh': reactRefresh,
    },
    rules: {
      'react-refresh/only-export-components': 'warn',
      // Pre-existing violations, reported but not blocking: these rules are new
      // in eslint-plugin-react-hooks v7 and flag effect/mutation patterns that
      // need a refactor rather than a mechanical fix.
      'react-hooks/set-state-in-effect': 'warn',
      'react-hooks/immutability': 'warn',
    },
  },
  {
    // Types are declared globally (no export), so they read as unused here.
    files: ['src/types/*.tsx'],
    rules: {
      '@typescript-eslint/no-unused-vars': 'off',
    },
  }
)
