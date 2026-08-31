# Installation


1. Copy .env from .env.example to .env
```bash
cp .env.example .env
```


2. Build
```bash
go build cmd/main.go
```
3. Migrations
create data folder
```bash
mkdir data
```

```bash
~/go/bin/goose -dir migrations sqlite3 ./data/expenses.db up
```

4. Create user
```bash
go run ./cmd/useradm create -login bob -name "Bob"
```

5. Run app
```bash
./main
```
or in docker

```bash
docker compose up
```

6. Dbgate

```bash
docker compose -f docker-compose.dbgate.yml up
```

7. Dump database - just copy data/expenses.db 


