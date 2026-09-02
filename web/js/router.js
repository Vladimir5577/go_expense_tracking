// Хешевый роутер: #/login, #/expenses, #/expense/new, #/expense/12, #/categories.
//
// Хеш выбран сознательно — он не требует ни фолбэка на сервере, ни настройки
// nginx, а на React Router ложится один в один, включая параметры вида ':id'.

const routes = new Map();
let notFoundHandler = null;

export function route(pattern, handler) {
    routes.set(pattern, handler);
}

export function setNotFound(handler) {
    notFoundHandler = handler;
}

export function currentPath() {
    return location.hash.slice(1) || '/';
}

export function navigate(path) {
    if (currentPath() === path) {
        // hashchange не сработает — перерисовываем вручную.
        resolve();
        return;
    }
    location.hash = path;
}

// Сопоставляет шаблон с путём: '/expense/:id' и '/expense/12' дают {id: '12'}.
// Возвращает null, если шаблон не подошёл.
function match(pattern, path) {
    const patternParts = pattern.split('/');
    const pathParts = path.split('/');
    if (patternParts.length !== pathParts.length) return null;

    const params = {};
    for (const [i, part] of patternParts.entries()) {
        if (part.startsWith(':')) {
            params[part.slice(1)] = decodeURIComponent(pathParts[i]);
            continue;
        }
        if (part !== pathParts[i]) return null;
    }
    return params;
}

// Шаблоны проверяются в порядке объявления, поэтому конкретный '/expense/new'
// должен объявляться раньше, чем '/expense/:id'.
function resolve() {
    const path = currentPath();

    for (const [pattern, handler] of routes) {
        const params = match(pattern, path);
        if (params) {
            handler(params);
            return;
        }
    }

    notFoundHandler?.(path);
}

export function startRouter() {
    window.addEventListener('hashchange', resolve);
    resolve();
}
