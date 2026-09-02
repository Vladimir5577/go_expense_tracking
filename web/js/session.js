// Сессия пользователя: вход, выход и проверка доступа к экрану.
//
// Отдельный модуль нужен, чтобы вьюхи не импортировали main.js ради logout —
// иначе получается цикл main → views → main.

import { api, clearToken, getToken, setToken } from './api.js';
import { navigate } from './router.js';
import { getState, resetState, setState } from './store.js';

export async function signIn(login, password) {
    const result = await api.login(login, password);
    setToken(result.token);
    setState({ user: result.user });
    return result.user;
}

export function logout() {
    clearToken();
    resetState();
    navigate('/login');
}

// Пускает на защищённый экран.
//
// Токен живёт 7 дней и может протухнуть, пока вкладка была открыта, поэтому
// один раз за загрузку страницы его пригодность проверяется запросом /api/me.
// Дальше пользователь берётся из состояния: переключение вкладок не должно
// стоить лишнего похода в сеть, а протухший токен всё равно поймает
// onUnauthorized на первом же реальном запросе.
export async function requireAuth() {
    if (!getToken()) {
        navigate('/login');
        return false;
    }

    if (getState().user) {
        return true;
    }

    try {
        setState({ user: await api.me() });
        return true;
    } catch {
        // 401 уже обработан в onUnauthorized; остальные ошибки тоже уводят
        // на логин — без пользователя показывать нечего.
        clearToken();
        navigate('/login');
        return false;
    }
}
