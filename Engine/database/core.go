package database

import (
	"bytes"
	"encoding/binary"
	"engine/utils"
	"errors"
	"io"
	"os"
	"path"
	"sync"
)

var (
	ErrEof      = errors.New("eof")
	ErrBof      = errors.New("bof")
	ErrEmpty    = errors.New("empty")
	ErrNotFound = errors.New("notfound")
	ErrClosed   = errors.New("closed")
)

const (
	BOF = -1
	EOF = -2

	CurrentDatabaseVersion = 1
)

type (
	NodeData []byte

	DatabaseHeaderInfos struct {
		Version           uint8
		FirstNodePosition int64
		LastNodePosition  int64
		NodesCount        int64
		TotalLength       int64
		Hash              [32]byte
	}

	HeaderNodeStruct struct {
		Position   int64
		Previous   int64
		Next       int64
		Deleted    bool
		DataLength int32
	}

	DBNode struct {
		Header HeaderNodeStruct
		Data   NodeData
	}

	DatabaseFile struct {
		fileName    string
		db          *os.File
		headerInfo  DatabaseHeaderInfos
		rootNode    *DBNode
		currentNode *DBNode
		mutex       *sync.Mutex
	}
)

/* IsOpen() return true if the database file is already open */
func (d *DatabaseFile) IsOpen() bool {

	if d.db == nil {
		return false
	}

	_, err := d.db.Stat()
	return err == nil
}

/* New() create new instance of DatabaseFile, but does not open it, and return it to the caller */
func (d *DatabaseFile) New(filename string) (result *DatabaseFile) {

	result = &DatabaseFile{
		fileName: filename,
		headerInfo: DatabaseHeaderInfos{
			Version: CurrentDatabaseVersion,
		},
	}

	return result
}

/* Open() open or create the database file if it does not exists and read the first record */
func (d *DatabaseFile) Open(datafileName string) error {
	d.fileName = datafileName

	if len(d.fileName) == 0 {
		return errors.New("database name is empty")
	}

	fullPath := path.Join(DatabasePath, d.fileName)

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return err
	}

	d.db = f

	fileStats, err := f.Stat()
	if err != nil {
		d.Close()
		return err
	}

	if fileStats.Size() == 0 {
		return d.createHeader()
	} else {
		d.readHeader()
	}

	return d.readFirstNode()
}

/* Close() closes the database file */
func (d *DatabaseFile) Close() error {
	if !d.IsOpen() {
		return ErrClosed
	}
	return d.db.Close()
}

/* getNode() move the pointer to position and read the node on that position */
func (d *DatabaseFile) getNode(position int64) (*DBNode, error) {
	if position == BOF {
		return nil, ErrBof
	}

	if position == EOF {
		return nil, ErrEof
	}

	_, err := d.db.Seek(position, io.SeekStart)
	if err != nil {
		return nil, err
	}

	node := new(DBNode)

	err = binary.Read(d.db, binary.LittleEndian, &node.Header)
	if err != nil {
		return nil, err
	}

	encData := make([]byte, node.Header.DataLength)
	err = binary.Read(d.db, binary.LittleEndian, encData)
	if err != nil {
		return nil, err
	}

	decData, err := utils.AESGCMDecrypt(encData)
	if err != nil {
		return nil, err
	}

	node.Data = make(NodeData, len(decData))
	copy(node.Data, decData)

	return node, err
}

func (d *DatabaseFile) updateNode(node *DBNode) error {
	existingNode, err := d.getNode(node.Header.Position)
	if err != nil {
		return err
	}

	if existingNode.Header.DataLength < node.Header.DataLength {
		return errors.New("new node has the data size greater than existing one")
	}

	_, err = d.db.Seek(node.Header.Position, io.SeekStart)
	if err != nil {
		return err
	}

	err = binary.Write(d.db, binary.LittleEndian, &d.currentNode.Header)
	if err != nil {
		return err
	}

	return nil
}

