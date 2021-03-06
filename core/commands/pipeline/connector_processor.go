package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/commands/connectors"
	"time"
)

func ProcessChannel(channel chan []map[string]interface{}, nodes connectors.Connector, beat Heartbeat) error {
	beat.Start()
	defer beat.Done()
	for {
		select {
		case content := <-channel:
			i := len(content)
			if i == 0 {
				return nil
			}
			err := nodes.Write(content)
			beat.Beat(i)
			if err != nil {
				return err
			}
		default:
		}
	}

}

func BatchRequest(values <-chan map[string]interface{}, maxItems int, maxTimeout time.Duration) chan []map[string]interface{} {
	batches := make(chan []map[string]interface{})

	go func() {
		defer close(batches)

		for keepGoing := true; keepGoing; {
			var batch []map[string]interface{}
			expire := time.After(maxTimeout)
			for {
				select {
				case value, ok := <-values:
					if !ok {
						keepGoing = false
						goto done
					}

					batch = append(batch, value)
					if len(batch) == maxItems {
						goto done
					}

				case <-expire:
					goto done
				}
			}

		done:
			if len(batch) > 0 {
				batches <- batch
			}
		}
	}()

	return batches
}

func BatchRequests(values <-chan []interface{}, size int, maxTimeout time.Duration) chan [][]interface{} {
	batches := make(chan [][]interface{})

	go func() {
		defer close(batches)

		for keepGoing := true; keepGoing; {
			batch := make([][]interface{}, 0, size)
			expire := time.After(maxTimeout)
			for {
				select {
				case value, ok := <-values:
					if !ok {
						keepGoing = false
						goto done
					}
					batch = append(batch, value)
					if len(batch) == size {
						goto done
					}

				case <-expire:
					goto done
				}
			}

		done:
			if len(batch) > 0 {
				println("Sending batch to postgres")
				batches <- batch
			}
		}
	}()

	return batches
}
