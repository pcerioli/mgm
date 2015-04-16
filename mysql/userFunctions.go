package mysql

import (
  "fmt"
  "database/sql"
  _ "github.com/go-sql-driver/mysql"
  "crypto/md5"
  "encoding/hex"
)

func (db Database) AddPendingUser(name string, email string, template string, password string, summary string) error {
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return err}
  defer con.Close()

  hasher := md5.New()
  hasher.Write([]byte(password))
  creds := hex.EncodeToString(hasher.Sum(nil))

  _, err = con.Exec("INSERT INTO users (name, email, gender, password, summary) VALUES(?, ?, ?, ?, ?)",
                   name, email, template, creds, summary)
  if err != nil {
    return err
  }
  return nil
}

func (db Database) IsEmailUnique(email string) (bool, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return false, err}
  defer con.Close()

  row := con.QueryRow("SELECT email FROM users WHERE email=?", email)
  var test string
  err = row.Scan(&test)
  if err != nil{
    if err == sql.ErrNoRows {
      return true, nil
    }
    return false, err
  }
  return false, nil
}

func (db Database) IsNameUnique(name string) (bool, error){
  con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
  if err != nil {return false, err}
  defer con.Close()

  row := con.QueryRow("SELECT name FROM users WHERE name=?", name)
  var test string
  err = row.Scan(&test)
  if err != nil{
    if err == sql.ErrNoRows {
      return true, nil
    }
    return false, err
  }
  return false, nil
}