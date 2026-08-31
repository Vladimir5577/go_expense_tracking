import { api } from '../api.js';
import { CATEGORY_COLORS, colorForCategory } from '../colors.js';
import { errorMessage } from '../messages.js';
import { getState, setState } from '../store.js';
import { appHeader, emptyState, escapeHtml, skeleton, toast } from './layout.js';

let root = null;
// id редактируемой категории; null — форма работает в режиме создания.
let editingId = null;
let selectedColor = CATEGORY_COLORS[0].key;

export function renderCategories(container) {
    root = container;
    editingId = null;

    root.innerHTML = `
        ${appHeader('categories')}
        <main class="mx-auto max-w-screen-sm px-3 pb-8" style="padding-bottom: calc(2rem + env(safe-area-inset-bottom))">
            <div id="form" class="pt-3"></div>
            <div id="list" class="pt-4">${skeleton(4)}</div>
        </main>
    `;

    root.querySelector('#logout').addEventListener('click', async () => {
        const { logout } = await import('../main.js');
        logout();
    });

    renderForm();
    void load();
}

async function load() {
    try {
        const categories = await api.listCategories();
        setState({ categories });
        renderList();
    } catch (error) {
        toast(errorMessage(error));
    }
}

// --- форма ---

function renderForm() {
    root.querySelector('#form').innerHTML = `
        <details id="form-details" class="card overflow-hidden">
            <summary class="flex min-h-11 cursor-pointer list-none items-center px-4 font-medium select-none">
                <span id="form-title">+ Новая категория</span>
            </summary>

            <form id="category-form" class="space-y-3 border-t border-neutral-200 p-4 dark:border-neutral-800" novalidate>
                <div>
                    <label class="label" for="name">Название</label>
                    <input class="field" id="name" name="name" required maxlength="100"
                           autocomplete="off" placeholder="Продукты">
                </div>

                <div>
                    <label class="label" for="description">Описание</label>
                    <input class="field" id="description" name="description" maxlength="500"
                           autocomplete="off" placeholder="Необязательно">
                </div>

                <div>
                    <span class="label">Цвет</span>
                    <div id="colors" class="flex flex-wrap gap-2"></div>
                </div>

                <div class="flex gap-2">
                    <button class="btn-primary flex-1" type="submit" id="category-submit">Создать</button>
                    <button class="btn-secondary hidden" type="button" id="category-cancel">Отмена</button>
                </div>
            </form>
        </details>
    `;

    renderColorPicker();
    root.querySelector('#category-form').addEventListener('submit', onSubmit);
    root.querySelector('#category-cancel').addEventListener('click', resetForm);
}

// Цвет выбирается из фиксированной палитры, а не свободным вводом: так набор
// цветов остаётся согласованным и заведомо читается в обеих темах.
function renderColorPicker() {
    const box = root.querySelector('#colors');

    box.innerHTML = CATEGORY_COLORS.map(
        (color) => `
        <button type="button" data-color="${color.key}" title="${color.title}" aria-label="${color.title}"
                class="size-11 rounded-full ${color.dot} ${
                    color.key === selectedColor
                        ? 'ring-2 ring-neutral-900 ring-offset-2 dark:ring-neutral-100 dark:ring-offset-neutral-900'
                        : ''
                }"></button>
    `,
    ).join('');

    for (const button of box.querySelectorAll('[data-color]')) {
        button.addEventListener('click', () => {
            selectedColor = button.dataset.color;
            renderColorPicker();
        });
    }
}

async function onSubmit(event) {
    event.preventDefault();

    const form = event.currentTarget;
    const submit = form.querySelector('#category-submit');

    const name = form.name.value.trim();
    if (!name) {
        toast('Введите название категории.');
        form.name.focus();
        return;
    }

    const body = {
        name,
        description: form.description.value.trim(),
        color: selectedColor,
    };

    submit.disabled = true;
    submit.textContent = 'Сохраняем…';

    try {
        if (editingId === null) {
            await api.createCategory(body);
        } else {
            await api.updateCategory(editingId, body);
        }
        resetForm();
        await load();
    } catch (error) {
        toast(errorMessage(error));
    } finally {
        submit.disabled = false;
        submit.textContent = editingId === null ? 'Создать' : 'Сохранить';
    }
}

