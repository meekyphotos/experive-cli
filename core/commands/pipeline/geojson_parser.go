package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/commands/json"
	"github.com/meekyphotos/experive-cli/core/utils"
	"github.com/valyala/fastjson"
	"sync"
)

func GeojsonParser(channel chan []byte, config *utils.Config, pool *fastjson.ParserPool) (chan map[string]interface{}, *sync.WaitGroup) {
	out := make(chan map[string]interface{}, 10000)
	var jsonWorkers sync.WaitGroup
	for i := 0; i < config.WorkerCount; i++ {
		jsonWorkers.Add(1)
		go func(index int) {
			actualWorker(channel, out, pool)
			jsonWorkers.Done()
		}(i)
	}
	return out, &jsonWorkers
}

func actualWorker(channel chan []byte, out chan map[string]interface{}, pool *fastjson.ParserPool) {
	for {
		select {
		case content := <-channel:
			if len(content) == 0 {
				return
			}
			parser := pool.Get()
			p, err := parser.ParseBytes(content)
			if err != nil {
				panic(err)
			}
			req := json.ParseJson(p)
			out <- req
			pool.Put(parser)
		default:
		}
	}
}
