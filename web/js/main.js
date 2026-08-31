import { api, clearToken, getToken, onUnauthorized } from './api.js';
import { navigate, route, setNotFound, startRouter } from './router.js';
import { resetState, setState } from './store.js';
import { renderLogin } from './views/login.js';

const root = document.querySelector('#app');

// Единственная точка обработки протухшей сессии: api.js зовёт её при 401,
// поэтому каждое место вызова не проверяет статус самостоятельно.
onUnauthorized(() => {
    resetState();
    navigate('/login');
});

route('/login', () => renderLogin(root));

route('/expenses', async () => {
    if (!(await requireAuth())) return;
    const { renderExpenses } = await import('./views/expenses.js');
    renderExpenses(root);
});

route('/categories', async () => {
    if (!(await requireAuth())) return;
    const { renderCategories } = await import('./views/categories.js');
    renderCategories(root);
});

route('/', () => navigate(getToken() ? '/expenses' : '/login'));
setNotFound(() => navigate('/'));

// requireAuth проверяет не только наличие токена, но и его пригодность:
// токен живёт 7 дней и может протухнуть, пока вкладка была открыта.
async function requireAuth() {
    if (!getToken()) {
        navigate('/login');
        return false;
    }

    try {
        const user = await api.me();
        setState({ user });
        return true;
    } catch {
        // 401 уже обработан в onUnauthorized; остальные ошибки тоже уводят
        // на логин — без пользователя показывать нечего.
        clearToken();
        navigate('/login');
        return false;
    }
}

export function logout() {
    clearToken();
    resetState();
    navigate('/login');
}

startRouter();