func (d *DatabaseFile) Write(data []byte) error {
	if d.mutex == nil {
		d.mutex = &sync.Mutex{}
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	lastNode, err := d.getLastNode()
	if err != nil && !errors.Is(err, ErrEmpty) {
		return err
	}

	newPosition, err := d.db.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	encData, err := utils.AESGCMEncrypt(data)
	if err != nil {
		return err
	}

	if lastNode == nil {
		lastNode = &DBNode{
			Header: HeaderNodeStruct{
				Position:   BOF,
				Previous:   BOF,
				Next:       EOF,
				Deleted:    false,
				DataLength: int32(len(encData)),
			},
			Data: encData,
		}
	} else {
		lastNode.Header.Next = newPosition
		d.updateNode(lastNode)
	}

	newPosition, _ = d.db.Seek(0, io.SeekEnd)
	d.headerInfo.LastNodePosition = newPosition

	d.currentNode = &DBNode{
		Header: HeaderNodeStruct{
			Position:   newPosition,
			Previous:   lastNode.Header.Position,
			Next:       EOF,
			Deleted:    false,
			DataLength: int32(len(encData)),
		},
		Data: encData,
	}

	err = binary.Write(d.db, binary.LittleEndian, d.currentNode.Header)
	if err != nil {
		return err
	}

	err = binary.Write(d.db, binary.LittleEndian, d.currentNode.Data)
	if err != nil {
		return err
	}

	d.headerInfo.NodesCount++
	d.headerInfo.TotalLength += int64(len(encData))

	return d.writeHeader()
}

func (d *DatabaseFile) WriteCurrent(data []byte) (err error) {
	node, err := d.getNode(d.currentNode.Header.Position)
	if err != nil {
		return err
	}

	encData, err := utils.AESGCMEncrypt(data)
	if err != nil {
		return err
	}

	dataPos := node.Header.Position + int64(binary.Size(HeaderNodeStruct{}))

	_, err = d.db.WriteAt(encData, dataPos)

	return err
}

func (d *DatabaseFile) Count() int64 {
	return d.headerInfo.NodesCount
}

func (d *DatabaseFile) DataLength() int64 {
	return d.headerInfo.TotalLength
}

func (d *DatabaseFile) readFirstNode() error {
	if d.Count() == 0 {
		return ErrEmpty
	}

	node, err := d.getNode(d.headerInfo.FirstNodePosition)

	d.rootNode = node
	d.currentNode = node

	return err
}

func (d *DatabaseFile) getLastNode() (*DBNode, error) {

	if d.Count() == 0 {
		return nil, ErrEmpty
	}

	_, err := d.db.Seek(d.headerInfo.LastNodePosition, io.SeekStart)
	if err != nil {
		return nil, err
	}

	var node = new(DBNode)
	err = binary.Read(d.db, binary.LittleEndian, &node.Header)
	if err != nil {
		return nil, err
	}

	node.Data = make(NodeData, node.Header.DataLength)
	d.currentNode = node

	return node, binary.Read(d.db, binary.LittleEndian, node.Data)
}

func (d *DatabaseFile) Read() ([]byte, error) {
	if d.rootNode == nil || d.currentNode == nil {
		d.readFirstNode()
	}

	node, err := d.getNode(d.currentNode.Header.Position)
	if errors.Is(err, ErrEof) {
		return nil, err
	}

	nextNode, err2 := d.getNode(node.Header.Next)
	if errors.Is(err2, ErrEof) {
		return node.Data, err2
	}

	if nextNode != nil {
		d.currentNode = nextNode
	}

	return node.Data, err
}

func (d *DatabaseFile) createHeader() error {
	_, err := d.db.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = binary.Write(d.db, binary.LittleEndian, d.headerInfo)
	if err != nil {
		return err
	}

	firstNodePosition, err := d.db.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	d.headerInfo.FirstNodePosition = firstNodePosition
	d.headerInfo.LastNodePosition = firstNodePosition
	d.headerInfo.NodesCount = 0

	return d.writeHeader()
}

func (d *DatabaseFile) readHeader() error {
	_, err := d.db.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	return binary.Read(d.db, binary.LittleEndian, &d.headerInfo)
}

func (d *DatabaseFile) writeHeader() error {
	_, err := d.db.Seek(0, io.SeekStart)

	if err != nil {
		return err
	}

	return binary.Write(d.db, binary.LittleEndian, d.headerInfo)
}

func (d *DatabaseFile) First() error {
	return d.readFirstNode()
}

func (d *DatabaseFile) Exists(data []byte) bool {

	err := d.First()
	if err != nil {
		return false
	}

	for {
		dataFromDB, err := d.Read()

		if bytes.Equal(data, dataFromDB) {
			return true
		}

		if errors.Is(err, ErrEof) {
			break
		}

	}
	return false
}

func (db *DatabaseFile) FindData(compareDataFunction func(data []byte) []byte) (result []byte, err error) {
	if compareDataFunction == nil {
		return nil, errors.New("'compareDataFunction' is nil")
	}

	err = db.First()
	if errors.Is(err, ErrEmpty) {
		return nil, err
	}

	for {
		data, err := db.Read()

		found := compareDataFunction(data)

		if found != nil || errors.Is(err, ErrEof) {
			result = found
			break
		}
	}

	return result, err
}

func (db *DatabaseFile) ForEach(forEachCallback func([]byte)) (err error) {
	if !db.IsOpen() {
		return ErrClosed
	}

	err = db.First()
	if err != nil {
		return err
	}

	for {
		data, err := db.Read()
		forEachCallback(data)

		if errors.Is(err, ErrEof) {
			break
		}

	}

	return err
}

func (d *DatabaseFile) Update(data []byte, whereFunc func([]byte) bool) (err error) {
	if !d.IsOpen() {
		return ErrClosed
	}

	err = d.First()
	if err != nil {
		return err
	}

	for {
		existingData, err := d.Read()
		found := whereFunc(existingData)
		if found {
			d.WriteCurrent(data)
			return nil
		}

		if errors.Is(err, ErrEof) {
			break
		}
	}

	return ErrNotFound
}

func init() {
	err := os.MkdirAll(DatabasePath, os.ModeDir)
	if err != nil {
		panic(err)
	}
}
