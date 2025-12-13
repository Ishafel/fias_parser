# Парсер ФИАС

Утилита командной строки для потоковой передачи данных GAR XML в JSONL.

## Сборка образа Docker
```
docker build -t fias-parser .
```

## Запуск из Docker
Подключите каталог с XML-файлами и укажите на него в CLI. Схемы уже встроены в образ по пути `/gar_schemas`.
```
docker run --rm \
  -v "$(pwd)/test:/data" \
  fias-parser \
  --xml /data \
  --schema-dir /gar_schemas
```

Приведенная выше команда передает каждый XML-файл из хост-каталога `test` в stdout. Перенаправьте stdout, чтобы сохранить записи JSONL, например:
```
docker run --rm \
  -v "$(pwd)/test:/data" \
  fias-parser \
  --xml /data \
  --schema-dir /gar_schemas > output.jsonl
```

If your XML uses a specific child element under the root, add `--element <NAME>`.

### Валидация количества записей
Парсер теперь определяет ожидаемое количество записей из каждого XML-файла. Если какую-либо запись не удается обработать, пропущенная запись идентифицируется, и все несоответствия добавляются в журнал предупреждений (по умолчанию `validation.log`).
```
docker run --rm \
  -v "$(pwd)/test:/data" \
  fias-parser \
  --xml /data \
  --schema-dir /gar_schemas \
  --warn-log /data/validation.log
```
Приведенная выше команда передает данные в stdout, но записывает предупреждения, такие как отсутствующее количество записей или конкретные пропущенные записи (с байтовым смещением), в `/data/validation.log`.
