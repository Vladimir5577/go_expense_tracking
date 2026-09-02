import { getToken, onUnauthorized } from './api.js';
import { navigate, route, setNotFound, startRouter } from './router.js';
import { requireAuth } from './session.js';
import { resetState } from './store.js';
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
    const { renderExpenses } = await import('./views/expenses/index.js');
    renderExpenses(root);
});

// Добавление и изменение траты — отдельный экран, а не форма внутри списка.
// '/expense/new' объявлен раньше '/expense/:id': роутер берёт первый
// подошедший шаблон.
route('/expense/new', async () => {
    if (!(await requireAuth())) return;
    const { renderExpenseForm } = await import('./views/expenses/form.js');
    renderExpenseForm(root, null);
});

route('/expense/:id', async ({ id }) => {
    if (!(await requireAuth())) return;

    const expenseId = Number(id);
    if (!Number.isInteger(expenseId) || expenseId <= 0) {
        navigate('/expenses');
        return;
    }

    const { renderExpenseForm } = await import('./views/expenses/form.js');
    renderExpenseForm(root, expenseId);
});

route('/categories', async () => {
    if (!(await requireAuth())) return;
    const { renderCategories } = await import('./views/categories.js');
    renderCategories(root);
});

route('/', () => navigate(getToken() ? '/expenses' : '/login'));
setNotFound(() => navigate('/'));

startRouter();
