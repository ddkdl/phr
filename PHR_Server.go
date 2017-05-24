/**********PHR_Server************/
//This file is for the main web page for the P2HR.
//The main function is responsible for routing between
// three html templates: Login, Edit, and View.
// More templates will be added. The purpose of this
// code is for a P2P PHR platform. The database used
// is MongoDB.
package main

import (
	"net/http"
	"text/template"
	"strings"
	"fmt"
	"encoding/json"
	"gopkg.in/mgo.v2" // mgo packages
	"gopkg.in/mgo.v2/bson"
	"./Encounters_JSON"
	"./MongoFuncs"
)

type Conditions struct{
	Id   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string
	Encounters []bson.ObjectId
}

type Encounters struct{
	Id   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Text string
}

type Friends struct{
	First string
	Last string
}

type Connections struct{
	Id   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Friends []Friends
	Doctors []Friends
}

type Person struct{
	Conditions []Conditions
	Encounters []Encounters
	Connections Connections
	Str string
}

type User struct{
	Condition string
	Code string
	Collection string
}

var tempString string
var user string; // Global variables for the username and password
var pass string;
var onskip int;

/********* Handler functions ********/
// Each service a specific html script
type Login struct{ //login page
}
type Edit struct{ //edit page
}
type Condition struct{ // view condition page
}


/*********** View Condition Handler *********/
// This uses the URL query string to determine which condition
// the user is quering. Once that string is determined the
// condition information is pulled from Mongo and the
// html template is parsed
func (this *Condition) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	str := req.URL.Path // pulls the query string
	str = string(str) // makes the input a string
	str = strings.Replace(str, "/Condition/", "", 1)

	// pulls from the local Mongo database
	session, err :=mgo.Dial("localhost") //dials the database
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("Sonja").C("Conditions") // Currently the database is hardcoded but once the login authorization is complete, the name should be the database
	var result Person
	err = c.Find(nil).All(&result.Conditions) //finds all the encounters with that condition
	c = session.DB("Sonja").C("Encounters")
	err = c.Find(nil).All(&result.Encounters)
	c = session.DB("Sonja").C("Connections")
	err = c.Find(nil).One(&result.Connections)
	result.Str = str
	if err != nil {
		panic(err)
	}
	tmpl, err := template.ParseFiles("Condition_View_Template.html") // Parse the html file
	if err == nil{
		tmpl.Execute(res, result) // execute the response with the struct for the pipelines
	}
}