function startEditing(category) {
    editingId = category.id;
    selectedColor = colorForCategory(category).key;

    const details = root.querySelector('#form-details');
    const form = root.querySelector('#category-form');

    details.open = true;
    root.querySelector('#form-title').textContent = 'Изменить категорию';
    form.name.value = category.name;
    form.description.value = category.description;
    renderColorPicker();

    root.querySelector('#category-submit').textContent = 'Сохранить';
    root.querySelector('#category-cancel').classList.remove('hidden');

    details.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    form.name.focus();
}

function resetForm() {
    editingId = null;
    selectedColor = CATEGORY_COLORS[0].key;

    const form = root.querySelector('#category-form');
    if (!form) return;

    root.querySelector('#form-title').textContent = '+ Новая категория';
    root.querySelector('#category-submit').textContent = 'Создать';
    root.querySelector('#category-cancel').classList.add('hidden');
    form.name.value = '';
    form.description.value = '';
    renderColorPicker();
}

// --- список ---

function renderList() {
    const { categories } = getState();
    const box = root.querySelector('#list');

    if (categories.length === 0) {
        box.innerHTML = emptyState('Категорий пока нет. Заведите первую — например, «Продукты».');
        return;
    }

    box.innerHTML = `
        <ul class="card divide-y divide-neutral-200 overflow-hidden dark:divide-neutral-800">
            ${categories.map(categoryRow).join('')}
        </ul>
    `;

    for (const row of box.querySelectorAll('[data-category]')) {
        const category = categories.find((item) => item.id === Number(row.dataset.category));

        row.querySelector('[data-edit]').addEventListener('click', () => startEditing(category));
        row.querySelector('[data-delete]').addEventListener('click', async () => {
            if (!confirm(`Удалить категорию «${category.name}»?`)) return;

            try {
                await api.deleteCategory(category.id);
                await load();
            } catch (error) {
                // Отдельная подсказка для 409: сервис запрещает удалять
                // категорию, на которой висят траты.
                toast(errorMessage(error));
            }
        });
    }
}

function categoryRow(category) {
    const color = colorForCategory(category);

    return `
        <li data-category="${category.id}" class="flex items-stretch">
            <button type="button" data-edit
                    class="flex min-h-14 flex-1 items-center gap-3 px-3 py-2 text-left">
                <span class="size-3 shrink-0 rounded-full ${color.dot}"></span>
                <span class="min-w-0 flex-1">
                    <span class="block truncate">${escapeHtml(category.name)}</span>
                    ${
                        category.description
                            ? `<span class="block truncate text-xs text-neutral-500 dark:text-neutral-400">
                                   ${escapeHtml(category.description)}
                               </span>`
                            : ''
                    }
                </span>
            </button>

            <button type="button" data-delete aria-label="Удалить"
                    class="flex min-h-14 w-11 shrink-0 items-center justify-center text-neutral-400
                           hover:text-red-600 dark:hover:text-red-400">
                <svg viewBox="0 0 20 20" fill="currentColor" class="size-5">
                    <path d="M8.75 3h2.5a.75.75 0 0 1 .75.75V4.5h3.25a.75.75 0 0 1 0 1.5h-.6l-.7 9.1A2.25 2.25 0 0 1 11.7 17H8.3a2.25 2.25 0 0 1-2.25-1.9L5.35 6h-.6a.75.75 0 0 1 0-1.5H8v-.75A.75.75 0 0 1 8.75 3Z"/>
                </svg>
            </button>
        </li>
    `;
}
