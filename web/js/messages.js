// Человеческие тексты для кодов ошибок API.
//
// Сервис отдаёт машинный код (`{"code": "category_has_expenses"}`) и намеренно
// не отдаёт текст: язык и формулировки — забота клиента.

const ERROR_MESSAGES = {
    network_error: 'Сервер недоступен. Проверьте соединение.',
    invalid_json: 'Не удалось отправить данные.',
    validation_failed: 'Проверьте заполненные поля.',
    unauthorized: 'Сессия истекла, войдите заново.',
    invalid_credentials: 'Неверный логин или пароль.',
    too_many_attempts: 'Слишком много попыток входа. Попробуйте через 15 минут.',
    not_found: 'Не найдено.',
    category_not_found: 'Категория не найдена.',
    expense_not_found: 'Трата не найдена.',
    category_name_exists: 'Категория с таким названием уже есть.',
    category_has_expenses: 'В категории есть траты — сначала перенесите или удалите их.',
    internal_error: 'Внутренняя ошибка сервиса. Попробуйте позже.',
};

export function errorMessage(error) {
    return ERROR_MESSAGES[error?.code] ?? 'Что-то пошло не так.';
}
