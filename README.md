# Worker Pool на Go с динамическим управлением

Данная работа реализует примитивный worker-pool на языке Go с возможностью динамически добавлять и удалять воркеров, которые обрабатывают строки данных.

## Описание работы с файлами

Эта программа считывает строки данных из текстового файла и записывает информацию о процессе обработки в лог-файл. Названия обоих файлов задаются через файл конфигурации `.env` для удобной настройки и изменения.

### Структура файлов

1. **Файл данных (`data.txt`)**: 
   - Этот текстовый файл содержит строки данных, которые обрабатываются воркерами.
   - Каждая строка передается в `JobQueue` пула воркеров, и каждый воркер забирает строку из очереди, обрабатывая её (например, выводя номер воркера и содержание строки).
   - Имя файла данных задается в `.env` файле с помощью переменной `DATA_FILE`.

2. **Лог-файл (`worker.log`)**:
   - Лог-файл используется для записи информации о процессе работы воркеров, таких как запуск, обработка строк и завершение работы каждого воркера.
   - Каждый воркер записывает в лог информацию о своих действиях, включая время, номер воркера и обрабатываемую строку, что помогает отслеживать статус и ход выполнения программы.
   - Имя лог-файла задается в `.env` файле с помощью переменной `LOG_FILE`.

## Возможности

- **Worker Pool**: Позволяет запускать и останавливать несколько воркеров, каждый из которых получает строки данных из общего канала и обрабатывает их.
- **Динамическое управление воркерами**: Воркеры могут добавляться и удаляться во время работы программы, что позволяет гибко управлять количеством обработчиков.
- **Конфигурация через файл .env**: Параметры для входного файла данных и файла логирования указываются в `.env` файле, что упрощает настройку без изменения кода.
- **Логирование**: Логи обрабатываемых данных и состояния воркеров записываются в указанный лог-файл.

## Структура

- **Worker**: структура, представляющая воркера. Каждый воркер получает данные из канала, обрабатывает их и логирует результат.
- **Pool**: структура, управляющая всеми воркерами. Реализует методы для добавления, удаления и остановки воркеров.
- **Основной цикл**: Поддерживает консольные команды `add`, `remove`, `exit` для управления воркерами.

## Установка и запуск

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/titoffon/worker-pool-go.git
   cd worker-pool-go```
   
2. Установите зависимости:

   ```bash
   go get -u github.com/joho/godotenv
3. Создайте файл .env с настройками:
   ```bash
   DATA_FILE=data.txt
   LOG_FILE=worker.logg

4. Запустите программу
   ```bash
   go run main.go

## Использование

Программа поддерживает следующие команды в консоли:

- **add**: добавляет нового воркера.
- **remove** `<id>`: удаляет воркера с указанным `ID`.
- **exit**: завершает работу программы, останавливая всех воркеров.

## Требования

- Go 1.16+
- Библиотека `godotenv` для работы с `.env` файлами

## Примечания

- Программа автоматически завершит работу с воркерами после прочтения всех данных из файла.
- Если в `.env` файле не указаны `DATA_FILE` или `LOG_FILE`, программа установит значения по умолчанию: `data.txt` и `worker.log`.



   
