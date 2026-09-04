package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(c *echo.Context, w io.Writer, name string, data any) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

var id = 0

type Contact struct {
	Name  string
	Email string
	ID    int
}

func newContact(name string, email string) Contact {
	id++
	return Contact{
		Name:  name,
		Email: email,
		ID:    id,
	}
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
}

func (d *Data) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}
func (d *Data) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.ID == id {
			return i
		}
	}
	return -1
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("Sumit", "test"),
			newContact("Foss", "foss@go.com"),
		},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.Renderer = newTemplate()
	e.Static("/css", "css")
	e.Static("/images", "images")

	page := newPage()

	e.GET("/", func(c *echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/contacts", func(c *echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Errors = map[string]string{
				"email": "already used email",
			}
			formData.Values = map[string]string{
				"email": email,
				"name":  name,
			}

			return c.Render(http.StatusUnprocessableEntity, "form", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)
		e.Logger.Info("Adding data", "name", name, "email", email)

		c.Render(http.StatusOK, "form", newFormData())
		return c.Render(http.StatusOK, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c *echo.Context) error {
		time.Sleep(3 * time.Second)
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(http.StatusUnprocessableEntity, "Invalid id")
		}
		index := page.Data.indexOf(id)
		if index == -1 {
			return c.String(http.StatusNotFound, "Contact not found")
		}
		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)
		return c.NoContent(http.StatusOK)
	})

	e.Any("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	if err := e.Start(":8090"); err != nil {
		e.Logger.Error("failed to start the server", "error", err)
	}
}
