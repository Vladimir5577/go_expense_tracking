// Раскрывающаяся панель с формой, работающая в двух режимах: создание и
// редактирование. У трат и у категорий формы разные, а вот переключение
// режимов, подписи кнопок, блокировка на время запроса и сброс — одинаковые,
// поэтому живут одним куском здесь.

const ICON_CHEVRON = `
    <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"
         class="size-5 shrink-0 text-neutral-400 transition-transform duration-200 group-open:rotate-180">
        <path d="M5.2 7.3a.75.75 0 0 1 1.1 0L10 11l3.7-3.7a.75.75 0 1 1 1.1 1.06l-4.25 4.25a.75.75 0 0 1-1.1 0L5.2 8.36a.75.75 0 0 1 0-1.06Z"/>
    </svg>
`;

export function formPanel(
    box,
    { createTitle, editTitle, createSubmit, editSubmit, fields, focusField, clearOnReset = [], onReset, onSubmit },
) {
    box.innerHTML = `
        <details class="group card overflow-hidden">
            <summary class="flex min-h-11 cursor-pointer list-none items-center gap-2 px-4
                            font-medium select-none">
                <span data-panel-title>${createTitle}</span>
                <span class="flex-1"></span>
                ${ICON_CHEVRON}
            </summary>

            <form class="space-y-3 border-t border-neutral-200 p-4 dark:border-neutral-800" novalidate>
                ${fields}

                <div class="flex gap-2">
                    <button class="btn-primary flex-1" type="submit" data-panel-submit>
                        ${createSubmit}
                    </button>
                    <button class="btn-secondary hidden" type="button" data-panel-cancel>Отмена</button>
                </div>
            </form>
        </details>
    `;

    const details = box.querySelector('details');
    const form = box.querySelector('form');
    const title = box.querySelector('[data-panel-title]');
    const submit = box.querySelector('[data-panel-submit]');
    const cancel = box.querySelector('[data-panel-cancel]');

    let editingId = null;

    function focus() {
        form.elements[focusField]?.focus();
    }

    function reset() {
        editingId = null;
        title.textContent = createTitle;
        submit.textContent = createSubmit;
        cancel.classList.add('hidden');

        for (const name of clearOnReset) {
            const field = form.elements[name];
            if (field) field.value = '';
        }

        onReset?.(form);
    }

    form.addEventListener('submit', async (event) => {
        event.preventDefault();

        submit.disabled = true;
        submit.textContent = 'Сохраняем…';

        try {
            const ok = await onSubmit(form, editingId);
            if (ok) {
                reset();
                // Фокус возвращается в первое поле: траты вносят пачками,
                // и лишний тап по полю суммы каждый раз — заметная цена.
                focus();
            }
        } finally {
            submit.disabled = false;
            submit.textContent = editingId === null ? createSubmit : editSubmit;
        }
    });

    cancel.addEventListener('click', reset);

    return {
        form,

        get editingId() {
            return editingId;
        },

        edit(id, values) {
            editingId = id;

            details.open = true;
            title.textContent = editTitle;
            submit.textContent = editSubmit;
            cancel.classList.remove('hidden');

            for (const [name, value] of Object.entries(values)) {
                const field = form.elements[name];
                if (field) field.value = String(value ?? '');
            }

            details.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            focus();
        },

        reset,
    };
}
