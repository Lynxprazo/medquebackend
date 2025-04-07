package booking

// here  one device will be used to make booking to day for adult  for child  it will be almost 3 child
import (
	"database/sql"
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
	BookingRequest struct {
		Username   string    `json:"username" validate:"required"`
		Time       time.Time `json:"time" validate:"required"`
		Department string    `json:"department" validate:"required"`
		Day        string    `json:"day" validate:"required"`
		Diseases   string    `json:"disease" validate:"required"`
		Doctor     string    `json:"doctor" validate:"required"`
		Secretkey  string    `json:"secretekey" validate:"required"`
		Section    string    `json:"section" validate:"required"`
		DeviceId   string    `json:"deviceId"`
		Age        string    `json:"age"`
	}
)

func Booking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Respond{
			Message: "Method isnt Allowed",
			Success: false,
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	tx, errTx := handlerconn.Db.Begin()

	if errTx != nil {
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong Transaction Failed",
			Success: false,
		})
		return
	}
	defer tx.Rollback()
	var BR BookingRequest
	err := json.NewDecoder(r.Body).Decode(&BR)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Respond{
			Message: "Something went wrong here",
			Success: false,
		})
		return
	}
	if BR.Section == "Guest" {
		err = HandleGeust(tx, BR.Username, BR.Secretkey, BR.Time, BR.Department, BR.Day, BR.Diseases)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Message: "Bad request",
				Success: false,
			})
			return
		}
	}
	if BR.Section == "Shared" {
		err = HandleshareDevice(tx, BR.Username, BR.Time, BR.Secretkey, BR.DeviceId, BR.Department, BR.Day, BR.Diseases)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Message: "Invalid request",
				Success: false,
			})
			return
		}
	}
	if BR.Section == "Specialgroup"{
		err = Handlespecialgroup(tx, BR.Time, BR.Department, BR.Username,BR.Secretkey,BR.Day,BR.Diseases)
		if err !=nil{
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Respond{
				Message: "Failed to make booking",
				Success: false,
			})
			return
		}
	}

	if err = tx.Commit(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
			Message: "Transaction failed to commit",
			Success: false,
		})
		return
	}
	json.NewEncoder(w).Encode(Respond{
		Message: "Successuly made booking",
		Success: true,
	})
	

}
func HandleGeust(tx *sql.Tx, username string, secretkey string, Time time.Time, department string, day string, diseases string) error {
	var hashedsecretekey string
	err := tx.QueryRow("SELECT secretkey FROM Users  WHERE username = $1", username).Scan(&hashedsecretekey)
	if err != nil {
		return fmt.Errorf("something went wrong here:%w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedsecretekey), []byte(secretkey))
	if err != nil {
		return fmt.Errorf("wrong secretkey or username %w", err)
	}
	err = CheckbookingRequest(tx, Time, username)
	if err != nil {
		return fmt.Errorf("something went wrong here ")
	}
	query := "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, diseases, secretkey)
	if err != nil {
		return fmt.Errorf("something went wrong here: %w", err)
	}
	return nil

}
func HandleshareDevice(tx *sql.Tx, username string, Time time.Time, secretekey string, deviceId string, department string, day string, diseases string) error {
	var hashedsecretekey string
	err := tx.QueryRow("SELECT secretekey FROM Users WHERE username = $1", username).Scan(&hashedsecretekey)
	if err != nil {
		return fmt.Errorf("something went wrong here %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedsecretekey), []byte(secretekey))
	if err != nil {
		return fmt.Errorf("something went wrong failed to proccess secretkey")
	}
	//check if is not same user try to make request for him self
	var matching bool
	query := "SELECT EXISTS(SELECT 1 FROM Users WHERE username = $1 AND secretekey = $2 AND deviceId = $3)"
	err = tx.QueryRow(query, username, hashedsecretekey, deviceId).Scan(&matching)

	if err != nil {
		return fmt.Errorf("failed to execute query")
	}
	if matching {
		return fmt.Errorf("bad request:you can not make your own booking using by  site is only for others: %w", err)
	}
	err = CheckbookingRequest(tx, Time, username)
	if err != nil {
		return fmt.Errorf("something went wrong failed to validate user: %w", err)
	}
	query = "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, diseases, secretekey)
	if err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}
	return nil
}

// This checks whether the personnel has already booked more than once—either for the same service or different ones. It also ensures thata personnel cannot make more than two bookings before completing their required medical test.
func CheckbookingRequest(tx *sql.Tx, newtime time.Time, username string) error {
	// check if person try to make more than two booking before complete one of each
	var Countbooking int
	err := tx.QueryRow("SELECT COUNT(*) FROM bookingList WHERE username = $1 AND status = 'processing'", username).Scan(&Countbooking)
	if err != nil {
		return fmt.Errorf("something went wrong failed to fetch data: %w", err)
	}
	if Countbooking >= 2 {
		return fmt.Errorf("booking limit reached: complete at least one medical test before booking again")
	}
	// check if personal try to make more than one booking within are day
	WindowStart := newtime.Add(-24 * time.Hour)
	WindowEnd := newtime.Add(24 * time.Hour)
	var windowCount int

	query := "SELECT COUNT(*) FROM bookingList WHERE username = $1 AND time BETWEEN $2 AND $3"
	err = tx.QueryRow(query, username, WindowStart, WindowEnd).Scan(&windowCount)
	if err != nil {
		return fmt.Errorf("something went wrong %w", err)
	}
	if windowCount > 0 {
		return fmt.Errorf("booking not allowed: you can only make one booking within a 24-hour period")
	}
	return nil
}
func Handlespecialgroup(tx *sql.Tx,  Time time.Time, department string, username string,secretekey string,day string,diseases string) error {
	 var existuser bool
     err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE username = $1 AND secretekey = $2)",username,secretekey).Scan(&existuser)
	 if err != nil{
		return fmt.Errorf("something went wrong here: %w",err)
	 }
	 if existuser{
		return fmt.Errorf("patient already exist in system")
	 }
	 query := "INSERT INTO bookingList (username,time,department,day,disease,secretekey) VALUES($1,$2,$3,$4,$5,$6)"
	_, err = tx.Exec(query, username, Time, department, day, diseases, secretekey)
	if err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}
	return nil
}
// htpp request for the Historical booking list

 func BookingHistory(w http.ResponseWriter, r *http.Request){
	  if r.Method != http.MethodPost{
          w.WriteHeader(http.StatusMethodNotAllowed)
		  json.NewEncoder(w).Encode(Respond{
			Message: "Bad request",
			Success: false,
		  })
		  return
	  }
	  tx,errTx := handlerconn.Db.Begin()
	   if errTx != nil{
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Respond{
		  Message: "Something went wrong Transaction failed begin ",
		  Success: false,
		})
		return
	   }

 }

//  create function  to return all history available for this  patient 

func ReturnAll(tx *sql.Tx,username string,){

}


