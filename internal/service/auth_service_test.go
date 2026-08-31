package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/config"
	"go_expense_service/internal/dto"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/model"
)

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:         strings.Repeat("s", 32),
		JWTTTL:            168 * time.Hour,
		Timezone:          time.UTC,
		BcryptCost:        bcrypt.MinCost, // тесты не должны ждать 250 мс на каждый вход
		LoginMaxAttempts:  3,
		LoginLockDuration: 15 * time.Minute,
		Clock:             helper.NewClock(),
	}
}

func testUser(t *testing.T, password string) *model.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return &model.User{ID: 1, Login: "vladimir", PasswordHash: string(hash)}
}

func TestLoginSuccess(t *testing.T) {
	repo := newFakeUserRepo(testUser(t, "correct-horse"))
	svc := NewAuthService(repo, testConfig())

	token, expiresAt, user, err := svc.Login(context.Background(), dto.LoginRequest{
		Login: "vladimir", Password: "correct-horse",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Fatal("токен пустой")
	}
	if user.ID != 1 {
		t.Fatalf("user.ID = %d, ожидался 1", user.ID)
	}
	if !expiresAt.After(time.Now()) {
		t.Fatal("срок действия токена должен быть в будущем")
	}
}

// Несуществующий логин и неверный пароль обязаны давать один и тот же ответ:
// иначе по разнице ответов перебираются валидные логины.
func TestLoginUnknownAndWrongPasswordAreIndistinguishable(t *testing.T) {
	repo := newFakeUserRepo(testUser(t, "correct-horse"))
	svc := NewAuthService(repo, testConfig())
	ctx := context.Background()

	_, _, _, errUnknown := svc.Login(ctx, dto.LoginRequest{Login: "нет-такого", Password: "any-password"})
	_, _, _, errWrong := svc.Login(ctx, dto.LoginRequest{Login: "vladimir", Password: "wrong-password"})

	var unknownApp, wrongApp *apperr.Error
	if !errors.As(errUnknown, &unknownApp) || !errors.As(errWrong, &wrongApp) {
		t.Fatalf("ожидались доменные ошибки, получено: %v / %v", errUnknown, errWrong)
	}
	if unknownApp.Code != apperr.CodeInvalidCredentials || wrongApp.Code != apperr.CodeInvalidCredentials {
		t.Fatalf("коды ошибок различаются: %s vs %s", unknownApp.Code, wrongApp.Code)
	}
}

func TestLoginLocksAfterMaxAttempts(t *testing.T) {
	cfg := testConfig()
	repo := newFakeUserRepo(testUser(t, "correct-horse"))
	svc := NewAuthService(repo, cfg)
	ctx := context.Background()

	for i := 1; i < cfg.LoginMaxAttempts; i++ {
		_, _, _, err := svc.Login(ctx, dto.LoginRequest{Login: "vladimir", Password: "wrong"})
		if !errors.Is(err, apperr.New(apperr.CodeInvalidCredentials, "")) {
			t.Fatalf("попытка %d: ожидался invalid_credentials, получено %v", i, err)
		}
		if repo.setFailureLocked != nil {
			t.Fatalf("блокировка сработала раньше времени, на попытке %d", i)
		}
	}

	// Последняя разрешённая попытка выставляет блокировку.
	if _, _, _, err := svc.Login(ctx, dto.LoginRequest{Login: "vladimir", Password: "wrong"}); err == nil {
		t.Fatal("ожидалась ошибка")
	}
	if repo.setFailureLocked == nil {
		t.Fatal("после исчерпания попыток должна выставляться блокировка")
	}
	if repo.setFailureAttempts != 0 {
		t.Fatalf("счётчик должен обнуляться при блокировке, получено %d", repo.setFailureAttempts)
	}

	// Пока блокировка активна, даже верный пароль не пускает.
	_, _, _, err := svc.Login(ctx, dto.LoginRequest{Login: "vladimir", Password: "correct-horse"})
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Code != apperr.CodeTooManyAttempts {
		t.Fatalf("ожидался too_many_attempts, получено: %v", err)
	}
}

func TestLoginResetsCounterOnSuccess(t *testing.T) {
	repo := newFakeUserRepo(testUser(t, "correct-horse"))
	svc := NewAuthService(repo, testConfig())
	ctx := context.Background()

	if _, _, _, err := svc.Login(ctx, dto.LoginRequest{Login: "vladimir", Password: "wrong"}); err == nil {
		t.Fatal("ожидалась ошибка на неверном пароле")
	}
	if repo.setFailureAttempts != 1 {
		t.Fatalf("счётчик = %d, ожидался 1", repo.setFailureAttempts)
	}

	if _, _, _, err := svc.Login(ctx, dto.LoginRequest{Login: "vladimir", Password: "correct-horse"}); err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !repo.resetCalled {
		t.Fatal("после успешного входа счётчик неудачных попыток должен обнуляться")
	}
}

func TestLoginRespectsExpiredLock(t *testing.T) {
	repo := newFakeUserRepo(testUser(t, "correct-horse"))
	past := time.Now().UTC().Add(-time.Minute)
	repo.users[1].LockedUntil = &past

	svc := NewAuthService(repo, testConfig())

	// Истёкшая блокировка не должна мешать входу.
	if _, _, _, err := svc.Login(context.Background(), dto.LoginRequest{
		Login: "vladimir", Password: "correct-horse",
	}); err != nil {
		t.Fatalf("вход при истёкшей блокировке: %v", err)
	}
}
