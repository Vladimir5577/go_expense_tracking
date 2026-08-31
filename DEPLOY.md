# Развертывание

Пример ниже для Ubuntu/Debian, каталога `/opt/expenses` и домена
`expenses.example.com`.

## 1. Собрать на локальной машине

```bash
npm --prefix frontend run css:build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main cmd/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o useradm cmd/useradm/main.go
```

На сервер отправляются только `main`, `useradm`, `web/`, `migrations/`.
Go, Node и `frontend/node_modules` на сервере не нужны.

## 2. Подготовить сервер

```bash
sudo useradd --system --home-dir /opt/expenses --shell /usr/sbin/nologin expenses
sudo mkdir -p /opt/expenses/{data,backups}
sudo chown -R expenses:expenses /opt/expenses

sudo apt update
sudo apt install -y nginx certbot python3-certbot-nginx sqlite3

GOOSE_VERSION=v3.27.2
curl -fsSL "https://github.com/pressly/goose/releases/download/${GOOSE_VERSION}/goose_linux_x86_64" -o /tmp/goose
sudo install -m 755 /tmp/goose /usr/local/bin/goose
```

## 3. Залить файлы

Локально:

```bash
rsync -avz --delete main useradm web/ migrations/ user@server:/tmp/expenses-deploy/
```

На сервере:

```bash
sudo rsync -a /tmp/expenses-deploy/ /opt/expenses/
sudo chown -R expenses:expenses /opt/expenses
sudo chmod +x /opt/expenses/main /opt/expenses/useradm
```

## 4. Создать конфиг

```bash
sudo -u expenses tee /opt/expenses/.env > /dev/null <<'EOF'
ENV=prod
SERVER_HOST=127.0.0.1
SERVER_PORT=8080
DB_PATH=./data/expenses.db
WEB_DIR=./web
JWT_SECRET=ЗАМЕНИТЬ
JWT_TTL_HOURS=168
APP_TIMEZONE=Europe/Moscow
BCRYPT_COST=12
LOGIN_MAX_ATTEMPTS=5
LOGIN_LOCK_MINUTES=15
EOF

sudo -u expenses sed -i "s/^JWT_SECRET=.*/JWT_SECRET=$(openssl rand -hex 32)/" /opt/expenses/.env
sudo chmod 600 /opt/expenses/.env
```

## 5. Накатить миграции

```bash
cd /opt/expenses
sudo -u expenses goose -dir migrations sqlite3 ./data/expenses.db up
```

## 6. Настроить запуск через systemd

```bash
sudo tee /etc/systemd/system/expenses.service > /dev/null <<'EOF'
[Unit]
Description=Expenses service
After=network.target

[Service]
User=expenses
Group=expenses
WorkingDirectory=/opt/expenses
ExecStart=/opt/expenses/main
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now expenses
curl -s localhost:8080/healthz
```

Ожидаемый ответ:

```json
{"status":"ok","service":"expenses"}
```

## 7. Создать пользователя

```bash
cd /opt/expenses
sudo -u expenses ./useradm create -login vladimir -name "Владимир"
```

## 8. Настроить nginx и HTTPS

```bash
sudo tee /etc/nginx/sites-available/expenses > /dev/null <<'EOF'
server {
    listen 80;
    server_name expenses.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

sudo ln -sf /etc/nginx/sites-available/expenses /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
sudo certbot --nginx -d expenses.example.com
```

## 9. Открыть порты

```bash
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

Порт `8080` не открываем: он только для nginx на этом же сервере.

## Обновление

Локально:

```bash
npm --prefix frontend run css:build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main cmd/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o useradm cmd/useradm/main.go
rsync -avz --delete main useradm web/ migrations/ user@server:/tmp/expenses-deploy/
```

На сервере:

```bash
sudo systemctl stop expenses
sudo rsync -a /tmp/expenses-deploy/ /opt/expenses/
sudo chown -R expenses:expenses /opt/expenses
cd /opt/expenses && sudo -u expenses goose -dir migrations sqlite3 ./data/expenses.db up
sudo systemctl start expenses
curl -s localhost:8080/healthz
```

## Бэкап

```bash
sudo -u expenses sqlite3 /opt/expenses/data/expenses.db \
  "VACUUM INTO '/opt/expenses/backups/expenses_$(date +%Y-%m-%d_%H%M).db'"
```

Скопировать бэкапы к себе:

```bash
rsync -avz user@server:/opt/expenses/backups/ ~/backups/expenses/
```

## Docker-вариант

Если используете `docker-compose.yml`, шаг с systemd не нужен:

```bash
cd /opt/expenses
goose -dir migrations sqlite3 ./data/expenses.db up
docker compose up -d
docker compose logs -f
```

## Диагностика

```bash
sudo journalctl -u expenses -f
sudo journalctl -u expenses -p err
```

Частые ошибки:

- `JWT_SECRET обязателен`: проверьте `/opt/expenses/.env`.
- `no such table: users`: не накатаны миграции.
- `attempt to write a readonly database`: неверный владелец файлов в `data/`.
- 404 на `/app.css`: не залит `web/` или неверный `WEB_DIR`.
