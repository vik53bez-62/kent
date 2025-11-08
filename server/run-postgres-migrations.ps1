<#=====================================================================
  PostgreSQL Migration Runner (Render)
  • Подключается к внешнему хосту Render
  • Выполняет все *.sql‑файлы из папки migrations
  • Автоматически задаёт SSL (sslmode=require) и порт 5432
=====================================================================#>

# ------------------- Параметры подключения -------------------
$host   = "dpg-d47bcnripnbc73cob290-a.oregon-postgres.render.com"
$port   = 5432
$dbName = "kent_afcg"
$user   = "kent"
$pwd    = "Fq0cqDWQiuuLRPAlEyrhbbK5sHGtYNxK"

# ------------------- Проверка наличия psql -------------------
$psql = (Get-Command psql -ErrorAction SilentlyContinue).Source
if (-not $psql) {
    Write-Error "psql не найден в PATH. Установите PostgreSQL client (https://www.postgresql.org/download/windows/) и добавьте его в PATH."
    exit 1
}

# ------------------- Устанавливаем переменные окружения -------------------
$env:PGPASSWORD = $pwd
$env:PGSSLMODE  = "require"
$env:PGPORT     = $port   # (можно опустить, если PGPORT не установлен)

# ------------------- Папка миграций -------------------
$migrationsDir = Join-Path -Path $PSScriptRoot -ChildPath "migrations"
if (-not (Test-Path $migrationsDir)) {
    Write-Error "Папка миграций не найдена: $migrationsDir"
    exit 1
}

# ------------------- Список *.sql‑файлов -------------------
$migrationFiles = Get-ChildItem -Path $migrationsDir -Filter "*.sql" |
                  Sort-Object Name

if ($migrationFiles.Count -eq 0) {
    Write-Warning "В папке $migrationsDir нет файлов *.sql"
    exit 0
}

# ------------------- Выполняем миграции -------------------
foreach ($file in $migrationFiles) {
    Write-Host "`n=== Выполняем миграцию: $($file.Name) ===" -ForegroundColor Yellow

    & $psql -h $host -U $user -d $dbName -f $file.FullName

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Миграция $($file.Name) завершилась с ошибкой (код $LASTEXITCODE). Остановка."
        exit $LASTEXITCODE
    } else {
        Write-Host "Миграция $($file.Name) выполнена успешно." -ForegroundColor Green
    }
}

Write-Host "`nВсе миграции выполнены!" -ForegroundColor Magenta