CREATE TABLE LngData
(
  Lid         character varying(100) NOT NULL
, Language_1  character varying(500) NULL
, Language_2  character varying(500) NULL
, Language_3  character varying(500) NULL
, CONSTRAINT LngDataPKEY PRIMARY KEY (Lid)
)
WITH (OIDS=FALSE)
;
