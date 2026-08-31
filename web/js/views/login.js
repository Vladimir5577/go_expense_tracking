import { api, setToken } from '../api.js';
import { errorMessage } from '../messages.js';
import { navigate } from '../router.js';
import { setState } from '../store.js';

export function renderLogin(root) {
    root.innerHTML = `
        <div class="flex min-h-dvh items-center justify-center p-4">
            <div class="card w-full max-w-sm p-6">
                <h1 class="text-2xl font-semibold">Расходы</h1>
                <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
                    Войдите, чтобы продолжить
                </p>

                <form id="login-form" class="mt-6 space-y-4" novalidate>
                    <div>
                        <label class="label" for="login">Логин</label>
                        <input class="field" id="login" name="login" required
                               autocomplete="username" autocapitalize="none"
                               autocorrect="off" spellcheck="false">
                    </div>

                    <div>
                        <label class="label" for="password">Пароль</label>
                        <input class="field" id="password" name="password" type="password"
                               required autocomplete="current-password">
                    </div>

                    <p id="login-error"
                       class="hidden rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950/60 dark:text-red-300"></p>

                    <button class="btn-primary w-full" type="submit" id="login-submit">Войти</button>
                </form>
            </div>
        </div>
    `;

    const form = root.querySelector('#login-form');
    const submit = root.querySelector('#login-submit');
    const errorBox = root.querySelector('#login-error');

    form.addEventListener('submit', async (event) => {
        event.preventDefault();

        const login = form.login.value.trim();
        const password = form.password.value;

        if (!login || !password) {
            showError(errorBox, 'Заполните логин и пароль.');
            return;
        }

        hideError(errorBox);
        submit.disabled = true;
        submit.textContent = 'Входим…';

        try {
            const result = await api.login(login, password);
            setToken(result.token);
            setState({ user: result.user });
            navigate('/expenses');
        } catch (error) {
            showError(errorBox, errorMessage(error));
            form.password.value = '';
            form.password.focus();
        } finally {
            submit.disabled = false;
            submit.textContent = 'Войти';
        }
    });

    form.login.focus();
}

function showError(box, text) {
    box.textContent = text;
    box.classList.remove('hidden');
}

function hideError(box) {
    box.classList.add('hidden');
}
