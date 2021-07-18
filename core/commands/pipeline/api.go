package pipeline

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"time"
)

type Progress struct {
	Counter *progressbar.ProgressBar
}

type Heartbeat interface {
	Start()
	Beat(amount int)
	Done()
}

type NoopBeat struct{}

func (n *NoopBeat) Start()          {}
func (n *NoopBeat) Beat(amount int) {}
func (n *NoopBeat) Done()           {}

type BufferedBeat struct {
	OperationName   string
	Amount          int
	lastBeat        int64
	amountProcessed int
}

func (b *BufferedBeat) Start() {
	fmt.Printf("[%s] Started", b.OperationName)
}
func (b *BufferedBeat) Beat(amount int) {
	b.amountProcessed += amount
	if b.amountProcessed%b.Amount == 0 {
		millis := time.Now().UnixNano() / int64(time.Millisecond)
		if b.lastBeat-millis > 1000 {
			// avoid writing log too frequently
			fmt.Printf("[%s] Progress %d", b.OperationName, b.amountProcessed)
			b.lastBeat = millis
		}
	}

}
func (b *BufferedBeat) Done() {
	fmt.Printf("[%s] Done. Processed %d items", b.OperationName, b.amountProcessed)
}

type ProgressBarBeat struct {
	OperationName string
	bar           *progressbar.ProgressBar
}

func (b *ProgressBarBeat) Start() {
	b.bar = progressbar.NewOptions64(-1, progressbar.OptionSetDescription(b.OperationName))

}
func (b *ProgressBarBeat) Beat(amount int) {
	err := b.bar.Add(amount)
	if err != nil {
		panic(err)
	}
}
func (b *ProgressBarBeat) Done() {
	err := b.bar.Finish()
	if err != nil {
		panic(err)
	}
}
