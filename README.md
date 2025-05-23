# Chief Checker | Чиф Чекер

[English](#english) | [Русский](#russian)

<a name="english"></a>
## English

### Description
Chief Checker is a powerful tool designed for checking and analyzing various aspects of web services and networks. Built with Go, it provides efficient and reliable performance for network operations.

### Features
- High-performance network operations
- Configurable logging system
- Modular architecture
- Cross-platform compatibility

### Installation
```bash
# Clone the repository
git clone https://github.com/ssq0-0/chief-checker.git

# Navigate to the project directory
cd chief-checker

# Build the project
go build -o chief-checker cmd/main.go
```

### Running the Application
```bash
./chief-checker
```

### Technical Details
- Built with Go 1.24
- Uses modern Go modules for dependency management
- Key dependencies:
  - github.com/refraction-networking/utls
  - github.com/sirupsen/logrus
  - github.com/stretchr/testify
  - golang.org/x/net

### Project Structure
```
.
├── cmd/            # Application entry points
├── internal/       # Private application code
├── pkg/           # Public library code
├── go.mod         # Go module definition
└── go.sum         # Go module checksums
```

### Configuration
Configuration of the application is located in the file `internal/config/appConfig/config.json`. Here is the detailed description of the parameters:

#### Main Parameters
- `concurrency` (int): Number of concurrency functions for processing requests (**grossly** speaking - threads)
- `logger_level` (string): Logging level (debug, info, warn, error)

#### Checker Parameters
##### DeBank
- `base_url` (string): Base URL of the DeBank API
- `endpoints` (object): Configuration of the API endpoints
  - `user_info`: Getting user information
  - `used_chains`: Getting the list of used blockchains
  - `token_balance_list`: Getting the list of token balances
  - `project_list`: Getting the list of projects in the portfolio
- `rotate_proxy` (boolean): Rotation of proxies
- `use_proxy_pool` (boolean): Using a proxy pool
- `reuse_proxy` (boolean): Reusing proxies
- `proxy_file_path` (string): Path to the proxy file
- `deadline_request` (int): Request timeout in seconds

### Donations
**EVM** 
**BTC**
**SOL**
**TRC20**

---

<a name="russian"></a>
## Русский

### Описание
Chief Checker - это мощный инструмент, разработанный для проверки и анализа различных аспектов веб-сервисов и сетей. Построенный на Go, он обеспечивает эффективную и надежную производительность для сетевых операций.

### Возможности
- Высокопроизводительные сетевые операции
- Настраиваемая система логирования
- Модульная архитектура
- Кросс-платформенная совместимость

### Установка
```bash
# Клонировать репозиторий
git clone https://github.com/ssq0-0/chief-checker.git

# Перейти в директорию проекта
cd chief-checker

# Настроить конфигурацию(см. ниже)
- При использовании ротационных прокси поставить true возле rotate_proxy. В таком случае в параметрах use_proxy_pool, reuse_proxy - false
- При использоании статичных прокси - false - rotate_proxy, use_proxy_pool - true. Есть два варианта - использовать на каждый запрос уникальный прокси, установить true в параметр reuse_proxy, иначе - false(429 ошибка будет встречаться очень часто)
# Собрать проект
go build -o chief-checker cmd/main.go
```

### Запуск приложения
```bash
./chief-checker
```

### Технические детали
- Построен на Go 1.24
- Использует современные Go модули для управления зависимостями
- Основные зависимости:
  - github.com/refraction-networking/utls
  - github.com/sirupsen/logrus
  - github.com/stretchr/testify
  - golang.org/x/net

### Структура проекта
```
.
├── cmd/            # Точки входа приложения
├── internal/       # Приватный код приложения
├── pkg/           # Публичный код библиотеки
├── go.mod         # Определение Go модуля
└── go.sum         # Контрольные суммы Go модуля
```

### Конфигурация
Конфигурация приложения находится в файле `internal/config/appConfig/config.json`. Вот подробное описание параметров:

#### Основные параметры
- `concurrency` (int): Количество конкурентных функций для обработки запросов(**грубо** говоря - потоков)
- `logger_level` (string): Уровень логирования (debug, info, warn, error)

#### Параметры чекеров
##### DeBank
- `base_url` (string): Базовый URL API DeBank
- `endpoints` (object): Конфигурация эндпоинтов API
  - `user_info`: Получение информации о пользователе
  - `used_chains`: Получение списка используемых блокчейнов
  - `token_balance_list`: Получение списка балансов токенов
  - `project_list`: Получение списка проектов в портфолио
- `rotate_proxy` (boolean): Использование ротационных прокси
- `use_proxy_pool` (boolean): Использование пула прокси
- `reuse_proxy` (boolean): Повторное использование прокси
- `proxy_file_path` (string): Путь к файлу с прокси
- `deadline_request` (int): Таймаут запросов в секундах

### Донаты
**EVM** 
**BTC**
**SOL**
**TRC20**