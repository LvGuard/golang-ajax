package main

import (
    "fmt"
    "net/http"
    "html/template"
    "database/sql"
    _ "github.com/lib/pq"
    "encoding/json"
    "path"
    "bytes"
    "regexp"
)

const (
    DB_USER     = "postgres"
    DB_PASSWORD = "p"
    DB_NAME     = "test"
)
var db *sql.DB

func CheckErr(err error) {
    if err != nil {
        panic("psql err: " + err.Error())
    }
}

type LNG_DATA struct {
    LNG_GREETINGS string
}

type UpdateCustomer struct {
    Cid            int
    FieldToUpdate  string
    ValueToUpdate  string
    LastUpdate     string
    ResponseError    string
    ResponseSuccess  string
}

type Customer struct {
    Cid         int
    FirstName   string
    LastName    string
    BirthDate   string
    Gender      string
    Email       string
    Address     string
    LastUpdate  string
    ResponseError    string
    ResponseSuccess  string
}
type DataTable struct {
    Customers      []Customer
}

type Context struct {
    SearchBy          string
    OrderByField      string
    OrderByDirection  string
    PagesCount   int
    CurrentPage  int
    Data         string
}


func DefaultHandler(w http.ResponseWriter, r *http.Request) {
    lngData := LNG_DATA{"Welcome to cutomer portal!"}
    
    fp := path.Join("templates", "MainPage.html")
    tmpl, err := template.ParseFiles(fp)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if err := tmpl.Execute(w, lngData); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func SanitizeOrderByField(orderByField string) string {
    switch orderByField {
    case
        "2",
        "3",
        "4",
        "5",
        "6",
        "7":
        return orderByField
    }
    return "2"
}

func SanitizeOrderByDirection(orderByDirection string) string {
    switch orderByDirection {
    case
        "asc",
        "desc":
        return orderByDirection
    }
    return "asc"
}

func DataTableHandler(w http.ResponseWriter, r *http.Request) {
    var context Context
    err := json.NewDecoder(r.Body).Decode(&context)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    //fmt.Printf("%v %v\n", context, context.PagesCount)

    query := `
        SELECT c.Cid, c.FirstName, c.LastName, to_char(c.BirthDate,'YYYY-MM-DD') BirthDate, c.Gender, c.Email, COALESCE(c.Address,''), c.LastUpdate
          FROM public.customers c
         WHERE '%'||$1||'%' = ''
            OR lower(c.FirstName)  like lower('%'||$1||'%')
            OR lower(c.LastName)   like lower('%'||$1||'%')
            OR to_char(c.BirthDate,'YYYY-MM-DD') like '%'||$1||'%'
            OR lower(c.Gender)     like lower('%'||$1||'%')
            OR lower(c.Email)      like lower('%'||$1||'%')
            OR lower(c.Address)    like lower('%'||$1||'%')
         ORDER BY ` + SanitizeOrderByField(context.OrderByField) + ` ` + SanitizeOrderByDirection(context.OrderByDirection) + `;`
    //fmt.Println(query);
    rows, err := db.Query(query, context.SearchBy)
    CheckErr(err)
    var dataTable DataTable
    for rows.Next() {
        //var created time.Time
        var cid int
        var firstName, lastName, birthDate, gender, email, address, lastUpdate  string
        err = rows.Scan(&cid, &firstName, &lastName, &birthDate, &gender, &email, &address, &lastUpdate)
        CheckErr(err)
        //fmt.Printf("%v | %v | %v | %v <br>\n", cid, firstName, lastName, birthDate)
        dataTable.Customers = append(dataTable.Customers, Customer{
            Cid:         cid,
            FirstName:   firstName,
            LastName:    lastName,
            BirthDate:   birthDate,
            Gender:      gender,
            Email:       email,
            Address:     address,
            LastUpdate:  lastUpdate,
        })
    }

    fp := path.Join("templates", "DataTable.html")
    tmpl, err := template.ParseFiles(fp)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var tpl bytes.Buffer
    if err := tmpl.Execute(&tpl, dataTable); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    context.Data = tpl.String()
    context.PagesCount = 2

    // create json response from struct
    a, err := json.Marshal(context)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    w.Write(a)
}


func IsValidFirstName(t string) (bool, string, string) {
    if (len(t)==0) {
        return false, "Mandatory(*) fields cannot be empty", ""
    }
    if (len(t)>100) {
        return false, "First Name maximal supported length is 100", ""
    }
    return true, "", ""
}

func IsValidLastName(t string) (bool, string, string) {
    if (len(t)==0) {
        return false, "Mandatory(*) fields cannot be empty", ""
    }
    if (len(t)>100) {
        return false, "Last Name maximal supported length is 100", ""
    }
    return true, "", ""
}

func IsValidBirthDate(t string) (bool, string, string) {
    if (len(t)==0) {
        return false, "Mandatory(*) fields cannot be empty", ""
    }
    if (len(t)!=10) {
        return false, "Please provide valid Birth Date", ""
    }
    if m, _ := regexp.MatchString("^[0-9]{4}-[0-9]{2}-[0-9]{2}$", t); !m {
        return false, "Please provide valid Birth Date in format YYYY-MM-DD", ""
    }
    //between 18 and 60 years old
    query := `
        SELECT ( (to_date($1,'YYYY-MM-DD') > (now() - interval '60 year'))
             AND (to_date($1,'YYYY-MM-DD') < (now() - interval '18 year')) )
        DateCheck;`
    rows, err := db.Query(query, t)
    if err != nil {
        return false, "Please provide valid Birth Date in format YYYY-MM-DD", ""
    }
    var dateCheck string
    for rows.Next() {
        err = rows.Scan(&dateCheck)
        if err != nil {
            return false, "Please provide valid Birth Date in format YYYY-MM-DD", ""
        }
    }
    if (dateCheck == "false") {
        return false, "Please provide valid Birth Date from 18 till 60 years", ""
    }
    return true, "", ""
}

func IsValidGender(t string) (bool, string, string) {
    if (len(t)==0) {
        return false, "Mandatory(*) fields cannot be empty", ""
    }
    switch t {
    case
        "Male",
        "Female":
        return true, "", ""
    }
    return false, "Gender field supports only Male and Female values", ""
}

func IsValidEmail(t string) (bool, string, string) {
    if (len(t)==0) {
        return false, "Mandatory(*) fields cannot be empty", ""
    }
    if (len(t)>100) {
        return false, "Email maximal supported length is 100", ""
    }
    if m, _ := regexp.MatchString(`^([\w\.\_]{2,10})@(\w{1,}).([a-z]{2,4})$`, t); !m {
        return false, "Please provide valid Email", ""
    }
    return true, "", ""
}

func IsValidAddress(t string) (bool, string, string) {
    if (len(t)>200) {
        return false, "Address maximal supported length is 200", ""
    }
    return true, "", ""
}


func InsertValidate(c Customer) (bool, string, string) {
    //All checks done here, as Javascript on client side could be easily disabled
    var r bool
    var r_err, r_ok string
    //FirstName
    if r, r_err, r_ok = IsValidFirstName(c.FirstName); !r {
        return false, r_err, r_ok
    }
    //LastName
    if r, r_err, r_ok = IsValidLastName(c.LastName); !r {
        return false, r_err, r_ok
    }
    //BirthDate
    if r, r_err, r_ok = IsValidBirthDate(c.BirthDate); !r {
        return false, r_err, r_ok
    }
    //Gender
    if r, r_err, r_ok = IsValidGender(c.Gender); !r {
        return false, r_err, r_ok
    }
    //Email
    if r, r_err, r_ok = IsValidEmail(c.Email); !r {
        return false, r_err, r_ok
    }
    //Address
    if r, r_err, r_ok = IsValidAddress(c.Address); !r {
        return false, r_err, r_ok
    }
    return true, "", ""
}

func UpdateValidate(uc UpdateCustomer) (bool, string, string) {
    //All checks done here, as Javascript on client side could be easily disabled
    switch uc.FieldToUpdate {
    case "FirstName":
        return IsValidFirstName(uc.ValueToUpdate)
    case "LastName":
        return IsValidLastName(uc.ValueToUpdate)
    case "BirthDate":
        return IsValidBirthDate(uc.ValueToUpdate)
    case "Gender":
        return IsValidGender(uc.ValueToUpdate)
    case"Email":
        return IsValidEmail(uc.ValueToUpdate)
    case "Address":
        return IsValidAddress(uc.ValueToUpdate)
    }
    return false, "Invalid operation, please re-load the page", ""
}

func InsertProcess(c Customer) (bool, string, string) {
    var cid int
    query := `
        INSERT INTO Customers
        (FirstName, LastName, BirthDate, Gender, Email, Address, LastUpdate)
        VALUES
        ($1, $2, $3, $4, $5, $6, NOW() )
        returning cid;`
    err := db.QueryRow(query, c.FirstName, c.LastName, c.BirthDate, c.Gender, c.Email, c.Address).Scan(&cid)
    if err != nil {
        fmt.Println("InsertProcess psql err: " + err.Error())
        return false, "psql err: " + err.Error(), ""
    }
    fmt.Println("InsertProcess added record with cid =", cid)
    return true, "", "New record successfully added"
}

func UpdateProcess(uc UpdateCustomer) (bool, string, string, string, string) {
    query := `
        SELECT c.` + uc.FieldToUpdate + ` OriginalValue
             , c.LastUpdate
          FROM Customers c
         WHERE c.cid = $1;`
    rows, err := db.Query(query, uc.Cid)
    if err != nil {
        return false, "psql err1: " + err.Error(), "", uc.ValueToUpdate, uc.LastUpdate
    }
    var originalValue, lastUpdate string
    for rows.Next() {
        err = rows.Scan(&originalValue, &lastUpdate)
        if err != nil {
            return false, "psql err2: " + err.Error(), "", uc.ValueToUpdate, uc.LastUpdate
        }
    }
    fmt.Println("OriginalValue:",originalValue,"-",uc.ValueToUpdate,":ValueToUpdate")
    if (originalValue == uc.ValueToUpdate) {
        return true, "", "", originalValue, lastUpdate
    }

    if (lastUpdate != uc.LastUpdate) {
        return false, "Warning: Value changed by another user session, please verify new value and update if necessary!", "", originalValue, lastUpdate
    }

    query = `UPDATE Customers
                SET ` + uc.FieldToUpdate + ` = $1
                  , LastUpdate = NOW()
              WHERE cid = $2
                AND lastUpdate = $3;`
    fmt.Println("Update query:", query)
    stmt, err := db.Prepare(query)
    if err != nil {
        return false, "psql err3: " + err.Error(), "", uc.ValueToUpdate, uc.LastUpdate
    }

    res, err := stmt.Exec(uc.ValueToUpdate, uc.Cid, uc.LastUpdate)
    if err != nil {
        return false, "psql err4: " + err.Error(), "", uc.ValueToUpdate, uc.LastUpdate
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return false, "psql err5: " + err.Error(), "", uc.ValueToUpdate, uc.LastUpdate
    }
    fmt.Println(rowsAffected, "rows affected")

    query = `
        SELECT c.` + uc.FieldToUpdate + ` OriginalValue
             , c.LastUpdate
          FROM Customers c
         WHERE c.cid = $1;`
    rows, err = db.Query(query, uc.Cid)
    if err != nil {
        return false, "psql err6: " + err.Error(), "", uc.ValueToUpdate, uc.LastUpdate
    }
    for rows.Next() {
        err = rows.Scan(&originalValue, &lastUpdate)
        if err != nil {
            return false, "psql err7: " + err.Error(), "", uc.ValueToUpdate, uc.LastUpdate
        }
    }
    
    if (rowsAffected!=1) {
        return false, "Warning: Value changed by another user session, please verify new value and update if necessary!", "", originalValue, lastUpdate
    } 
    return true, "", "1 record updated at " + lastUpdate, originalValue, lastUpdate
}

func InsertHandler(w http.ResponseWriter, r *http.Request) {
    //parse request to struct
    var c Customer
    err := json.NewDecoder(r.Body).Decode(&c)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    
    var res bool
    res, c.ResponseError, c.ResponseSuccess = InsertValidate(c)
    if (res) {
        res, c.ResponseError, c.ResponseSuccess = InsertProcess(c)
    }

    // create json response from struct
    a, err := json.Marshal(c)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    w.Write(a)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
    //parse request to struct
    var uc UpdateCustomer
    err := json.NewDecoder(r.Body).Decode(&uc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    
    fmt.Println("Update - input",uc.Cid,uc.FieldToUpdate,uc.ValueToUpdate,uc.LastUpdate,uc.ResponseError,uc.ResponseSuccess)
    var res bool
    res, uc.ResponseError, uc.ResponseSuccess = UpdateValidate(uc)
    if (res) {
        res, uc.ResponseError, uc.ResponseSuccess, uc.ValueToUpdate, uc.LastUpdate = UpdateProcess(uc)
    }

    fmt.Println("Update - output",uc.Cid,uc.FieldToUpdate,uc.ValueToUpdate,uc.LastUpdate,uc.ResponseError,uc.ResponseSuccess)
    // create json response from struct
    a, err := json.Marshal(uc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    w.Write(a)
}


func main() {
    fmt.Printf("Start\n")

    fmt.Printf("Connecting\n")
    var err error
    dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
    db, err = sql.Open("postgres", dbinfo)
    CheckErr(err)
    defer db.Close()

    fmt.Printf("HTTP handler startup\n")

    http.HandleFunc("/", DefaultHandler)
    http.HandleFunc("/DataTable", DataTableHandler)
    http.HandleFunc("/Insert", InsertHandler)
    http.HandleFunc("/Update", UpdateHandler)
    http.ListenAndServe(":8080", nil)

    fmt.Printf("End\n")

}
