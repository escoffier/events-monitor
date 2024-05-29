package main

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events/exchange"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/typeurl/v2"
)

func main() {
	containerdCli, err := containerd.New("/run/containerd/containerd.sock", containerd.WithTimeout(time.Second*5))
	if err != nil {
		panic(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = containerdCli.IsServing(ctx)
	if err != nil {
		panic(fmt.Sprintf("check containerd daemon is serving failed,%v", err))
	}
	version, err := containerdCli.Version(ctx)
	if err != nil {
		panic(fmt.Sprintf("get containerd version failed,%v", err))
	}
	fmt.Println(version)

	newExchange := exchange.NewExchange()

	filters := []string{
		fmt.Sprintf(`topic=="%s"`, runtime.TaskStartEventTopic),
		fmt.Sprintf(`topic=="%s"`, runtime.TaskPausedEventTopic),
		fmt.Sprintf(`topic=="%s"`, runtime.TaskResumedEventTopic),
		fmt.Sprintf(`topic=="%s"`, runtime.TaskDeleteEventTopic),
		fmt.Sprintf(`topic=="%s"`, runtime.TaskExitEventTopic),
	}

	subCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	msg, errs := containerdCli.Subscribe(subCtx, filters...)
	for {
		select {
		case m := <-msg:
			v, err := typeurl.UnmarshalAny(m.Event)
			if err != nil {
				fmt.Printf("failed to unmarshal event %v\n", err)
				// l.Get().Err(err).Any("event", m).Msg("failed to unmarshal event")
				return
			}
			var (
				containerId string
				pid         uint32
			)
			action := ""
			switch t := v.(type) {
			case *events.TaskStart:
				containerId = t.ContainerID
				pid = t.Pid
				action = "start"
				time.Sleep(time.Second) //todo
			case *events.TaskDelete:
				containerId = t.ContainerID
				pid = t.Pid
				action = "delete"
			case *events.TaskExit:
				containerId = t.ContainerID
				pid = t.Pid
				action = "exit"
			case *events.TaskPaused:
				containerId = t.ContainerID
				action = "pause"
			case *events.TaskResumed:
				containerId = t.ContainerID
				action = "resume"
			default:
				fmt.Printf("containerd ignore event, namespace:%s,topic:%s,event:%s\n", m.Namespace, m.Topic, m.Event.GetTypeUrl())
				return
			}
			fmt.Printf("containerd received event ,topic:%s,namespace:%s,containerId:%s, action:%s, pid:%d\n",
				m.Topic, m.Namespace, containerId, action, pid)

		case err := <-errs:
			if err != nil {
				fmt.Printf("err returned for containerd events. try to restart: %v\n", err)
				// try to restart listening to container streams
				msg, errs = newExchange.Subscribe(subCtx)
				time.Sleep(time.Second * 3)
			}
		}
	}

}
