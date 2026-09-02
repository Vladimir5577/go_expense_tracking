// Управление категориями: список и форма создания/редактирования.

import { api } from '../api.js';
import { CATEGORY_COLORS, colorForCategory } from '../colors.js';
import { errorMessage } from '../messages.js';
import { getState, setState } from '../store.js';
import { actionSheet, emptyState, listRow, onRowMenu, renderShell, setBusy, skeleton, toast } from './layout.js';
import { formPanel } from './panel.js';

const DEFAULT_COLOR = CATEGORY_COLORS[0].key;

let sections = null;
let panel = null;

export function renderCategories(root) {
    const main = renderShell(
        root,
        'categories',
        `
        <div id="form" class="pt-3"></div>
        <div id="list" class="pt-4">${skeleton(4)}</div>
    `,
    );

    sections = {
        form: main.querySelector('#form'),
        list: main.querySelector('#list'),
    };

    panel = createCategoryForm(sections.form);
    onRowMenu(sections.list, openMenu);

    void load();
}

async function load() {
    setBusy(sections.list, true);

    try {
        setState({ categories: await api.listCategories() });
        renderList();
    } catch (error) {
        toast(errorMessage(error));
    } finally {
        setBusy(sections.list, false);
    }
}

// --- форма ---

// Цвет выбирается из фиксированной палитры, а не свободным вводом: так набор
// цветов остаётся согласованным и заведомо читается в обеих темах. Под кружками
// лежат обычные радиокнопки — выбор тогда не только виден, но и объявляется
// скринридером, а стрелки на клавиатуре работают сами собой.
function colorPicker() {
    return `
        <fieldset>
            <legend class="label">Цвет</legend>
            <div class="flex flex-wrap gap-2">
                ${CATEGORY_COLORS.map(
                    (color) => `
                    <label class="cursor-pointer">
                        <input type="radio" name="color" value="${color.key}" class="peer sr-only"
                               ${color.key === DEFAULT_COLOR ? 'checked' : ''}>
                        <span class="sr-only">${color.title}</span>
                        <span aria-hidden="true"
                              class="block size-11 rounded-full ${color.dot}
                                     peer-checked:ring-2 peer-checked:ring-neutral-900 peer-checked:ring-offset-2
                                     peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2
                                     peer-focus-visible:outline-blue-500
                                     dark:peer-checked:ring-neutral-100 dark:peer-checked:ring-offset-neutral-900"></span>
                    </label>
                `,
                ).join('')}
            </div>
        </fieldset>
    `;
}

function createCategoryForm(box) {
    return formPanel(box, {
        createTitle: '+ Новая категория',
        editTitle: 'Изменить категорию',
        createSubmit: 'Создать',
        editSubmit: 'Сохранить',
        focusField: 'name',
        clearOnReset: ['name', 'description'],
        onReset: (form) => {
            form.color.value = DEFAULT_COLOR;
        },
        onSubmit: submitCategory,
        fields: `
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

            ${colorPicker()}
        `,
    });
}

async function submitCategory(form, editingId) {
    const name = form.name.value.trim();
    if (!name) {
        toast('Введите название категории.');
        form.name.focus();
        return false;
    }

    const body = {
        name,
        description: form.description.value.trim(),
        color: form.color.value,
    };

    try {
        if (editingId === null) {
            await api.createCategory(body);
        } else {
            await api.updateCategory(editingId, body);
        }
    } catch (error) {
        toast(errorMessage(error));
        return false;
    }

    await load();
    return true;
}

// --- список ---

function renderList() {
    const { categories } = getState();

    if (categories.length === 0) {
        sections.list.innerHTML = emptyState('Категорий пока нет. Заведите первую — например, «Продукты».');
        return;
    }

    sections.list.innerHTML = `
        <ul class="card divide-y divide-neutral-200 overflow-hidden dark:divide-neutral-800">
            ${categories
                .map((category) =>
                    listRow({
                        id: category.id,
                        accent: colorForCategory(category).bar,
                        title: category.name,
                        subtitle: category.description,
                    }),
                )
                .join('')}
        </ul>
    `;
}

function openMenu(id) {
    const category = findCategory(id);
    if (!category) return;

    actionSheet({
        title: category.name,
        subtitle: category.description,
        actions: [
            { title: 'Изменить', onClick: () => startEditing(id) },
            { title: 'Удалить', danger: true, onClick: () => void removeCategory(id) },
        ],
    });
}

function startEditing(id) {
    const category = findCategory(id);
    if (!category) return;

    panel.edit(category.id, {
        name: category.name,
        description: category.description,
        color: colorForCategory(category).key,
    });
}

async function removeCategory(id) {
    const category = findCategory(id);
    if (!category) return;

    if (panel.editingId === id) {
        panel.reset();
    }

    try {
        await api.deleteCategory(id);
        await load();
    } catch (error) {
        // Сервис запрещает удалять категорию, на которой висят траты, —
        // об этом расскажет текст из messages.js.
        toast(errorMessage(error));
        return;
    }

    toast(`Категория «${category.name}» удалена`, {
        variant: 'neutral',
        action: 'Отменить',
        onAction: () => void restoreCategory(category),
    });
}

// Как и у трат, отмена — это создание заново с новым id. Удалить категорию
// можно только пустой, так что терять нечему.
async function restoreCategory(category) {
    try {
        await api.createCategory({
            name: category.name,
            description: category.description,
            color: category.color,
        });
        await load();
    } catch (error) {
        toast(errorMessage(error));
    }
}

function findCategory(id) {
    return getState().categories.find((category) => category.id === id);
}
