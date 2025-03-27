#!/bin/bash
set -ex

psql -U $POSTGRES_USER -d $POSTGRES_DB > /var/log/postgresql/setup.log <<EOF
CREATE DATABASE test;
\c test
CREATE SCHEMA yezzey;
INSERT INTO pg_extension (oid, extowner, extnamespace, extrelocatable, extname, extversion) VALUES (8555, 10, 11, false, 'yezzey', '1.8.4');
CREATE TABLE yezzey.yezzey_expire_hint (x_path text, lsn pg_lsn);
CREATE TABLE yezzey.yezzey_virtual_index(x_path text);
INSERT INTO yezzey.yezzey_virtual_index VALUES ('/encrypted_file_old_key');