package elemental

import (
	"net/url"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserResponse is the response to CreateUser
type CreateUserResponse struct {
	Success bool
	Data    string
}

// CreateUser creates a user
func (e *Elemental) CreateUser(name string, password string) CreateUserResponse {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return CreateUserResponse{
			Success: false,
			Data:    err.Error(),
		}
	}

	var uid string
	count := 1
	for count != 0 {
		uid, err = GenerateRandomStringURLSafe(16)
		if err != nil {
			return CreateUserResponse{
				Success: false,
				Data:    err.Error(),
			}
		}

		// Check if name taken
		res, err := e.db.Query("SELECT COUNT(1) FROM users WHERE uid=? LIMIT 1", name)
		if err != nil {
			return CreateUserResponse{
				Success: false,
				Data:    err.Error(),
			}
		}
		defer res.Close()
		res.Next()
		err = res.Scan(&count)
		if err != nil {
			return CreateUserResponse{
				Success: false,
				Data:    err.Error(),
			}
		}
	}

	// Check if name taken
	res, err := e.db.Query("SELECT COUNT(1) FROM users WHERE name=? LIMIT 1", name)
	if err != nil {
		return CreateUserResponse{
			Success: false,
			Data:    err.Error(),
		}
	}
	defer res.Close()
	res.Next()
	err = res.Scan(&count)
	if err != nil {
		return CreateUserResponse{
			Success: false,
			Data:    err.Error(),
		}
	}
	if count == 1 {
		return CreateUserResponse{
			Success: false,
			Data:    "Account already exists!",
		}
	}

	_, err = e.db.Exec("INSERT INTO users VALUES( ?, ?, ?, ? )", name, uid, string(hashedPassword), `["Air", "Earth", "Fire", "Water"]`)
	if err != nil {
		return CreateUserResponse{
			Success: false,
			Data:    err.Error(),
		}
	}

	return CreateUserResponse{
		Success: true,
		Data:    uid,
	}
}

func (e *Elemental) createUser(c *fiber.Ctx) error {
	name, err := url.PathUnescape(c.Params("name"))
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    err.Error(),
		})
	}
	password, err := url.PathUnescape(c.Params("password"))
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    err.Error(),
		})
	}

	resp := e.CreateUser(name, password)
	return c.JSON(map[string]interface{}{
		"success": resp.Success,
		"data":    resp.Data,
	})
}

func (e *Elemental) loginUser(c *fiber.Ctx) error {
	name, err := url.PathUnescape(c.Params("name"))
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    err.Error(),
		})
	}
	password, err := url.PathUnescape(c.Params("password"))
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    err.Error(),
		})
	}

	// Check if user exists
	res, err := e.db.Query("SELECT COUNT(1) FROM users WHERE name=?", name)
	if err != nil {
		return err
	}
	defer res.Close()
	var count int
	res.Next()
	err = res.Scan(&count)
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    err.Error(),
		})
	}
	if count == 0 {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    "Invalid username",
		})
	}

	res, err = e.db.Query("SELECT uid, password FROM users WHERE name=? LIMIT 1", name)
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    err.Error(),
		})
	}
	defer res.Close()
	var uid string
	var pwd string
	res.Next()
	err = res.Scan(&uid, &pwd)
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    err.Error(),
		})
	}

	err = bcrypt.CompareHashAndPassword([]byte(pwd), []byte(password))
	if err != nil {
		return c.JSON(map[string]interface{}{
			"success": false,
			"data":    "Invalid password",
		})
	}

	return c.JSON(map[string]interface{}{
		"success": true,
		"data":    uid,
	})
}

func (e *Elemental) newAnonymousUser(c *fiber.Ctx) error {
	count := 1
	var name string
	var err error
	for count != 0 {
		name, err = GenerateRandomStringURLSafe(8)
		if err != nil {
			return c.JSON(map[string]interface{}{
				"success": false,
				"data":    err.Error(),
			})
		}

		// Check if name taken
		res, err := e.db.Query("SELECT COUNT(1) FROM users WHERE name=? LIMIT 1", name)
		if err != nil {
			return c.JSON(map[string]interface{}{
				"success": false,
				"data":    err.Error(),
			})
		}
		defer res.Close()
		res.Next()
		err = res.Scan(&count)
		if err != nil {
			return c.JSON(map[string]interface{}{
				"success": false,
				"data":    err.Error(),
			})
		}
	}
	return c.JSON(map[string]interface{}{
		"success": true,
		"data":    name,
	})
}
