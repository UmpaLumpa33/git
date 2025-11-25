package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"math/rand"
)

// Кастомные ошибки
var (
	ErrInsufficientFunds   = errors.New("недостаточно средств на счете")
	ErrInvalidAmount       = errors.New("сумма должна быть больше нуля")
	ErrAccountNotFound     = errors.New("счет не найден")
	ErrSameAccountTransfer = errors.New("нельзя переводить самому себе")
)

// Основной интерфейс для работы со счетами
type AccountService interface {
	Deposit(amount float64) error               // Пополнение
	Withdraw(amount float64) error              // Снятие
	Transfer(to *Account, amount float64) error // Перевод
	GetBalance() float64                        // Текущий баланс
	GetStatement() string                       // История операций
}

// Хранилище данных
type Storage interface {
	SaveAccount(account *Account) error             // Сохранить аккаунт
	LoadAccount(accountID string) (*Account, error) // Загрузить аккаунт
	GetAllAccounts() ([]*Account, error)            // Все существующие аккаунты
}

// Представление банковского счета
type Account struct {
	ID        string
	Name      string
	Balance   float64
	Statement []string
}

// Пополняем счет
func (acct *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	acct.Balance += amount
	acct.addTransaction(fmt.Sprintf("Пополнение: +%.2f руб.", amount))
	return nil
}

// Снимаем средства
func (acct *Account) Withdraw(amount float64) error {
	if amount > acct.Balance || amount <= 0 {
		return ErrInsufficientFunds
	}
	acct.Balance -= amount
	acct.addTransaction(fmt.Sprintf("Снятие: -%.2f руб.", amount))
	return nil
}

// Перевод на другой счет
func (acct *Account) Transfer(to *Account, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if acct.ID == to.ID {
		return ErrSameAccountTransfer
	}
	err := acct.Withdraw(amount)
	if err != nil {
		return err
	}
	to.Deposit(amount)
	acct.addTransaction(fmt.Sprintf("Перевод на %s: -%.2f руб.", to.Name, amount))
	to.addTransaction(fmt.Sprintf("Получено от %s: +%.2f руб.", acct.Name, amount))
	return nil
}

// Добавляет операцию в журнал транзакций
func (acct *Account) addTransaction(desc string) {
	acct.Statement = append(acct.Statement, desc)
}

// Баланс счета
func (acct *Account) GetBalance() float64 {
	return acct.Balance
}

// Список транзакций
func (acct *Account) GetStatement() string {
	return strings.Join(acct.Statement, "\n")
}

// Реализация хранилища в памяти
type InMemoryStorage struct {
	mu       sync.Mutex
	accounts map[string]*Account
}

// Новый экземпляр хранилища
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		accounts: make(map[string]*Account),
	}
}

// Сохранение аккаунта
func (ims *InMemoryStorage) SaveAccount(account *Account) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()
	ims.accounts[account.ID] = account
	return nil
}

