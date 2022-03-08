package database

type BlockDB struct {
	db DatabaseFile
}

func (b *BlockDB) Open() (err error) {
	err = b.db.Open(BlocksFileName)
	return err
}

func (b *BlockDB) Add(block []byte) (err error) {

	if !b.db.IsOpen() {
		return ErrClosed
	}

	err = b.db.Write(block)

	return err
}

func (b *BlockDB) Close() {
	b.db.Close()
}

func (b *BlockDB) Last() (result []byte, err error) {
	node, err := b.db.getLastNode()
	if err != nil {
		return nil, err
	}

	result = node.Data
	return result, err
}

func (b *BlockDB) LoadData(loaderFunc func([]byte)) {

	b.db.ForEach(loaderFunc)

}
