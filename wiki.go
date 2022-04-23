package main

import (
	"html/template" //add html/template to the list of imports. We also won't be using fmt anymore, so we have to remove that.
	"log"
	"net/http"
	"os"
	"regexp"
)

//Data Structure
type Page struct {
	Title string
	Body  []byte
}
//SavingPage
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}
//LoadingPage
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)//os.ReadFile returns []byte and error.
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}



//Handling non-existent pages
//An http.ResponseWriter value assembles the HTTP server's response; by writing to it, we send data to the HTTP client.
//An http.Request is a data structure that represents the client HTTP request.
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title) 
    //This is because it ignores the error return value from loadPage and continues to try and fill out the template with no data
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}//Instead, if the requested Page doesn't exist, it should redirect the client to the edit Page so the content may be created:

//Editing Pages
//The function editHandler loads the page (or, if it doesn't exist, create an empty Page struct), and displays an HTML form.
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

//Saving Pages
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")//The value returned by FormValue is of type string
	p := &Page{Title: title, Body: []byte(body)}
    //We must convert that value to []byte before it will fit into the Page struct. We use []byte(body) to perform the conversion.
    //The page title (provided in the URL)and the form's only field, Body, are stored in a new Page.
	err := p.save()//The save() method is then called to write the data to a file
    //Any errors that occur during p.save() will be reported to the user.
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        //http.Error function sends a specified HTTP response code (in this case "Internal Server Error") and error message. 
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)//the client is redirected to the /view/ page.
}//writer,request,page

//Template caching
//we create a global variable named templates, and initialize it with ParseFiles.
//template.Must is a convenience wrapper that panics when passed a non-nil error value, and otherwise returns the *Template unaltered.
//The ParseFiles function takes any number of string arguments that identify our template files
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

//Error Handling 
//renderTemplate function to call the templates.ExecuteTemplate method with the name of the appropriate template:
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//validation
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

//define a wrapper function
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //The returned function is called a closure because it encloses values defined outside of it.
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2]) //The variable fn will be one of our save, edit, or view handlers.
	} //w= http.ResponseWriter, r= *http.Request, match
} //In this case, the variable fn (the single argument to makeHandler) is enclosed by the closure.

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}//ListenAndServe always returns an error, since it only returns when an unexpected error occurs. 
//In order to log that error we wrap the function call with log.Fatal.
