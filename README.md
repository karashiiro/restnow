# restnow
Sample Fiber API generator.

Drop an input file on the binary, and get a project folder in the same directory.

Input file example:
```json
{
    "name": "sample",
    "repoName": "example.com/sample",
    "defaultPort": 7552,
    "routes": {
        "something": {
            "1": {
                "a": ["GET", "POST", "PUT"]
            },
            "2": ["GET"],
            "3": ["GET", "DELETE"]
        }
    }
}
```

Output file:
```
package main

import (
	"flag"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func main() {
	port := flag.Uint("port", 7552, "server binding port")
	flag.Parse()

	app := fiber.New()

	app.Get("/something/1/a", func(ctx *fiber.Ctx) error {
		return nil
	})

	app.Post("/something/1/a", func(ctx *fiber.Ctx) error {
		return nil
	})

	app.Put("/something/1/a", func(ctx *fiber.Ctx) error {
		return nil
	})

	app.Get("/something/2", func(ctx *fiber.Ctx) error {
		return nil
	})

	app.Get("/something/3", func(ctx *fiber.Ctx) error {
		return nil
	})

	app.Delete("/something/3", func(ctx *fiber.Ctx) error {
		return nil
	})

	app.Listen(":" + strconv.Itoa(int(*port)))
}
```
