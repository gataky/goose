
Note
====

I only use the for development.  This should not be ran on remote databases.


Install
======

`go get -v github.com/sir-wiggles/goose`

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

Goose keeps track of the migrations ran in a table called goosey in your database. To initialize this table you'll need to run `goose init` which will create a table that keeps track of the migrations.

You can also initialize goose with a starting commit hash `goose init a1b2c3...`.

Config
======

Put a `.goose.yaml` file in your home director with the folling content. The first three are fields are required. `templates` is still a work in progress.

```
database-url: <postgres-url>

migration-repository: <path-to-schema-directory>
migration-directory: <migrations-directory>

templates:
  schema:
    up:
      BEGIN;

      -- insert your code here --

      INSERT INTO schema_version (version, logged_by) VALUES ('{{.Migration}}', '{{.Author}}');

      COMMIT;

    down:
      BEGIN;

      -- insert your code here --

      DELETE FROM schema_version where version = '{{.Migration}}';

      COMMIT;


  data:
    up: |
      -- insert your code here --

      INSERT INTO schema_version (version, logged_by) VALUES ('{{.Migration}}', '{{.Author}}');

    down: |
      -- insert your code here --

      DELETE FROM schema_version where version = '{{.Migration}}';

```

Warnings
========

Only ever apply migrations on the same branch as previous runs.  Running on different branches could lead to commit hashes not being found in cases of squashing.  If you really want to run it on a different branch.  Make sure you rollback your changes on the test branch and reapply them on the main branch.


