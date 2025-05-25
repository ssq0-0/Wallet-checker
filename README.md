# Wallet Checker 

[English](#english) | [Русский](#russian)

<a name="russian"></a>
## Русский

### Описание
Chief Checker - это мощный инструмент, разработанный для проверки и анализа различных аспектов веб-сервисов и сетей. Построенный на Go, он обеспечивает эффективную и надежную производительность.

### Возможности
- Debank парсер. Собирает всю информацию о кошельке и записывает в файл в формате (общий баланс - баласн по каждой используемой сети - балансы в пулах ликвидности), в конце подводит общую статистику по общему балансу в usd и в каждом токене.
- Rabby парсер. Делает все тоже самое, что и Debank.

### Установка
```bash
# Клонировать репозиторий
git clone https://github.com/ssq0-0/Wallet-checker.git

# Перейти в директорию проекта
cd /Wallet-checker

# Настроить конфигурацию(см. ниже)
- При использовании ротационных прокси поставить true возле rotate_proxy. В таком случае в параметрах use_proxy_pool - false
- При использоании статичных прокси - false - rotate_proxy, use_proxy_pool - true. 

**Прокси можно вставлять в файл в любом из форматов**:
- `user:pass@host:port`
- `user:pass:host:port`
- `host:port@user:pass`
- `host:port:user:pass`
- `http://user:pass@host:port`
- `http://user:pass:host:port`
- `http://host:port@user:pass`
- `http://host:port:user:pass`
- `https://user:pass@host:port`
- `https://user:pass:host:port`
- `https://host:port@user:pass`
- `https://host:port:user:pass`
- `socks5://user:pass@host:port`
- `socks5://user:pass@host:port`
- `socks5://host:port@user:pass`
- `socks5://host:port:user:pass`

# Собрать проект
go build -o Wallet-checker cmd/main.go
```

### Запуск приложения
```bash
./Wallet-checker
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
└── go.sum         # Зависимости Go-модуля
```

### Конфигурация
Конфигурация приложения находится в файле `internal/config/appConfig/config.json`. Вот подробное описание параметров:

#### Основные параметры
- `concurrency` (int): Количество конкурентных функций для обработки запросов(**грубо** говоря - потоков)
- `logger_level` (string): Уровень логирования (debug, info, warn, error)

#### Параметры чекеров
##### DeBank
- `base_url` (string): Базовый URL API
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
**EVM** 0xb5017C6CD09e55fd8461ed10b6b03Da67798e99d

**BTC** bc1qyaspnva86536ccy4vz92vp646m4u7u0drfjxhq

**SOL** 7SXTWNzqKyewN2LtoWTDpppMmXNysRrxrPFtosnnEGmx

**TRC20** TBVbyuoC1xuuQkKi62b8hgnj4UX5YwNSjQ

[NodeMaven Proxy](https://nodemaven.com/?ref_id=1933c54f)
---

<a name="english"></a>
## English

### Description
Chief Checker is a powerful tool designed for checking and analyzing various aspects of web services and networks. Built with Go, it provides efficient and reliable performance for network operations.

### Features
- DeBank parser. Collects all wallet information and writes it to a file in a format (total balance - balance by each used network - balances in liquidity pools), and provides overall statistics on total balance in USD and in each token.
- Rabby parser. Identical to Debank module

### Installation
```bash
# Clone the repository
git clone https://github.com/ssq0-0/Wallet-checker.git

# Navigate to the project directory
cd Wallet-checker

# Configure the application (see below)
- For rotating proxies, set `rotate_proxy` to true and `use_proxy_pool` to false.
- For static proxies, set `rotate_proxy` to false and `use_proxy_pool` to true.

**You can list proxies in the file in any of the following formats**:
- `user:pass@host:port`
- `user:pass:host:port`
- `host:port@user:pass`
- `host:port:user:pass`
- `http://user:pass@host:port`
- `http://user:pass:host:port`
- `http://host:port@user:pass`
- `http://host:port:user:pass`
- `https://user:pass@host:port`
- `https://user:pass:host:port`
- `https://host:port@user:pass`
- `https://host:port:user:pass`
- `socks5://user:pass@host:port`
- `socks5://user:pass:host:port`
- `socks5://host:port@user:pass`
- `socks5://host:port:user:pass`

# Build the project
go build -o Wallet-checker cmd/main.go
```

### Running the Application
```bash
./Wallet-checker
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
├── pkg/            # Public library code
├── go.mod          # Go module definition
└── go.sum          # Go module checksums
```

### Configuration
Configuration of the application is located in the file `internal/config/appConfig/config.json`. Here is the detailed description of the parameters:

#### Main Parameters
- `concurrency` (int): Number of concurrent functions for processing requests (**grossly** speaking - threads)
- `logger_level` (string): Logging level (debug, info, warn, error)

#### Checker Parameters
##### DeBank
- `base_url` (string): Base URL of the API
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
**EVM** 0xb5017C6CD09e55fd8461ed10b6b03Da67798e99d

**BTC** bc1qyaspnva86536ccy4vz92vp646m4u7u0drfjxhq

**SOL** 7SXTWNzqKyewN2LtoWTDpppMmXNysRrxrPFtosnnEGmx

**TRC20** TBVbyuoC1xuuQkKi62b8hgnj4UX5YwNSjQ

[NodeMaven Proxy](https://nodemaven.com/?ref_id=1933c54f)

---
