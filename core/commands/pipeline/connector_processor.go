package pipeline

import "github.com/meekyphotos/experive-cli/core/commands/connectors"

func ProcessChannel(channel chan []map[string]interface{}, db connectors.Connector) error {
	for {
		select {
		case content := <-channel:
			if len(content) == 0 {
				return nil
			}
			err := db.Write(content)
			if err != nil {
				return err
			}
		default:
		}
	}
}
