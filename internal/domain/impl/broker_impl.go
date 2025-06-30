package broker_impl

import (
	"sync"
	"time"

	"github.com/AnatoliySar/test/internal/domain"
)

type Queue struct {
	messages chan domain.Message
}

type Broker struct {
	queuesMux  sync.RWMutex
	queues     map[string]*Queue
	maxQueues  int
	maxMsgs    int
	waiters    map[string][]chan domain.Message
	waitersMux sync.Mutex
}

func NewBroker(maxQueues, maxMsgs int) *Broker {
	return &Broker{
		queues:    make(map[string]*Queue),
		maxQueues: maxQueues,
		maxMsgs:   maxMsgs,
		waiters:   make(map[string][]chan domain.Message),
	}
}

func (b *Broker) Put(queueName, msg string) error {
	b.queuesMux.RLock()
	queue, exists := b.queues[queueName]
	b.queuesMux.RUnlock()

	// Если очередь не существует, создаем новую
	if !exists {
		b.queuesMux.Lock()
		defer b.queuesMux.Unlock()

		if len(b.queues) >= b.maxQueues {
			return domain.ErrQueueLimit
		}

		// Дополнительная проверка на случай параллельного создания очереди
		if queue, exists = b.queues[queueName]; !exists {
			queue = &Queue{
				messages: make(chan domain.Message, b.maxMsgs),
			}

			b.queues[queueName] = queue
		}
	}

	// Проверяем, есть ли ожидающие получатели
	b.waitersMux.Lock()

	if waiters, ok := b.waiters[queueName]; ok && len(waiters) > 0 {
		// Отправляем сообщение первому ожидающему
		waiter := waiters[0]
		b.waiters[queueName] = waiters[1:]
		b.waitersMux.Unlock()

		waiter <- domain.Message{Data: msg, Timestamp: time.Now()}
		return nil
	}

	b.waitersMux.Unlock()

	// Если нет ожидающих, добавляем в очередь
	select {
	case queue.messages <- domain.Message{Data: msg, Timestamp: time.Now()}:
		return nil
	default:
		return domain.ErrQueueFull
	}
}

// Get получает сообщение из очереди с возможностью ожидания
func (b *Broker) Get(queueName string, timeout time.Duration) (domain.Message, error) {
	b.queuesMux.RLock()
	queue, exists := b.queues[queueName]
	b.queuesMux.RUnlock()

	if !exists {
		queue = &Queue{
			messages: make(chan domain.Message, b.maxMsgs),
		}
		b.queuesMux.Lock()
		b.queues[queueName] = queue
		b.queuesMux.Unlock()
	}

	select {
	case msg := <-queue.messages:
		return msg, nil
	default:
		// Если сообщения нет и timeout = 0, возвращаем ошибку сразу
		if timeout == 0 {
			return domain.Message{}, domain.ErrMessageNotFound
		}

		// Создаем канал для ожидания сообщения
		msgChan := make(chan domain.Message, 1)
		b.waitersMux.Lock()
		b.waiters[queueName] = append(b.waiters[queueName], msgChan)
		b.waitersMux.Unlock()

		select {
		case msg := <-msgChan:
			return msg, nil
		case <-time.After(timeout):
			// Удаляем наш канал из ожидающих
			b.waitersMux.Lock()
			defer b.waitersMux.Unlock()
			for i, ch := range b.waiters[queueName] {
				if ch == msgChan {
					b.waiters[queueName] = append(b.waiters[queueName][:i], b.waiters[queueName][i+1:]...)
					break
				}
			}
			return domain.Message{}, domain.ErrMessageNotFound
		}
	}
}
