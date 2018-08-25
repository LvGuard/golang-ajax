package main

import (
    "fmt"
    "net/http"
    "net/url"
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

func GetLngData(lid string, language string) string {
    var lngValue string
    query := `
        SELECT COALESCE(Language_` + language + `,'LNG_'||lid) LngValue
          FROM LngData
         WHERE lid = $1;`
    rows, err := db.Query(query, lid)
    if err != nil {
        fmt.Println("ERROR: psql lng1: ",err.Error());
        return "LNG_"+lid
    }
    for rows.Next() {
        err = rows.Scan(&lngValue)
        if err != nil {
            fmt.Println("ERROR: psql lng2: ",err.Error());
            return "LNG_"+lid
        }
    }
    if (lngValue == "") {
        fmt.Println("ERROR: ",lid,"not found in LngData table for Language",language);
        return "LNG_"+lid
    }
    return lngValue
}

type LNG_DATA struct {
    Language string
}
func (l LNG_DATA) LNG(lid string) string {
    return GetLngData(lid, l.Language)
}

type UpdateCustomer struct {
    Cid            int
    FieldToUpdate  string
    ValueToUpdate  string
    LastValue      string
    LastUpdate     string
    ResponseError    string
    ResponseSuccess  string
    ReloadFlag       bool
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
func (c Customer) IsSelectedGender(gender string) bool {
    return (gender == c.Gender)
}

type DataTable struct {
    Customers    []Customer
    Language     string
}
func (dt DataTable) LNG(lid string) string {
    return GetLngData(lid, dt.Language)
}

type Context struct {
    SearchBy          string
    OrderByField      string
    OrderByDirection  string
    PagesCount   int
    CurrentPage  int
    Data         string
    Language     string
}

func SanitizeLng(lng string) string {
    switch lng {
    case
        "2",
        "3":
        return lng
    }
    return "1"
}
func GetLng(r_url string) string {
    m, _ := url.ParseQuery(r_url)
    var lng string
    if _, ok := m["lng"]; ok {
        lng=m["lng"][0]
    }
    return SanitizeLng(lng)
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
    if (r.URL.Path != "/") { return }

    lngData := LNG_DATA{GetLng(r.URL.RawQuery)}

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

func DataTableLoad(context *Context, dataTable *DataTable) {
    var recordsCount int
    var recordsPerPage int = 10
    var limitOffset string = fmt.Sprintf("limit %v offset %v", recordsPerPage, recordsPerPage*context.CurrentPage)
    //fmt.Println(limitOffset)
    query := `
        SELECT count(1) recordsCount
          FROM Customers c
         WHERE '%'||$1||'%' = ''
            OR lower(c.FirstName)  like lower('%'||$1||'%')
            OR lower(c.LastName)   like lower('%'||$1||'%')
            OR to_char(c.BirthDate,'YYYY-MM-DD') like '%'||$1||'%'
            OR lower(c.Gender)     like lower('%'||$1||'%')
            OR lower(c.Email)      like lower('%'||$1||'%')
            OR lower(c.Address)    like lower('%'||$1||'%')
        ;`
    rows, err := db.Query(query, context.SearchBy)
    CheckErr(err)
    for rows.Next() {
        err = rows.Scan(&recordsCount)
        CheckErr(err)
    }
    context.PagesCount = (recordsCount / recordsPerPage) + 1
    //fmt.Println("PagesCount: ",context.PagesCount,"CurrentPage: ", context.CurrentPage, "Records per page: ",recordsPerPage)

    query = `
        SELECT c.Cid, c.FirstName, c.LastName, to_char(c.BirthDate,'YYYY-MM-DD') BirthDate, c.Gender, c.Email, COALESCE(c.Address,''), c.LastUpdate
          FROM public.customers c
         WHERE '%'||$1||'%' = ''
            OR lower(c.FirstName)  like lower('%'||$1||'%')
            OR lower(c.LastName)   like lower('%'||$1||'%')
            OR to_char(c.BirthDate,'YYYY-MM-DD') like '%'||$1||'%'
            OR lower(c.Gender)     like lower('%'||$1||'%')
            OR lower(c.Email)      like lower('%'||$1||'%')
            OR lower(c.Address)    like lower('%'||$1||'%')
         ORDER BY ` + SanitizeOrderByField(context.OrderByField) + ` ` + SanitizeOrderByDirection(context.OrderByDirection) + `
        ` + limitOffset + `
        ;`
    //fmt.Println(query);
    rows, err = db.Query(query, context.SearchBy)
    CheckErr(err)
    dataTable.Language = SanitizeLng(context.Language)
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
}

func DataTableHandler(w http.ResponseWriter, r *http.Request) {
    var context Context
    err := json.NewDecoder(r.Body).Decode(&context)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }

    var dataTable DataTable
    DataTableLoad(&context, &dataTable)

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

    // create json response from struct
    a, err := json.Marshal(context)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    w.Write(a)
}


func IsValidFirstName(t string, language string) (bool, string, string) {
    if (len(t)==0) {
        return false, GetLngData("MANDATORY_FIELDS_CANNOT_BE_EMPTY", language), ""
    }
    if (len(t)>100) {
        return false, "First Name maximal supported length is 100", ""
    }
    return true, "", ""
}

func IsValidLastName(t string, language string) (bool, string, string) {
    if (len(t)==0) {
        return false, GetLngData("MANDATORY_FIELDS_CANNOT_BE_EMPTY", language), ""
    }
    if (len(t)>100) {
        return false, "Last Name maximal supported length is 100", ""
    }
    return true, "", ""
}

func IsValidBirthDate(t string, language string) (bool, string, string) {
    if (len(t)==0) {
        return false, GetLngData("MANDATORY_FIELDS_CANNOT_BE_EMPTY", language), ""
    }
    if (len(t)!=10) {
        return false, GetLngData("PLEASE_PROVIDE_VALID_BIRTH_DATE", language), ""
    }
    if m, _ := regexp.MatchString("^[0-9]{4}-[0-9]{2}-[0-9]{2}$", t); !m {
        return false, GetLngData("PLEASE_PROVIDE_VALID_BIRTH_DATE_IN_FORMAT", language), ""
    }
    //between 18 and 60 years old
    query := `
        SELECT ( (to_date($1,'YYYY-MM-DD') > (now() - interval '60 year'))
             AND (to_date($1,'YYYY-MM-DD') < (now() - interval '18 year')) )
        DateCheck;`
    rows, err := db.Query(query, t)
    if err != nil {
        return false, GetLngData("PLEASE_PROVIDE_VALID_BIRTH_DATE_IN_FORMAT", language), ""
    }
    var dateCheck string
    for rows.Next() {
        err = rows.Scan(&dateCheck)
        if err != nil {
            return false, GetLngData("PLEASE_PROVIDE_VALID_BIRTH_DATE_IN_FORMAT", language), ""
        }
    }
    if (dateCheck == "false") {
        return false, GetLngData("PLEASE_PROVIDE_VALID_BIRTH_DATE_FROM_TO", language), ""
    }
    return true, "", ""
}

func IsValidGender(t string, language string) (bool, string, string) {
    if (len(t)==0) {
        return false, GetLngData("MANDATORY_FIELDS_CANNOT_BE_EMPTY", language), ""
    }
    switch t {
    case
        "Male",
        "Female":
        return true, "", ""
    }
    return false, GetLngData("GENDER_FIELDS_SUPPORTS_ONLY_VALUES", language), ""
}

func IsValidEmail(t string, language string) (bool, string, string) {
    if (len(t)==0) {
        return false, GetLngData("MANDATORY_FIELDS_CANNOT_BE_EMPTY", language), ""
    }
    if (len(t)>100) {
        return false, "Email maximal supported length is 100", ""
    }
    if m, _ := regexp.MatchString(`^([\w\.\_]{2,30})@(\w{1,})\.([a-z]{2,4})$`, t); !m {
        return false, GetLngData("PLEASE_PROVIDE_VALID_EMAIL", language), ""
    }
    return true, "", ""
}

func IsValidAddress(t string, language string) (bool, string, string) {
    if (len(t)>200) {
        return false, "Address maximal supported length is 200", ""
    }
    return true, "", ""
}


func InsertValidate(c *Customer, language string) bool {
    //All checks done here, as Javascript on client side could be easily disabled
    var r bool
    //FirstName
    if r, c.ResponseError, c.ResponseSuccess = IsValidFirstName(c.FirstName, language); !r {
        return false
    }
    //LastName
    if r, c.ResponseError, c.ResponseSuccess = IsValidLastName(c.LastName, language); !r {
        return false
    }
    //BirthDate
    if r, c.ResponseError, c.ResponseSuccess = IsValidBirthDate(c.BirthDate, language); !r {
        return false
    }
    //Gender
    if r, c.ResponseError, c.ResponseSuccess = IsValidGender(c.Gender, language); !r {
        return false
    }
    //Email
    if r, c.ResponseError, c.ResponseSuccess = IsValidEmail(c.Email, language); !r {
        return false
    }
    //Address
    if r, c.ResponseError, c.ResponseSuccess = IsValidAddress(c.Address, language); !r {
        return false
    }
    return true
}

func UpdateValidate(uc UpdateCustomer, language string) (bool, string, string) {
    //All checks done here, as Javascript on client side could be easily disabled
    switch uc.FieldToUpdate {
    case "FirstName":
        return IsValidFirstName(uc.ValueToUpdate, language)
    case "LastName":
        return IsValidLastName(uc.ValueToUpdate, language)
    case "BirthDate":
        return IsValidBirthDate(uc.ValueToUpdate, language)
    case "Gender":
        return IsValidGender(uc.ValueToUpdate, language)
    case"Email":
        return IsValidEmail(uc.ValueToUpdate, language)
    case "Address":
        return IsValidAddress(uc.ValueToUpdate, language)
    }
    return false, GetLngData("INVALID_OPERATION_PLEASE_RELOAD", language), ""
}

func InsertProcess(c *Customer, language string) bool {
    query := `
        INSERT INTO Customers
        (FirstName, LastName, BirthDate, Gender, Email, Address, LastUpdate)
        VALUES
        ($1, $2, $3, $4, $5, $6, NOW() )
        returning cid;`
    err := db.QueryRow(query, c.FirstName, c.LastName, c.BirthDate, c.Gender, c.Email, c.Address).Scan(&c.Cid)
    if err != nil {
        c.ResponseError = "psql ins1: " + err.Error()
        fmt.Println(c.ResponseError)
        return false
    }
    //fmt.Println("InsertProcess added record with cid =", c.Cid)
    c.ResponseSuccess = GetLngData("NEW_RECORD_SUCCESSFULLY_ADDED", language)
    return true
}

func UpdateProcess(uc UpdateCustomer, language string) (bool, string, string, string, string, bool) {
    query := `UPDATE Customers
                 SET ` + uc.FieldToUpdate + ` = $1
                   , LastUpdate = NOW()
               WHERE cid = $2
                 AND ` + uc.FieldToUpdate + ` = $3
                 AND lastUpdate = $4;`
    stmt, err := db.Prepare(query)
    if err != nil {
        return false, "psql err1: " + err.Error(), "", uc.LastValue, uc.LastUpdate, false
    }
    res, err := stmt.Exec(uc.ValueToUpdate, uc.Cid, uc.LastValue, uc.LastUpdate) // also checking LastValue in case value was changed in DB manually without updating LastUpdate
    if err != nil {
        return false, "psql err2: " + err.Error(), "", uc.LastValue, uc.LastUpdate, false
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return false, "psql err3: " + err.Error(), "", uc.LastValue, uc.LastUpdate, true
    }
    //fmt.Println(rowsAffected, "rows affected")

    var dbLastValue, dbLastUpdate string
    tmpSelect := uc.FieldToUpdate
    if (uc.FieldToUpdate == "BirthDate") {tmpSelect = "to_char(BirthDate,'YYYY-MM-DD')"}
    query = `
        SELECT ` + tmpSelect + ` dbLastValue
             , LastUpdate dbLastUpdate
          FROM Customers
         WHERE cid = $1;`
    rows, err := db.Query(query, uc.Cid)
    if err != nil {
        return false, "psql err4: " + err.Error(), "", uc.LastValue, uc.LastUpdate, true
    }
    for rows.Next() {
        err = rows.Scan(&dbLastValue, &dbLastUpdate)
        if err != nil {
            return false, "psql err5: " + err.Error(), "", uc.LastValue, uc.LastUpdate, true
        }
    }

    if (rowsAffected!=1) {
        if (dbLastValue!=uc.ValueToUpdate) {
            return false, GetLngData("VALUE_CHANGED_BY_ANOTHER_USER_RELOADED", language), "", dbLastValue, dbLastUpdate, true
        }
        return true, "", GetLngData("VALUE_ALREADY_UPDATED_BY_ANOTHER_USER_RELOADED", language)+dbLastValue, dbLastValue, dbLastUpdate, true
    } 
    return true, "", GetLngData("UPDATED_SUCCESSFULLY_OLD_VALUE", language)+uc.LastValue+GetLngData("NEW_VALUE", language)+dbLastValue+"\"", dbLastValue, dbLastUpdate, false
}

func InsertHandler(w http.ResponseWriter, r *http.Request) {
    language := GetLng(r.URL.RawQuery)

    //parse request to struct
    var c Customer
    err := json.NewDecoder(r.Body).Decode(&c)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    
    var res bool
    res = InsertValidate(&c, language)
    if (res) {
        res = InsertProcess(&c, language)
    }

    // create json response from struct
    a, err := json.Marshal(c)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    w.Write(a)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
    language := GetLng(r.URL.RawQuery)

    //parse request to struct
    var uc UpdateCustomer
    err := json.NewDecoder(r.Body).Decode(&uc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    
    //fmt.Println("Update - input",uc.Cid,uc.FieldToUpdate,uc.ValueToUpdate,uc.LastValue,uc.LastUpdate,uc.ResponseError,uc.ResponseSuccess,uc.ReloadFlag)
    var res bool
    res, uc.ResponseError, uc.ResponseSuccess = UpdateValidate(uc, language)
    if (res) {
        res, uc.ResponseError, uc.ResponseSuccess, uc.LastValue, uc.LastUpdate, uc.ReloadFlag = UpdateProcess(uc, language)
    }

    //fmt.Println("Update - output",uc.Cid,uc.FieldToUpdate,uc.ValueToUpdate,uc.LastValue,uc.LastUpdate,uc.ResponseError,uc.ResponseSuccess,uc.ReloadFlag)
    // create json response from struct
    a, err := json.Marshal(uc)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    w.Write(a)
}

func DeleteProcess(cid int) (bool, string) {
    stmt, err := db.Prepare("delete from Customers where cid=$1")
    if err != nil { return false, err.Error() }

    res, err := stmt.Exec(cid)
    if err != nil { return false, err.Error() }

    deletedRows, err := res.RowsAffected()
    if err != nil { return false, err.Error() }

    if deletedRows != 1 { return false, fmt.Sprintf("deletedRows(%v) != 1",deletedRows) }

    return true, ""
}

func InitDb() {
    fmt.Printf("Connecting\n")
    var err error
    dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
    db, err = sql.Open("postgres", dbinfo)
    CheckErr(err)
}

func main() {
    fmt.Printf("Start\n")
    InitDb()

    fmt.Printf("HTTP handler startup\n")
    http.HandleFunc("/", DefaultHandler)
    http.HandleFunc("/DataTable", DataTableHandler)
    http.HandleFunc("/Insert", InsertHandler)
    http.HandleFunc("/Update", UpdateHandler)
    http.ListenAndServe(":8080", nil)

    defer db.Close()
    fmt.Printf("End\n")
}
