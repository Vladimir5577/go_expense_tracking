// Конфиг лежит в корне, а не в frontend/: eslint ищет его, поднимаясь от
// проверяемых файлов (web/js), и конфиг из соседнего каталога он не увидит.
//
// Импортов здесь нет намеренно. node_modules живёт в frontend/, из корня пакеты
// не резолвятся, а тянуть их по относительному пути внутрь чужого каталога —
// хрупко. Поэтому вместо js.configs.recommended правила перечислены явно:
// их немного, и видно, за чем именно следим.

export default [
    {
        files: ['web/js/**/*.js'],

        languageOptions: {
            ecmaVersion: 2024,
            sourceType: 'module',
            globals: {
                document: 'readonly',
                window: 'readonly',
                location: 'readonly',
                localStorage: 'readonly',
                fetch: 'readonly',
                setTimeout: 'readonly',
                clearTimeout: 'readonly',
                URL: 'readonly',
                console: 'readonly',
            },
        },

        linterOptions: {
            reportUnusedDisableDirectives: 'error',
        },

        rules: {
            // Главное, ради чего линтер здесь вообще нужен: опечатка в имени и
            // забытый после правки импорт иначе всплывают только в браузере.
            'no-undef': 'error',
            'no-unused-vars': ['error', { argsIgnorePattern: '^_' }],

            // Ошибки, которые молча меняют поведение.
            'no-const-assign': 'error',
            'no-dupe-args': 'error',
            'no-dupe-keys': 'error',
            'no-duplicate-case': 'error',
            'no-duplicate-imports': 'error',
            'no-fallthrough': 'error',
            'no-func-assign': 'error',
            'no-self-assign': 'error',
            'no-sparse-arrays': 'error',
            'no-unreachable': 'error',
            'no-unsafe-negation': 'error',
            'use-isnan': 'error',
            'valid-typeof': 'error',

            'no-async-promise-executor': 'error',

            eqeqeq: ['error', 'smart'],
            'prefer-const': 'error',
        },
    },
];