// Поиск аккаунта по ID
func (ims *InMemoryStorage) LoadAccount(accountID string) (*Account, error) {
	ims.mu.Lock()
	defer ims.mu.Unlock()
	account, found := ims.accounts[accountID]
	if !found {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

// Получение всех существующих аккаунтов
func (ims *InMemoryStorage) GetAllAccounts() ([]*Account, error) {
	ims.mu.Lock()
	defer ims.mu.Unlock()
	var result []*Account
	for _, acc := range ims.accounts {
		result = append(result, acc)
	}
	return result, nil
}

// Главная логика программы
func main() {
	reader := bufio.NewReader(os.Stdin)
	storage := NewInMemoryStorage()

	for {
		fmt.Println("\nБанковская система:")
		fmt.Println("1. Создать счет")
		fmt.Println("2. Пополнить счет")
		fmt.Println("3. Снять средства")
		fmt.Println("4. Перевести другому счету")
		fmt.Println("5. Баланс")
		fmt.Println("6. История операций")
		fmt.Println("7. Выход")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			createAccount(reader, storage)
		case "2":
			deposit(reader, storage)
		case "3":
			withdraw(reader, storage)
		case "4":
			transfer(reader, storage)
		case "5":
			showBalance(reader, storage)
		case "6":
			showHistory(reader, storage)
		case "7":
			fmt.Println("До свидания!")
			os.Exit(0)
		default:
			fmt.Println("Некорректный выбор.")
		}
	}
}

// Функция создания нового счета
func createAccount(r *bufio.Reader, storage *InMemoryStorage) {
	fmt.Print("Введите свое имя: ")
	name, _ := r.ReadString('\n')
	name = strings.TrimSpace(name)

	// Генерируем случайный идентификатор счета
	accountID := randomID()

	// Создаем новый объект Account
	account := &Account{
		ID:        accountID,
		Name:      name,
		Balance:   0,
		Statement: make([]string, 0),
	}

	// Сохраняем созданный аккаунт
	storage.SaveAccount(account)
	fmt.Printf("Ваш счет №%s успешно создан!\n", accountID)
}

// Функция внесения суммы на счет
func deposit(r *bufio.Reader, storage *InMemoryStorage) {
	fmt.Print("Введите номер счета: ")
	accountID, _ := r.ReadString('\n')
	accountID = strings.TrimSpace(accountID)

	account, err := storage.LoadAccount(accountID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("Введите сумму пополнения: ")
	amountStr, _ := r.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := parseFloat64(amountStr)
	if err != nil {
		fmt.Println("Некорректная сумма.")
		return
	}

	err = account.Deposit(amount)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Пополнено: %.2f руб.\n", amount)
}

// Функция снятия средств
func withdraw(r *bufio.Reader, storage *InMemoryStorage) {
	fmt.Print("Введите номер счета: ")
	accountID, _ := r.ReadString('\n')
	accountID = strings.TrimSpace(accountID)

	account, err := storage.LoadAccount(accountID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("Введите сумму снятия: ")
	amountStr, _ := r.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := parseFloat64(amountStr)
	if err != nil {
		fmt.Println("Некорректная сумма.")
		return
	}

	err = account.Withdraw(amount)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Снято: %.2f руб.\n", amount)
}

// Функция перевода средств
func transfer(r *bufio.Reader, storage *InMemoryStorage) {
	fmt.Print("Введите номер Вашего счета: ")
	fromID, _ := r.ReadString('\n')
	fromID = strings.TrimSpace(fromID)

	fmt.Print("Введите номер счета получателя: ")
	toID, _ := r.ReadString('\n')
	toID = strings.TrimSpace(toID)

	fromAcct, err := storage.LoadAccount(fromID)
	if err != nil {
		fmt.Println(err)
		return
	}

	toAcct, err := storage.LoadAccount(toID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("Введите сумму перевода: ")
	amountStr, _ := r.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := parseFloat64(amountStr)
	if err != nil {
		fmt.Println("Некорректная сумма.")
		return
	}

	err = fromAcct.Transfer(toAcct, amount)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Перечислено: %.2f руб.\n", amount)
}

// Функция отображения баланса
func showBalance(r *bufio.Reader, storage *InMemoryStorage) {
	fmt.Print("Введите номер счета: ")
	accountID, _ := r.ReadString('\n')
	accountID = strings.TrimSpace(accountID)

	account, err := storage.LoadAccount(accountID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Текущий баланс: %.2f руб.\n", account.GetBalance())
}

// Функция просмотра истории операций
func showHistory(r *bufio.Reader, storage *InMemoryStorage) {
	fmt.Print("Введите номер счета: ")
	accountID, _ := r.ReadString('\n')
	accountID = strings.TrimSpace(accountID)

	account, err := storage.LoadAccount(accountID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("История операций:\n", account.GetStatement())
}

// Преобразование строки в число типа float64
func parseFloat64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// Простая генерация случайного идентификатора счета
func randomID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]rune, 8)
	for i := range b {
		b[i] = rune(letters[rand.Intn(len(letters))])
	}
	return "ACC-" + string(b)
}

// Инициализация генератора случайных чисел
func init() {
	rand.Seed(time.Now().UnixNano())
}
