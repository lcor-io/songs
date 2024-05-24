package middlewares

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"

	"lcor.io/songs/src/services"
)

// Store sessions in memory for now
var store = session.New(session.Config{
	Expiration: time.Hour * 24 * 30,
	KeyLookup:  "cookie:songs_session",
})

func SessionMiddleware(c fiber.Ctx) error {
	session, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error getting session")
	}

	// Assign or retrieve an userId for the current session
	userId := session.Get("user_id")

	// TODO: Until proper authentication, users are automatically created while
	// navigating the site
	if userId == nil || !services.UserExists(userId.(string)) {
		nbUsers := len(services.GetUsers())
		newUser := services.CreateUser("Anonymous-" + fmt.Sprintf("%d", nbUsers+1))
		session.Set("user_id", newUser.ID)
	}

	if err := session.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error setting session")
	}

	// Save the userId in locals to pass it to other middlewares
	fiber.Locals(c, "session", userId)
	return c.Next()
}
