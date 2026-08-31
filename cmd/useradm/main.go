// Утилита управления пользователями.
//
// Регистрации через API в сервисе нет — пользователи заводятся только отсюда.
//
//	useradm create -login vladimir [-name "Владимир"]
//	useradm passwd -login vladimir
//	useradm list
//	useradm unlock -login vladimir
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"go_expense_service/internal/config"
	"go_expense_service/internal/repository"
	"go_expense_service/internal/service"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	if err := run(os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run(command string, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := config.ConnectDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	svc := service.NewUserService(repository.NewUserRepository(db), cfg)
	ctx := context.Background()

	switch command {
	case "create":
		return createUser(ctx, svc, args)
	case "passwd":
		return changePassword(ctx, svc, args)
	case "list":
		return listUsers(ctx, svc)
	case "unlock":
		return unlockUser(ctx, svc, args)
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("неизвестная команда %q", command)
	}
}

func createUser(ctx context.Context, svc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	login := fs.String("login", "", "логин пользователя")
	name := fs.String("name", "", "отображаемое имя (необязательно)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *login == "" {
		return errors.New("укажите -login")
	}

	password, err := readNewPassword()
	if err != nil {
		return err
	}

	user, err := svc.Create(ctx, *login, password, *name)
	if err != nil {
		return err
	}

	fmt.Printf("Пользователь создан: id=%d login=%s\n", user.ID, user.Login)
	return nil
}

func changePassword(ctx context.Context, svc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("passwd", flag.ExitOnError)
	login := fs.String("login", "", "логин пользователя")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *login == "" {
		return errors.New("укажите -login")
	}

	password, err := readNewPassword()
	if err != nil {
		return err
	}

	if err := svc.ChangePassword(ctx, *login, password); err != nil {
		return err
	}

	fmt.Println("Пароль изменён. Блокировка входа снята.")
	fmt.Println("Внимание: ранее выданные токены остаются действительными до истечения срока.")
	fmt.Println("Чтобы погасить их все — смените JWT_SECRET и перезапустите сервис.")
	return nil
}

func listUsers(ctx context.Context, svc *service.UserService) error {
	users, err := svc.List(ctx)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("Пользователей нет. Создайте первого: useradm create -login <логин>")
		return nil
	}

	fmt.Printf("%-5s %-20s %-25s %-10s %s\n", "ID", "ЛОГИН", "ИМЯ", "ПОПЫТКИ", "БЛОКИРОВКА ДО")
	now := time.Now().UTC()
	for _, u := range users {
		locked := "-"
		if u.IsLocked(now) {
			locked = u.LockedUntil.Local().Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-5d %-20s %-25s %-10d %s\n", u.ID, u.Login, u.Name, u.FailedAttempts, locked)
	}
	return nil
}

func unlockUser(ctx context.Context, svc *service.UserService, args []string) error {
	fs := flag.NewFlagSet("unlock", flag.ExitOnError)
	login := fs.String("login", "", "логин пользователя")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *login == "" {
		return errors.New("укажите -login")
	}

	if err := svc.Unlock(ctx, *login); err != nil {
		return err
	}

	fmt.Println("Блокировка входа снята, счётчик неудачных попыток обнулён.")
	return nil
}

// readNewPassword читает пароль дважды и сверяет.
//
// Флагом -password пароль не принимаем сознательно: он осел бы в истории shell
// и был бы виден в выводе ps.
func readNewPassword() (string, error) {
	fd := int(os.Stdin.Fd())

	// Не терминал (запуск из скрипта или пайпа) — читаем строку как есть,
	// подтверждение в этом случае бессмысленно.
	if !term.IsTerminal(fd) {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && line == "" {
			return "", fmt.Errorf("не удалось прочитать пароль: %w", err)
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	fmt.Print("Пароль: ")
	first, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать пароль: %w", err)
	}

	fmt.Print("Повторите пароль: ")
	second, err := term.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать пароль: %w", err)
	}

	if string(first) != string(second) {
		return "", errors.New("пароли не совпадают")
	}
	return string(first), nil
}

func usage() {
	fmt.Fprint(os.Stderr, `Управление пользователями сервиса учёта расходов.

Использование:
  useradm create -login <логин> [-name <имя>]   создать пользователя
  useradm passwd -login <логин>                 сменить пароль
  useradm list                                  список пользователей
  useradm unlock -login <логин>                 снять блокировку входа

Пароль запрашивается интерактивно и не отображается при вводе.
`)
}
