package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Worker структура
type Worker struct {
    ID       int
    JobQueue chan string
    Quit     chan bool
}


// NewWorker создает новый воркер
func NewWorker(id int, jobQueue chan string) *Worker {
    return &Worker{
        ID:       id,
        JobQueue: jobQueue,
        Quit:     make(chan bool, 1), // Буферизованный канал
    }
}

// Start запускает воркера для обработки данных
func (w *Worker) Start(wg *sync.WaitGroup, logger *log.Logger) {
    defer wg.Done()
    for {
        select {
        case job, ok := <-w.JobQueue:
            if !ok {
                logger.Printf("Воркер %d: канал закрыт, завершаю работу\n", w.ID)
                return
            }
            logger.Printf("Воркер %d: обработал строку: %s\n", w.ID, job)
			time.Sleep(time.Second)
        case <-w.Quit:
			logger.Printf("Воркер %d: отправлен сигнал на остановку и удаление\n", w.ID)
            logger.Printf("Воркер %d: удалён\n", w.ID)
            return
        }
    }
}

// Stop останавливает воркера
func (w *Worker) Stop() {
    w.Quit <- true
}

// Pool структура для управления воркерами
type Pool struct {
    JobQueue  chan string
    Workers   map[int]*Worker
    SyncGroup sync.WaitGroup
    Logger    *log.Logger
    nextWorkerID int // Счётчик для ID воркеров
}

// NewPool создает новый pool
func NewPool(logger *log.Logger) *Pool {
    return &Pool{
        JobQueue: make(chan string),
        Workers:  make(map[int]*Worker),
        Logger:   logger,
    }
}

// AddWorker добавляет нового воркера в pool
func (p *Pool) AddWorker() {
    newID := len(p.Workers) + 1
    worker := NewWorker(newID, p.JobQueue)
    p.Workers[p.nextWorkerID] = worker
    p.nextWorkerID++
    p.SyncGroup.Add(1)
    go worker.Start(&p.SyncGroup, p.Logger)
    fmt.Printf("Воркер %d запущен\n", newID)
}

// RemoveWorker удаляет воркера из pool по ID
func (p *Pool) RemoveWorker(id int) {
    worker, exists := p.Workers[id]
    if !exists {
        fmt.Printf("Воркер с ID %d не найден\n", id)
        return
    }
    worker.Stop()
    delete(p.Workers, id)
    fmt.Printf("Воркер %d удален\n", id)
}

// Stop завершает работу всех воркеров и закрывает pool
func (p *Pool) Stop() {
    for id := range p.Workers {
        //worker.Stop()
		delete(p.Workers, id)
        p.Logger.Printf("Воркер %d: отправлен сигнал на остановку и удаление\n", id)
    }

    p.SyncGroup.Wait()
}

func main() {
    err := godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки .env файла")
		return
	}

    // Получаем имена файлов из переменных окружения
	dataFile := os.Getenv("DATA_FILE")
	loggerFile := os.Getenv("LOG_FILE")

    if dataFile == "" || loggerFile == "" {
		fmt.Println("Не заданы DATA_FILE или LOG_FILE в .env файле")
        fmt.Println("Установлены дефолтные значения")
        dataFile = "data.txt"
        loggerFile = "worker.log"
		return
	}

    // Открываем лог-файл
    logFile, err := os.OpenFile(loggerFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        fmt.Println("Ошибка при открытии файла логов:", err)
        return
    }
    defer logFile.Close()

    logger := log.New(logFile, "", log.LstdFlags)

    pool := NewPool(logger)

    // Канал для сигнализации о завершении
    done := make(chan struct{})

	exit := make(chan struct{})

    finished := make(chan struct{}) //Канал для завершения обработки файла

    // Запускаем обработку команд пользователя
    go func() {
        scanner := bufio.NewScanner(os.Stdin)
        fmt.Println("Команды:")
        fmt.Println("add    - добавить воркер")
        fmt.Println("remove - удалить воркер по ID")
        fmt.Println("exit   - выйти из программы")

        for {
            fmt.Print("> ")
            if !scanner.Scan() {
                break
            }
            select {
                case <-finished: // Выход, если файл уже обработан
                    return
			    default:
                    command := scanner.Text()
                    switch command {
                    case "add":
                        pool.AddWorker()
                    case "remove":
                        if len(pool.Workers) == 0 {
                            fmt.Println("Нет запущенных воркеров для удаления")
                            continue
                        }
                        fmt.Print("Введите ID воркера для удаления: ")
                        if !scanner.Scan() {
                            break
                        }
                        idStr := scanner.Text()
                        id, err := strconv.Atoi(idStr)
                        if err != nil {
                            fmt.Println("Некорректный ID воркера")
                            continue
                        }
                        pool.RemoveWorker(id)

                    case "exit":
                        close(exit)
                        fmt.Println("Завершение работы...")
                        pool.Stop()
                        close(done) // Сигнализируем о завершении
                        return
                    default:
                        fmt.Println("Неизвестная команда")
                    }
            }
        }
    }()

    // Чтение данных из файла и отправка их в канал
	go func() {
		file, err := os.Open(dataFile)
		if err != nil {
			fmt.Println("Ошибка при открытии файла данных:", err)
			pool.Stop()
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case pool.JobQueue <- line:
			case <-exit: //если завершение до конца файла
				close(pool.JobQueue)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Ошибка при чтении файла данных:", err)
		}

		// Закрываем каналы и завершаем работу
        
		close(pool.JobQueue)
		close(exit)
        close(finished) // Сигнализируем о завершении работы с файлом
		fmt.Println("Чтение завершено. Завершение работы...")
		pool.Stop()
		close(done)
	}()

    // Ожидание сигнала завершения
    <-done
}