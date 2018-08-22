CREATE TABLE Customers
(
  Cid serial NOT NULL
, FirstName  character varying(100) NOT NULL
, LastName   character varying(500) NOT NULL
, BirthDate  date NOT NULL
, Gender     character varying(6) NOT NULL
, Email      character varying(100) NOT NULL
, Address    character varying(200)
, LastUpdate timestamp NOT NULL
, CONSTRAINT CustomersPKEY PRIMARY KEY (Cid)
)
WITH (OIDS=FALSE)
;