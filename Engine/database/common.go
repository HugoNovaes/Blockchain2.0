package database

const (
	DatabasePath         = "./db"
	BlocksFileName       = "blocks.dat"
	TransactionsFileName = "transactions.dat"
	AccountsFileName     = "accounts.dat"
)

type (
	FindDataCallback = func([]byte) []byte

	IDataTable interface {
		Open() (IDataTable, error)
		Close()
		Find(FindDataCallback) ([]byte, error)
		Save() error
		Delete() error
		First() error
		Next() error
		Last() error
	}

	DataTable struct {
		dataFile *DatabaseFile
	}
)

func (d DataTable) Open(datafileName string) (err error) {
	d.dataFile = &DatabaseFile{}
	err = d.dataFile.Open(datafileName)
	return err
}

func (d DataTable) Close() {
	d.dataFile.Close()
}

func (d DataTable) Find(FindDataCallback) (result []byte) {
	return result
}

func (d DataTable) Save() (err error) {
	return err
}

func (d DataTable) Delete() (err error) {
	return err
}

func (d DataTable) First() (err error) {
	return err
}

func (d DataTable) Next() (err error) {
	return err
}

func (d DataTable) Last() (err error) {
	return err
}
