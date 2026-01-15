package closer

import (
	"context"
)

// Closer определяет интерфейс для ресурсов, которые нужно закрыть
type Closer interface {
	Close(ctx context.Context) error
	Name() string
}
