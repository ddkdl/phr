package MongoFuncs

/********** Update Mongo ************/
//takes as input the database and collect string, the
//database addr, and the new and old mongo documents
func update(databaseAddr string,old Person,new Person,database string,collection string ){
     session, err := mgo.Dial(databaseAddr)// dials the database address
     if err != nil{
			 panic(err)
		 }
     defer session.Close()
     session.SetMode(mgo.Monotonic, true)
     d := session.DB(database).C(collection)
		 for i:=0; i < len(old.Conditions); i++{ //iterates through all documents
			 err = d.Update(old.Conditions[i], &new.Conditions[i]) // updates
			 if err != nil{
				 panic(err)
			 }
		 }
		 fmt.Println("I updated")
}
