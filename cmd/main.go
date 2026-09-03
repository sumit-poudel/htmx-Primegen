package main

import (
	"html/template"
	"io"
	"net/http"

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

type Contact struct {
	Name  string
	Email string
}

func newContact(name string, email string) Contact {
	return Contact{
		Name:  name,
		Email: email,
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

	e.Any("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	if err := e.Start(":8090"); err != nil {
		e.Logger.Error("failed to start the server", "error", err)
	}
}
