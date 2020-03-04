set LANG=en
set PATH=%PATH%;C:\tools\pgsql\bin

mkdir build
mkdir build\postgres

pg_ctl -D build\postgres status
