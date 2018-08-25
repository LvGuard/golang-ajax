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
        v1 Context
        v2 Customer
    }{
        {true, Context{
                    SearchBy:          "First_Name.Last_Name@longdummytestemailaddressforgolangunittests.go",
                    OrderByField:      "1",
                    OrderByDirection:  "asc",
                    CurrentPage:       0,
                    Language:          "1",
        },     Customer{
                    FirstName:         "First Name",
                    LastName:          "Last_Name",
                    BirthDate:         "1995-09-01",
                    Gender:            "Male",
                    Email:             "First_Name.Last_Name@longdummytestemailaddressforgolangunittests.go",
                    Address:           "",
        }},
    }
    InitDb()
    for _, table := range tables {
        if (!InsertProcess(&table.v2, table.v1.Language) ) {
            t.Errorf("Preparation1: InsertProcess failed: %v", table.v2)
        }
        fmt.Println("Preparation1: Inserted record with cid",table.v2.Cid)
        
        var dataTable DataTable
        DataTableLoad(&table.v1, &dataTable)

        var found bool
        for _, c := range dataTable.Customers {
            if ( (c.Email == table.v2.Email) && (c.Cid == table.v2.Cid)) {
                found = true
            }
        }
        if (!found) {
            t.Errorf("Test failed as %v not found within %v", table.v2, dataTable)
        }

        r, err := DeleteProcess(table.v2.Cid)
        if (!r) {
            t.Errorf("DeleteProcess failed to delete Cid %v (%v). Full struct %v", table.v2.Cid, err, table.v2)
        }
        fmt.Println("Deleted record with cid",table.v2.Cid)
    }
    db.Close()
}
