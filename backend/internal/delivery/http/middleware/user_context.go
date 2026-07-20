package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	domainrepo "github.com/abdullahPrasetio/waphafiz/internal/domain/repository"
	"github.com/abdullahPrasetio/waphafiz/pkg/auth"
)

const (
	LocalUserID        = "user_id"
	LocalFamilyGroupID = "family_group_id"
	LocalUserRole      = "user_role"
)

// UserContext loads the authenticated user from DB and stores their
// family_group_id and role in Fiber locals for downstream handlers.
func UserContext(userRepo domainrepo.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := auth.GetClaims(c)
		if claims == nil {
			return c.Next()
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return c.Next()
		}

		user, err := userRepo.FindByID(c.UserContext(), userID)
		if err != nil {
			return c.Next()
		}

		c.Locals(LocalUserID, user.ID)
		c.Locals(LocalUserRole, string(user.Role))
		if user.FamilyGroupID != nil {
			c.Locals(LocalFamilyGroupID, *user.FamilyGroupID)
		}

		return c.Next()
	}
}

// GetUserID extracts the user UUID from Fiber locals set by UserContext middleware.
func GetUserID(c *fiber.Ctx) uuid.UUID {
	v, _ := c.Locals(LocalUserID).(uuid.UUID)
	return v
}

// GetFamilyGroupID extracts the family group UUID from Fiber locals.
func GetFamilyGroupID(c *fiber.Ctx) uuid.UUID {
	v, _ := c.Locals(LocalFamilyGroupID).(uuid.UUID)
	return v
}

// GetUserRole extracts the user role string from Fiber locals.
func GetUserRole(c *fiber.Ctx) string {
	v, _ := c.Locals(LocalUserRole).(string)
	return v
}
