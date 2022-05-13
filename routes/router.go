package routes

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func remove(s []Client, i int) []Client {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

type ClientQuery struct {
	UID string `json:"uid"`
}

type Client struct {
	Conn *websocket.Conn
	UID  string
}

type Hub struct {
	sync.Mutex
	c map[string][]Client
}

func InitRouter() {
	clients := Hub{
		c: make(map[string][]Client),
	}
	app := fiber.New()

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/api/v2/chat/:id", websocket.New(func(c *websocket.Conn) {
		id := c.Params("id")

		if id == "" {
			return
		}

		uid := c.Query("uid")

		if uid == "" {
			return
		}

		clients.c[id] = append(clients.c[id], Client{
			Conn: c,
			UID:  uid,
		})

		// // websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				for i, client := range clients.c[id] {
					if client.UID == uid {
						clients.c[id] = remove(clients.c[id], i)
						break
					}
				}
				break
			}

			for i, client := range clients.c[id] {
				if err = client.Conn.WriteMessage(mt, msg); err != nil {
					log.Println("write:", err)
					clients.c[id] = remove(clients.c[id], i)
					continue
				}
			}
		}

	}))

	log.Fatal(app.Listen(":3004"))
}
