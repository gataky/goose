About
=====

Goose is a SQL migration tool that applies migrations relative to the last migration file ran.  You can apply N number of migrations, up or down, to your database and goose will keep track of everything for you. What makes goose different is that is uses git to determin the order of migrations to execute. To determine this order goose runs `git log --pretty='format:%Cred%H|%aD' --name-status --diff-filter=A --reverse` which produces

```
e965f4511fce6ae61e1cfdcf174f61cfd4fe920b|Wed, 11 Nov 2020 22:53:49 -0800
A       20201023_010000_a_o_solv-20201023_010000_a/down.sql
A       20201023_010000_a_o_solv-20201023_010000_a/up.sql

cac4966fa648df678b9f59117d085b40d647ef19|Wed, 11 Nov 2020 22:54:06 -0800
A       20201023_020000_b_o_solv-20201023_020000_b/down.sql
A       20201023_020000_b_o_solv-20201023_020000_b/up.sql

e0ca0a9d0afe2d168ed09efe2f859f76bcfd109f|Wed, 11 Nov 2020 22:54:20 -0800
A       20201023_030000_c_o_solv-20201023_030000_c/down.sql
A       20201023_030000_c_o_solv-20201023_030000_c/up.sql

6a8f40ecd57b264da0d0492af62b577f626bfbe1|Wed, 11 Nov 2020 22:54:56 -0800
A       20201023_040000_d_o_solv-20201023_040000_d/down.sql
A       20201023_040000_d_o_solv-20201023_040000_d/up.sql

76499a490b9c0006100d963e6006f72cf56c6826|Wed, 11 Nov 2020 22:55:07 -0800
A       20201023_050000_e_o_solv-20201023_050000_e/down.sql
A       20201023_050000_e_o_solv-20201023_050000_e/up.sql

9ebb39681a4428cc5693ea2d926e5f73711ce9a4|Wed, 11 Nov 2020 22:55:17 -0800
A       20201023_060000_f_o_solv-20201023_060000_f/down.sql
A       20201023_060000_f_o_solv-20201023_060000_f/up.sql

cc7eff6ea9e68da4265bc834afda28f9a9db05a8|Wed, 11 Nov 2020 22:55:31 -0800
A       20201023_070000_g_o_solv-20201023_070000_g/down.sql
A       20201023_070000_g_o_solv-20201023_070000_g/up.sql
```

Goose will parse this output and apply the migration in a top down approach. 

Terms
=====

* migration: a single script to run against the database
* batch    : a group of migrations to run against the database
* marker   : a marker of the migrations that were last in their batch
* hash     : the git hash of those added files

Initialization
==============

Goose keeps track of the migrations ran in a table called goosey in your database. To initialize this table you'll need to run `goose init` which will create a table that has six important fields:
* created_at : The time a migration was made. This information is encoded in the directory name
* merged_at  : The time a migration was committed with the repository
* executed_at: The time a migration was ran against the database
* hash       : The commit hash of when these files were added to the repository
* author     : The author of a migration. This information is encoded in the directory name
* batch      : The batch that this migration relates to

Making a new migration
======================

To make a migration run "goose make (firstname) (lastname) (message)" this will create directory with the format yyyymmdd_hhmmss_firstname_lastname_message and in that directory two files, up.sql and down.sql, will be created where you'll write your SQL scripts. Other files added do these directories will be ignored by goose.

The structure of your migration directory will look like 
```
.
├── 20201023_010000_a_o_solv-20201023_010000_a
│   ├── down.sql
│   └── up.sql
├── 20201023_020000_b_o_solv-20201023_020000_b
│   ├── down.sql
│   └── up.sql
├── 20201023_030000_c_o_solv-20201023_030000_c
│   ├── down.sql
│   └── up.sql
├── 20201023_040000_d_o_solv-20201023_040000_d
│   ├── down.sql
│   └── up.sql
├── 20201023_050000_e_o_solv-20201023_050000_e
│   ├── down.sql
│   └── up.sql
├── 20201023_060000_f_o_solv-20201023_060000_f
│   ├── down.sql
│   └── up.sql
└── 20201023_070000_g_o_solv-20201023_070000_g
    ├── down.sql
    └── up.sql
```
