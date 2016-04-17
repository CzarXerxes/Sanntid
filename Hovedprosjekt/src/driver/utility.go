package driver


//Global variables which are used by multiple modules
var BackupOrderFilePath = "/home/student/Desktop/Heis/backupOrders.gob"
var IP = "129.241.187.153"
var Port = "30000"

func BoolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func IntToBool(i int) bool {
	if i == 1 {
		return true
	} else {
		return false
	}
}

func Save(path string, object interface{}) error {
	file, err := os.Create(path)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func Load(path string, object interface{}) error {
	file, err := os.Open(path)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}

func Check(e error) bool {
	if e != nil {
		return false
	}
	return true
}
