package main

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
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
	Id    int
}

func newContact(name string, email string) Contact {
	id++
	return Contact{
		Name:  name,
		Email: email,
		Id:    id,
	}

}

func indexOfContact(page Page, id int) (int, error) {
	for i, contact := range page.Data.Contacts {
		if contact.Id == id {
			return i, nil
		}
	}
	return 0, errors.New("Contact not found")
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("A", "a@a.com"),
			newContact("B", "b@b.com"),
			newContact("C", "c@c.com"),
		},
	}
}

func (d *Data) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: map[string]string{},
		Errors: map[string]string{},
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
	e.Use(middleware.Logger())
	e.Static("/images", "images")
	e.Static("/css", "css")

	page := newPage()
	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Errors["email"] = "Email already exists"
			formData.Values["name"] = name
			formData.Values["email"] = email

			return c.Render(422, "form", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		c.Render(http.StatusOK, "form", newFormData())

		return c.Render(http.StatusOK, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		time.Sleep(4 * time.Second)
		idstr := c.Param("id")
		id, err := strconv.Atoi(idstr)
		if err != nil {
			return c.String(400, "Invalid id")

		}
		i, err := indexOfContact(page, id)
		if err != nil {
			return c.String(404, "id doesn't exist")
		}

		page.Data.Contacts = append(page.Data.Contacts[:i], page.Data.Contacts[i+1:]...)

		return c.NoContent(200)

	})
	e.Logger.Fatal(e.Start(":42069"))
}