/************** Edit Service Handler **************/
// This pulls all the conditions in the local database
// and posts them to the web page. The user then can drag
// and drop medical information into a condition. This
// infomation is then indexed to that condition in the
// database
func (this *Edit) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	session, err :=mgo.Dial("localhost")// dials the local database
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("Sonja").C("Conditions")
	var result Person
	var old Person
	err = c.Find(nil).All(&result.Conditions) // two copies are needed for the updating
	err = c.Find(nil).All(&old.Conditions)
	c = session.DB("Sonja").C("Encounters")
	err = c.Find(nil).All(&result.Encounters)
	err = c.Find(nil).All(&old.Encounters)
	c = session.DB("Sonja").C("Connections")
	err = c.Find(nil).One(&result.Connections)
	err = c.Find(nil).One(&old.Connections)

	if err != nil {
		panic(err)
	}

	if req.Method == "POST"{ //There are two types of POST requests that can occur in the page: either a file POST, or a post from the drag and drop
		req.ParseMultipartForm(0)
		if len(req.Form["NewRecSub"]) >= 1{ //A POST request is sent once a file is uploaded
			conFigFile, _, err := req.FormFile("up") // Looks for the ID "up" for the FormFile. This tag is defined in the html template
			if err != nil{
				panic(err)
			}
			var En Encounter
    	jsonParser := json.NewDecoder(conFigFile) // The FHIR file is decoded from its JSON format to a struct
    	if err = jsonParser.Decode(&En); err != nil {
        fmt.Println("parsing config file", err.Error())
    	}
			var new Encounters
			new.Text = En.Reason[0].Text // Currently only the text is of interest however the other fields will be of interest as the PHR becomes more sophisticated
			session, err := mgo.Dial("localhost")
      if err != nil{
 			 panic(err)
 		 }
      defer session.Close()
      session.SetMode(mgo.Monotonic, true)
      d := session.DB("Sonja").C("Encounters") // for a proof of concept the encounters are only used
			err = d.Insert(&new) // Insert the new encounters. In the future a for loop must be used to cycle through all the medical infomation
			if err != nil{
				panic(err)
			}
    	fmt.Println(En)

		} else{ // if a drag and drop POST request was sent
			var u User
	    if req.Body != nil {
	      err := json.NewDecoder(req.Body).Decode(&u) // decodes the JSON message
				if err != nil {
	  			http.Error(res, err.Error(), 400)
	  			return
				}
				fmt.Println("The Code is", u.Code)
				fmt.Println("The Condition Name is", u.Condition)
				fmt.Println("The Collection is ", u.Collection)
	    }
			var ObjIdSlice []string
			ObjIdSlice = append(ObjIdSlice, u.Code[13:37]) // removes the "ObjectID" from the string
			ObjIdString := strings.Join(ObjIdSlice, "")
			var count int
			for i:=0; i < len(result.Conditions); i++{
				if u.Condition == result.Conditions[i].Name{
					count = i
					fmt.Println("The count is:", i) // finds which condition was selected
				}
			}
			result.Conditions[count].Encounters = append(result.Conditions[count].Encounters, bson.ObjectIdHex(ObjIdString)) // indexes the new information to the condition
			fmt.Println(result.Conditions[count].Encounters)
			update("localhost", old, result, "Sonja", "Conditions") // update the database
		}
	}

	tmpl, err := template.ParseFiles("Condition_Edit_Template.html") // Parse the html file1
	if err == nil{
		tmpl.Execute(res, result) // execute the response with the struct for the pipelines
	}
}

/**********Login Service Handler***********/
//This handler is used for displaying the login page.
//The username and password is posted back to the server,
//this infomation will be used later for authentication
//puposes
func (this *Login) ServeHTTP(res http.ResponseWriter, req *http.Request) { //Serves the web page
	onskip := 0;
	//Awaits for a method form from the user
	if req.Method == "GET" { // Get form
	} else { //post form
		req.ParseForm() //Parse the form input
		req.ParseMultipartForm(0) //Parse Mulitpart forms - (This is used specifically for the file uploading in the home page)
		onskip =1
		// Sign In post request from the login html page
		if len(req.Form["SignIn"]) >= 1{ //if a login is intiated
			fmt.Println("Login")
			session, err :=mgo.Dial("localhost") // All data on this user is pulled from Mongo and then parsed into the html home template
			if err != nil {
				panic(err)
			}
			defer session.Close()
			session.SetMode(mgo.Monotonic, true)
			c := session.DB("Sonja").C("Conditions")
			var result Person
	    err = c.Find(nil).All(&result.Conditions)
			c = session.DB("Sonja").C("Encounters")
			err = c.Find(nil).All(&result.Encounters)
			c = session.DB("Sonja").C("Connections")
			err = c.Find(nil).One(&result.Connections)
	    if err != nil {
	      panic(err)
	    }
			tmpl, err := template.ParseFiles("Home_Template.html") // Parse the html file
			if err == nil{
			 	tmpl.Execute(res, result) // execute the response with the struct for the pipelines
			}
		}
		}
		// If no request method is made, the login html page will be serviced
	if onskip == 0{
		tmpl, err := template.ParseFiles("Login_Template.html") // grabs the template file
		if err != nil{
			panic(err)
		}
		tmpl.Execute(res, "")
	}
}

/***********Main Function***********/
// This function is responsible for all the routing
func main() {
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images")))) // This directs all images to the images folder
	http.Handle("/", new(Login)) // login service handler
	http.Handle("/Edit", new(Edit)) // edit service handler
	http.Handle("/Condition/", new(Condition)) // view condition handler

	http.ListenAndServe(":8080", nil) //Listens on Port 8080

}
