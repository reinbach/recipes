package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	STATIC_URL  = "/static/"
	STATIC_ROOT = "static/"
	RECIPE_DIR  = "recipes/"
)

var (
	recipes []Recipe
)

type Context struct {
	Title  string
	Static string
	Data   interface{}
}

type Recipe struct {
	Title string
	Path  string
}

func (r *Recipe) Name() string {
	return strings.Title(strings.Replace(r.Title, "_", " ", -1))
}

func (r *Recipe) Body() string {
	data, err := ioutil.ReadFile(r.Path)
	if err != nil {
		log.Print("Failed to read file: ", err)
		return fmt.Sprintf("Unable to get Recipe: %s", r.Name())
	}
	body := strings.Replace(string(data), "\n", "<br />", -1)
	return body
}

func GetRecipeByTitle(t string) (Recipe, error) {
	for _, r := range recipes {
		if r.Title == t {
			return r, nil
		}
	}

	return Recipe{}, fmt.Errorf("Not Found")
}

func SetRecipe(p string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		r := Recipe{
			Title: info.Name(),
			Path:  p,
		}
		recipes = append(recipes, r)
	}

	return nil
}

func WalkRecipes() {
	// clear recipes
	recipes = make([]Recipe, 0)

	err := filepath.Walk(RECIPE_DIR, SetRecipe)
	if err != nil {
		log.Print("Error walking recipes: ", err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	WalkRecipes()
	context := Context{Title: "Recipes", Data: recipes}
	Render(w, "index", context)
}

func RecipeHandler(w http.ResponseWriter, r *http.Request) {
	// Make sure recipes is not empty
	if len(recipes) == 0 {
		WalkRecipes()
	}
	title := r.FormValue("title")
	recipe, err := GetRecipeByTitle(title)
	if err != nil {
		log.Print("Failed to find recipe: ", err)
	}
	context := Context{
		Title: fmt.Sprintf("Recipe: %s", recipe.Name()),
		Data:  template.HTML(recipe.Body()),
	}
	Render(w, "recipe", context)
}

func Render(w http.ResponseWriter, tmpl string, context Context) {
	context.Static = STATIC_URL
	tmpl_list := []string{
		"templates/base.html",
		fmt.Sprintf("templates/%s.html", tmpl),
	}
	t, err := template.ParseFiles(tmpl_list...)
	if err != nil {
		log.Print("Template parsing error: ", err)
	}

	err = t.Execute(w, context)
	if err != nil {
		log.Print("Template executing error: ", err)
	}
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	static_file := r.URL.Path[len(STATIC_URL):]
	if len(static_file) != 0 {
		f, err := http.Dir(STATIC_ROOT).Open(static_file)
		if err == nil {
			content := io.ReadSeeker(f)
			http.ServeContent(w, r, static_file, time.Now(), content)
			return
		}
	}

	http.NotFound(w, r)
}

func main() {
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/recipe/", RecipeHandler)
	http.HandleFunc(STATIC_URL, StaticHandler)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
