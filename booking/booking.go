package booking

// here  one device will be used to make booking to day for adult  for child  it will be almost 3 child
import (
	"encoding/json"
	"fmt"
	handlerconn "medquemod/db_conn"
	"net/http"
	"time"
)

type (
	Respond struct {
		Message string      `json:"message"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}
	BookingRequest struct {
		Time       time.Time `json:"time" validate:"required"`
		Department string `json:"department" validate:"required"`
		Day        string `json:"day" validate:"required"`
		Diseases   string `json:"desease" validate:"required"`
		Doctor     string `json:"doctor" validate:"required"`
	}
	Secretkey struct{
		Secretkey string `json:"deviceId"`
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

}

func HandleGeust(username string, secretkey string, time int, department string, day string, diseases string)error {
	tx, errTx := handlerconn.Db.Begin()
	defer tx.Rollback()
	if errTx != nil {
		return fmt.Errorf("something went wrong here Transaction failed: %w",errTx)
	}
	var idExist bool
    err:= tx.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE secretkey = $1)",secretkey).Scan(&idExist)

	if err !=nil{
		return fmt.Errorf("something went wrong: %w",err)
	}
	if !idExist {
		return fmt.Errorf("user doesnt exist on system")
	}
    err = tx.QueryRow("SELECT  EXISTS(SELECT 1 FROM Users WHERE secretkey = $1 AND username = $2)",secretkey,username).Scan(&idExist)

	if err != nil{
		return fmt.Errorf("something went wrong here: %w",err)
	}
	if !idExist {
		return fmt.Errorf("wrong username or secretkey")
	}
	
	
	query := "INSERT INTO bookingList (username,time,department,day,disease) VALUES($1,$2,$3,$4,$5,$6)"
	_,err = tx.Exec(query,username,time,department,day,diseases,secretkey)
	if err !=nil{
		return fmt.Errorf("something went wrong here: %w",err)
	}
	if err = tx.Commit();err !=nil{
		return fmt.Errorf("something went wrong here: %w",err)
	}
	return nil

}
func handlechild(){

}
func handleshareDevice() {

}
// This checks whether the personnel has already booked more than once—either for the same service or different ones. It also ensures thata personnel cannot make more than two bookings before completing their required medical test.
 
func CheckbookingRequest(newtime time.Time, secretkey string,username string)error{
	tx,errTx := handlerconn.Db.Begin()
	defer tx.Rollback()
	if errTx !=nil{
		return fmt.Errorf("something went wrong transaction failed:%w",errTx)
	}
	// check if person try to make more than two booking before complete one of each
	var Countbooking int
	err := tx.QueryRow("SELECT username COUNT(*) WHERE username = $1",username).Scan(&Countbooking)
	// check if personal try to make more than one booking within are day
	WindowStart := newtime.Add(-24 * time.Hour)
	WindowEnd := newtime.Add(24 * time.Hour)




   
    

}
