package main

import "testing"

import "fmt"

func TestIsValidGender(t *testing.T) {
    tables := []struct {
        v string
        r bool
    }{
        {"Male", true},
        {"male", false},
        {"Male1", false},
        {"Female", true},
        {"female", false},
        {"femalemlefemalemlefemalemlefemalemlefemalemlefemalemlefemalemlefemalemle", false},
        {"", false},
    }
    InitDb()
    for _, table := range tables {
        r, _, _ := IsValidGender(table.v, "1")
        if (r != table.r) {
           t.Errorf("IsValidGender was incorrect (%v) on %v", r, table.v)
        }
    }
    db.Close()
}

func TestSanitizeOrderByDirection(t *testing.T) {
    tables := []struct {
        v string
        r string
    }{
        {"asc", "asc"},
        {"desc", "desc"},
        {"", "asc"},
        {"Desc", "asc"},
        {"femalemlefemalemlefemalemlefemalemlefemalemlefemalemlefemalemlefemalemle", "asc"},
    }
    InitDb()
    for _, table := range tables {
        r := SanitizeOrderByDirection(table.v)
        if (r != table.r) {
           t.Errorf("SanitizeOrderByDirection was incorrect (%v) on %v", r, table.v)
        }
    }
    db.Close()
}

func TestIsValidBirthDate(t *testing.T) {
    tables := []struct {
        v string
        r bool
    }{
        {"08-09-1978", false},
        {"1980-09-01", true},
        {"1980/09/01", false},
        {"1980/09/011980/09/011980/09/011980/09/011980/09/011980/09/011980/09/011980/09/011980/09/01", false},
        {"1900-09-01", false},
        {"2050-09-01", false},
        {"1979-02-29", false},
        {"1979-03-32", false},
        {"", false},
    }
    InitDb()
    for _, table := range tables {
        r, _, _ := IsValidBirthDate(table.v, "1")
        if (r != table.r) {
           t.Errorf("IsValidBirthDate was incorrect (%v) on %v", r, table.v)
        }
    }
    db.Close()
}


func TestInsertValidate(t *testing.T) {
    tables := []struct {
        r bool
        v Customer
    }{
        {false, Customer{
                    FirstName:   "",
                    LastName:    "Last Name",
                    BirthDate:   "2020-09-01",
                    Gender:      "Male",
                    Email:       "info@tea.lt",
                    Address:     "",
        }},
        {false, Customer{
                    FirstName:   "First Name",
                    LastName:    "Last Name",
                    BirthDate:   "1990-09-01",
                    Gender:      "Male",
                    Email:       "info@tealt",
                    Address:     "",
        }},
        {true, Customer{
                    FirstName:   "First Name",
                    LastName:    "Last Name",
                    BirthDate:   "1990-09-01",
                    Gender:      "Male",
                    Email:       "info@tea.lt",
                    Address:     "",
        }},
    }
    InitDb()
    for _, table := range tables {
        r := InsertValidate(&table.v, "1")
        if (r != table.r) {
           t.Errorf("InsertValidate was incorrect (%v) on %v", r, table.v)
        }
    }
    db.Close()
}

func TestInsertProcess(t *testing.T) {
    tables := []struct {
        r bool
        v Customer
    }{
        {true, Customer{
                    FirstName:   "First Name",
                    LastName:    "Last Name",
                    BirthDate:   "1990-09-01",
                    Gender:      "Male",
                    Email:       "info@tea.lt",
                    Address:     "",
        }},
    }
    InitDb()
    for _, table := range tables {
        r := InsertProcess(&table.v, "1")
        if (r != table.r) {
           t.Errorf("InsertProcess was incorrect (%v) on %v", r, table.v)
        }
        fmt.Println("Inserted record with cid",table.v.Cid)
        r, err := DeleteProcess(table.v.Cid)
        if (!r) {
           t.Errorf("DeleteProcess failed to delete Cid %v (%v). Full struct %v", table.v.Cid, err, table.v)
        }
        fmt.Println("Deleted record with cid",table.v.Cid)
    }
    db.Close()
}

func TestDataTableLoad(t *testing.T) {
    tables := []struct {
        r bool
        v Context
    }{
        {true, Context{
                    SearchBy:          "[123] 'Specific' First Name -- for test purposes",
                    OrderByField:      "99",
                    OrderByDirection:  "invalid direction asc",
                    CurrentPage:       0,
                    Language:          "1",
        }},
        {true, Context{
                    SearchBy:          "(не-'латинские' символы) [144] Very Specific LAST Name -- for test purposes",
                    OrderByField:      "3",
                    OrderByDirection:  "desc",
                    CurrentPage:       0,
                    Language:          "3",
        }},
        {true, Context{
                    SearchBy:          "1992-10-18",
                    OrderByField:      "4",
                    OrderByDirection:  "asc",
                    CurrentPage:       0,
                    Language:          "1",
        }},
        {true, Context{
                    SearchBy:          "Female",
                    OrderByField:      "5",
                    OrderByDirection:  "desc",
                    CurrentPage:       0,
                    Language:          "2",
        }},
        {true, Context{
                    SearchBy:          "First_Name.Last_Name@longdummytestemailaddressforgolangunittests.go",
                    OrderByField:      "6",
                    OrderByDirection:  "asc",
                    CurrentPage:       0,
                    Language:          "1",
        }},
        {true, Context{
                    SearchBy:          "[-=+] 'Test' address in test city with digits 12974527 72628 98271",
                    OrderByField:      "7",
                    OrderByDirection:  "asc",
                    CurrentPage:       0,
                    Language:          "2",
        }},
        {false, Context{
                    SearchBy:          "no such combintion for tests in test sample",
                    OrderByField:      "1",
                    OrderByDirection:  "asc",
                    CurrentPage:       0,
                    Language:          "1",
        }},
    }
    c := Customer{
                    FirstName:         "[123] 'Specific' First Name -- for test purposes",
                    LastName:          "(не-'латинские' символы) [144] Very Specific LAST Name -- for test purposes",
                    BirthDate:         "1992-10-18",
                    Gender:            "Female",
                    Email:             "First_Name.Last_Name@longdummytestemailaddressforgolangunittests.go",
                    Address:           "[-=+] 'Test' address in test city with digits 12974527 72628 98271",
    }
    InitDb()
    //Preparations
    if (!InsertProcess(&c, "1") ) {
        t.Errorf("Preparation1: InsertProcess failed: %v", c)
    }
    fmt.Println("Preparation1: Inserted record with cid",c.Cid)
    //Testing
    for _, table := range tables {
        var dataTable DataTable
        DataTableLoad(&table.v, &dataTable)

        var found bool
        for _, c1 := range dataTable.Customers {
            if ( (c1.Cid == c.Cid)&&(c1.FirstName == c.FirstName)&&(c1.LastName == c.LastName)&&(c1.BirthDate == c.BirthDate)&&(c1.Gender == c.Gender)&&(c1.Email == c.Email)&&(c1.Address == c.Address) ) {
                found = true
            }
        }
        if (found != table.r) {
            t.Errorf("Test failed as (%v!=%v): %v (%v)found within %v", found, table.r, c, found, dataTable)
        }

    }
    //Clean-Up
    r, err := DeleteProcess(c.Cid)
    if (!r) {
        t.Errorf("DeleteProcess failed to delete Cid %v (%v). Full struct %v", c.Cid, err, c)
    }
    fmt.Println("Deleted record with cid",c.Cid)
    db.Close()
}
