package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/commands/osm"
	"github.com/meekyphotos/experive-cli/core/dataproviders"
	"github.com/valyala/fastjson"
	"strings"
)

func ProcessNodeEnrichment(wayChannel <-chan *osm.Way, store dataproviders.Store, beat Heartbeat) {
	beat.Start()
	defer beat.Done()
	arena := fastjson.Arena{}
	for {
		select {
		case content := <-wayChannel:
			if content == nil {
				return
			}
			arena.Reset()
			if len(content.NodeIds) > 0 {
				nodes := store.FindMany(content.NodeIds...)
				if len(nodes) > 0 {
					for id, n := range nodes {
						var address, extratags *fastjson.Value
						if n.Exists("address") {
							address = n.Get("address")
						} else {
							address = arena.NewObject()

						}

						if n.Exists("extratags") {
							extratags = n.Get("extratags")
						} else {
							extratags = arena.NewObject()
						}

						delete(content.Tags, "source")
						if len(content.Tags) > 0 {
							for k, v := range content.Tags {
								if strings.HasPrefix(k, "name") {
									lang := k[4:]
									if lang == "" || lang == ":it" || lang == ":en" || lang == ":nl" {
										address.Set("addr:name"+lang, arena.NewString(v))
									}
								} else if strings.HasPrefix(k, "addr:") {
									address.Set(k, arena.NewString(v))
								} else if "admin_level" == k {
									address.Set("addr:admin_level", arena.NewString(v))
								} else {
									extratags.Set("way:"+k, arena.NewString(v))
								}
							}

							n.Set("address", address)
							n.Set("extratags", extratags)
							store.Save(&dataproviders.INode{
								Id:      id,
								Content: n.MarshalTo([]byte{}),
							})
						}
					}
				}
			}
			beat.Beat(1)
		default:
		}
	}

}
